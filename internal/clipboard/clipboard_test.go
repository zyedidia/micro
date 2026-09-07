package clipboard

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/micro-editor/micro/v2/internal/screen"
	"github.com/micro-editor/tcell/v2"
	"github.com/zyedidia/clipper"
)

// controlledClipboard keeps tests independent of installed clipboard tools and
// lets them hold detection open without launching a real clipboard command.
type controlledClipboard struct {
	started chan struct{}
	release chan struct{}
	err     error
	text    map[string][]byte
}

func (c *controlledClipboard) Init() error {
	c.started <- struct{}{}
	<-c.release
	return c.err
}

func (c *controlledClipboard) ReadAll(reg string) ([]byte, error) {
	return c.text[reg], nil
}

func (c *controlledClipboard) WriteAll(reg string, text []byte) error {
	c.text[reg] = text
	return nil
}

func setupClipboard(t *testing.T) *controlledClipboard {
	t.Helper()
	t.Setenv("PATH", t.TempDir())
	savedClips, savedExternal, savedMethod := clipper.Clipboards, external, CurrentMethod
	savedInternal, savedMulti := internal, multi
	c := &controlledClipboard{
		started: make(chan struct{}, 10),
		release: make(chan struct{}),
		text:    make(map[string][]byte),
	}
	clipper.Clipboards = []clipper.Clipboard{c}
	external = nil
	internal = make(internalClipboard)
	multi = make(multiClipboard)
	t.Cleanup(func() {
		c.finish()
		if external != nil {
			waitFor(t, external.done, "clipboard cleanup")
		}
		clipper.Clipboards, external, CurrentMethod = savedClips, savedExternal, savedMethod
		internal, multi = savedInternal, savedMulti
	})
	return c
}

func (c *controlledClipboard) finish() {
	select {
	case <-c.release:
	default:
		close(c.release)
	}
}

func waitFor(t *testing.T, ch <-chan struct{}, operation string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", operation)
	}
}

func TestInitializeFailurePreservesSelectedMethod(t *testing.T) {
	c := setupClipboard(t)
	c.err = errors.New("clipboard unavailable")
	SetMethod("external")
	done := make(chan struct{})
	var initErr error
	go func() {
		initErr = Initialize(External)
		close(done)
	}()
	waitFor(t, c.started, "clipboard detection")
	SetMethod("terminal")
	if err := Initialize(Terminal); err != nil {
		t.Fatal(err)
	}
	close(c.release)
	waitFor(t, done, "failed clipboard detection")
	if initErr == nil {
		t.Fatal("external initialization should fail")
	}
	if CurrentMethod != Terminal {
		t.Fatalf("failed external detection changed selected method to %v; want Terminal", CurrentMethod)
	}
}

// A timeout is only a deadlock guard: the probe stays blocked until the test
// explicitly releases it, regardless of the speed of the machine running it.
func beforeDetectionFinishes(t *testing.T, c *controlledClipboard, operation func() error) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- operation() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		c.finish()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("operation did not finish after releasing detection")
		}
		t.Fatal("operation waited for external clipboard detection")
	}
}

func startBlockedDetection(t *testing.T, c *controlledClipboard) {
	t.Helper()
	SetMethod("external")
	beforeDetectionFinishes(t, c, func() error {
		InitAsync(External)
		return nil
	})
	waitFor(t, c.started, "clipboard detection")
}

func TestInitAsyncPreservesLaterMethod(t *testing.T) {
	for _, method := range []string{"terminal", "internal"} {
		t.Run(method, func(t *testing.T) {
			c := setupClipboard(t)
			c.err = errors.New("clipboard unavailable")
			startBlockedDetection(t, c)
			selected := SetMethod(method)
			if err := Initialize(selected); err != nil {
				t.Fatal(err)
			}
			c.finish()
			waitFor(t, external.done, "failed detection")
			if CurrentMethod != selected {
				t.Fatalf("detection changed method to %v; want %v", CurrentMethod, selected)
			}
		})
	}
}

func checkRegister(r Register) error {
	if err := Write("clipboard text", r); err != nil {
		return err
	}
	if text, err := Read(r); err != nil || text != "clipboard text" {
		return fmt.Errorf("Read(%d) = %q, %v", r, text, err)
	}
	for i, text := range []string{"first", "second"} {
		if err := WriteMulti(text, r, i, 2); err != nil {
			return err
		}
	}
	for i, want := range []string{"first", "second"} {
		if text, err := ReadMulti(r, i, 2); err != nil || text != want {
			return fmt.Errorf("ReadMulti(%d, %d) = %q, %v; want %q", r, i, text, err, want)
		}
	}
	return nil
}

func TestLocalRegistersDoNotWaitForDetection(t *testing.T) {
	c := setupClipboard(t)
	SetMethod("internal")
	// Internal storage is also available before startup initializes anything.
	beforeDetectionFinishes(t, c, func() error { return checkRegister(ClipboardReg) })
	startBlockedDetection(t, c)
	for _, method := range []string{"external", "terminal", "internal"} {
		SetMethod(method)
		for _, r := range []Register{0, 7, -3} {
			beforeDetectionFinishes(t, c, func() error { return checkRegister(r) })
		}
	}
	for _, r := range []Register{ClipboardReg, PrimaryReg} {
		beforeDetectionFinishes(t, c, func() error { return checkRegister(r) })
	}
}

type clipboardScreen struct {
	tcell.Screen
	text map[string]string
}

func (s *clipboardScreen) SetClipboard(text, reg string) error {
	s.text[reg] = text
	return nil
}

func (s *clipboardScreen) GetClipboard(reg string) error {
	screen.Events <- tcell.NewEventPaste(s.text[reg], "")
	return nil
}

func TestTerminalClipboardDoesNotWaitForDetection(t *testing.T) {
	c := setupClipboard(t)
	startBlockedDetection(t, c)
	savedScreen, savedEvents := screen.Screen, screen.Events
	s := &clipboardScreen{text: make(map[string]string)}
	screen.Screen = s
	screen.Events = make(chan tcell.Event, 1)
	t.Cleanup(func() { screen.Screen, screen.Events = savedScreen, savedEvents })
	SetMethod("terminal")
	beforeDetectionFinishes(t, c, func() error {
		for _, r := range []Register{ClipboardReg, PrimaryReg} {
			if err := Write("terminal text", r); err != nil {
				return err
			}
		}
		s.text["clipboard"], s.text["primary"] = "terminal paste", "primary paste"
		for r, want := range map[Register]string{ClipboardReg: "terminal paste", PrimaryReg: "primary paste"} {
			if text, err := Read(r); err != nil || text != want {
				return fmt.Errorf("terminal Read(%d) = %q, %v; want %q", r, text, err, want)
			}
		}
		return nil
	})
	if s.text["c"] != "terminal text" || s.text["p"] != "terminal text" {
		t.Fatalf("terminal writes = %v", s.text)
	}
}

func TestExternalClipboardWaitsForDetection(t *testing.T) {
	readText := func(text string, err error, want string) error {
		if err != nil || text != want {
			return fmt.Errorf("read = %q, %v; want %q", text, err, want)
		}
		return nil
	}
	operations := []struct {
		name string
		run  func(Register) error
		want string
	}{
		{"read", func(r Register) error {
			text, err := Read(r)
			return readText(text, err, "firstsecond")
		}, "firstsecond"},
		{"write", func(r Register) error { return Write("written", r) }, "written"},
		{"read multi", func(r Register) error {
			text, err := ReadMulti(r, 0, 2)
			return readText(text, err, "first")
		}, "firstsecond"},
		{"write multi", func(r Register) error { return WriteMulti("updated", r, 0, 2) }, "updatedsecond"},
	}
	for _, r := range []Register{ClipboardReg, PrimaryReg} {
		for _, operation := range operations {
			t.Run(fmt.Sprintf("%d/%s", r, operation.name), func(t *testing.T) {
				c := setupClipboard(t)
				reg := clipper.RegClipboard
				if r == PrimaryReg {
					reg = clipper.RegPrimary
				}
				c.text[reg] = []byte("firstsecond")
				multi.writeText("first", r, 0, 2)
				multi.writeText("second", r, 1, 2)
				startBlockedDetection(t, c)
				// Initializing another method must not mark the external probe ready.
				if err := Initialize(SetMethod("terminal")); err != nil {
					t.Fatal(err)
				}
				SetMethod("external")
				done := make(chan error, 1)
				go func() { done <- operation.run(r) }()
				select {
				case err := <-done:
					t.Fatalf("external clipboard operation finished before detection: %v", err)
				case <-time.After(20 * time.Millisecond):
				}
				c.finish()
				select {
				case err := <-done:
					if err != nil {
						t.Fatal(err)
					}
				case <-time.After(5 * time.Second):
					t.Fatal("external clipboard operation did not finish")
				}
				if string(c.text[reg]) != operation.want {
					t.Fatalf("external register %s = %q", reg, c.text[reg])
				}
			})
		}
	}
}

func TestPendingDetectionIsReused(t *testing.T) {
	c := setupClipboard(t)
	startBlockedDetection(t, c)
	pending := external
	for _, refresh := range []bool{false, true} {
		if d := detect(refresh); d != pending {
			t.Fatal("a new probe replaced detection that was still in progress")
		}
	}
}

func TestFailedDetectionFallsBackAndCanBeRetried(t *testing.T) {
	c := setupClipboard(t)
	c.err = errors.New("clipboard unavailable")
	startBlockedDetection(t, c)
	c.finish()
	waitFor(t, external.done, "failed detection")
	for _, r := range []Register{ClipboardReg, PrimaryReg} {
		if err := checkRegister(r); err != nil {
			t.Fatal(err)
		}
	}
	// Explicit initialization can discover a backend installed after startup.
	c.err = nil
	if err := Initialize(External); err != nil {
		t.Fatal(err)
	}
	if err := checkRegister(ClipboardReg); err != nil {
		t.Fatal(err)
	}
	if string(c.text[clipper.RegClipboard]) != "firstsecond" {
		t.Fatal("retry did not use the external clipboard")
	}
}

func TestInitializeRefreshesCompletedDetection(t *testing.T) {
	c := setupClipboard(t)
	c.finish()
	SetMethod("external")
	if err := Initialize(External); err != nil {
		t.Fatal(err)
	}
	// Replace the available provider. Ordinary operations keep their backend,
	// but explicitly selecting external again must discover the replacement.
	replacement := &controlledClipboard{
		started: make(chan struct{}, 1), release: c.release,
		text: map[string][]byte{clipper.RegClipboard: []byte("replacement")},
	}
	clipper.Clipboards = []clipper.Clipboard{replacement}
	if err := checkRegister(ClipboardReg); err != nil {
		t.Fatal(err)
	}
	if string(replacement.text[clipper.RegClipboard]) != "replacement" {
		t.Fatal("ordinary clipboard operations re-detected the backend")
	}
	if err := Initialize(External); err != nil {
		t.Fatal(err)
	}
	if text, err := Read(ClipboardReg); err != nil || text != "replacement" {
		t.Fatalf("refreshed clipboard = %q, %v", text, err)
	}
}
