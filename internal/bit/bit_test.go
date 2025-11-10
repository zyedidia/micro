package bit

import "testing"

func TestBand(t *testing.T) {
	got := Band(0xF0F0, 0x0FF0)
	want := uint64(0x0F0)
	if got != want {
		t.Errorf("Band failed: got %#x, want %#x", got, want)
	}
}

func TestBor(t *testing.T) {
	got := Bor(0xF0F0, 0x0FF0)
	want := uint64(0xFFF0)
	if got != want {
		t.Errorf("Bor failed: got %#x, want %#x", got, want)
	}
}

func TestBxor(t *testing.T) {
	got := Bxor(0xAAAA, 0x5555)
	want := uint64(0xFFFF)
	if got != want {
		t.Errorf("Bxor failed: got %#x, want %#x", got, want)
	}
}

func TestBnot(t *testing.T) {
	got := Bnot(0x0)
	want := ^uint64(0x0)
	if got != want {
		t.Errorf("Bnot failed: got %#x, want %#x", got, want)
	}
}

func TestLshift(t *testing.T) {
	got := Lshift(1, 8)
	want := uint64(256)
	if got != want {
		t.Errorf("Lshift failed: got %d, want %d", got, want)
	}

	if Lshift(1, 64) != 0 {
		t.Errorf("Lshift should zero out when n >= 64")
	}
}

func TestRshift(t *testing.T) {
	got := Rshift(0x100, 8)
	want := uint64(1)
	if got != want {
		t.Errorf("Rshift failed: got %d, want %d", got, want)
	}

	if Rshift(0x100, 64) != 0 {
		t.Errorf("Rshift should zero out when n >= 64")
	}
}

func TestArshift(t *testing.T) {
	// positive
	got := Arshift(1024, 2)
	want := int64(256)
	if got != want {
		t.Errorf("Arshift positive failed: got %d, want %d", got, want)
	}

	// negative, arithmetic shift should preserve sign bit
	got = Arshift(-1024, 2)
	want = -256
	if got != want {
		t.Errorf("Arshift negative failed: got %d, want %d", got, want)
	}

	// shift overflow
	if Arshift(-1, 64) != -1 {
		t.Errorf("Arshift(-1,64) should stay -1 (all bits 1)")
	}
}
