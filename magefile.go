//go:build mage
// +build mage

package main

import (
	"fmt"

	"github.com/bitfield/script"
	"github.com/magefile/mage/sh"
)

var (
	GOOS, _    = script.Exec("go env GOHOSTOS").String()
	GOARCH, _  = script.Exec("go env GOHOSTARCH").String()
	doublet    = map[string]string{"GOOS": GOOS, "GOARCH": GOARCH}
	VERSION, _ = sh.OutputWith(doublet, "go", "run", "tools/build-version.go", "2>", "/dev/null")

	HASH, _                    = script.Exec("git rev-parse --short HEAD").String()
	DATE, _                    = script.Exec("echo $(GOOS=$(go env GOHOSTOS) GOARCH=$(go env GOHOSTARCH) go run tools/build-date.go)").String()
	GOBIN, _                   = script.Exec("echo $(go env GOPATH)/bin").String()
	DEBUGVAR                   = "-X github.com/zyedidia/micro/v2/internal/util.Debug=ON"
	VSCODE_TESTS_BASE_URL      = "'https://raw.githubusercontent.com/microsoft/vscode/e6a45f4242ebddb7aa9a229f85555e8a3bd987e2/src/vs/editor/test/common/model/'"
	GOHOSTOS, _                = script.Exec("go env GOHOSTOS").String()
	GOHOSTARCH, _              = script.Exec("go env GOHOSTARCH").String()
	GOVARS                     = fmt.Sprintf("-X github.com/zyedidia/micro/v2/internal/util.Version=%s-X github.com/zyedidia/micro/v2/internal/util.CommitHash=%s -X 'github.com/zyedidia/micro/v2/internal/util.CompileDate=%s'", VERSION, HASH, DATE)
	CGO_ENABLED, _             = script.Exec("go env CGO_ENABLED").String()
	ADDITIONAL_GO_LINKER_FLAGS = ""
)

func init() {
	if GOHOSTOS == "Darwin" {
		//DARWIN_FLAGS, _ := script.Exec("echo $(GOOS=$(go env GOHOSTOS) GOARCH=$(go env GOHOSTARCH) go run tools/info-plist.go $(go env GOOS) $(GOOS=$(go env GOHOSTOS) GOARCH=$(go env GOHOSTARCH) go run tools/build-version.go)").String()
		//ADDITIONAL_GO_LINKER_FLAGS += DARWIN_FLAGS
		CGO_ENABLED = "1"
	}
}

var Default = Build

func Build() {
	Generate()
	Build_quick()
}

func Build_quick() error {
	env := map[string]string{"CGO_ENABLED": CGO_ENABLED}
	fmt.Println(doublet)
	flagld := fmt.Sprintf("-s -w %s %s", GOVARS, ADDITIONAL_GO_LINKER_FLAGS)
	if err := sh.RunWith(env, "go", "build", "-trimpath", "-ldflags "+`"`+flagld+`"`, "./cmd/micro"); err != nil {
		return err
	}
	return nil
}

func Build_dbg() error {
	env := map[string]string{"CGO_ENABLED": CGO_ENABLED}
	flagld := fmt.Sprintf(`"%s"`, "%s %s", ADDITIONAL_GO_LINKER_FLAGS, DEBUGVAR)
	if err := sh.RunWith(env, "go", "build", "-trimpath", "-ldflags", flagld, "./cmd/micro"); err != nil {
		return err
	}
	return nil
}

func Build_tags() {
	Fetch_tags()
	Build()
}

func Build_all() {
	Build()
}

func Install() error {
	Generate()
	flagld := fmt.Sprintf(`"%s"`, "-s -w %s %s", GOVARS, ADDITIONAL_GO_LINKER_FLAGS)
	if err := sh.Run("go", "install", "-ldflags", flagld, "./cmd/micro"); err != nil {
		return err
	}
	return nil
}

func Install_all() {
	Install()
}

func Fetch_tags() error {
	if err := sh.Run("git", "fetch", "--tags", "--force"); err != nil {
		return err
	}
	return nil
}

func Generate() error {
	env := map[string]string{"GOOS": GOHOSTOS, "GOARCH": GOHOSTARCH}
	if err := sh.RunWith(env, "go", "generate", "./runtime"); err != nil {
		return err
	}
	return nil
}

func Testgen() error {
	if err := sh.Run("mkdir", "-p", "tools/vscode-tests"); err != nil {
		return err
	}
	if err := sh.Run("cd", "tools/vscode-tests", "&&", "curl", "--remote-name-all", VSCODE_TESTS_BASE_URL+"{editableTextModelAuto,editableTextModel,model.line}.test.ts"); err != nil {
		return err
	}
	if err := sh.Run("tsc", "tools/vscode-tests/*.ts", ">", "/dev/null;", "true"); err != nil {
		return err
	}
	if err := sh.Run("go", "run", "tools/testgen.go", "tools/vscode-tests/*.js", ">", "buffer_generated_test.go"); err != nil {
		return err
	}
	if err := sh.Run("mv", "buffer_generated_test.go", "internal/buffer"); err != nil {
		return err
	}
	if err := sh.Run("gofmt", "-w", "internal/buffer/buffer_generated_test.go"); err != nil {
		return err
	}
	return nil
}

func Test() error {
	if err := sh.Run("go", "test", "./internal/..."); err != nil {
		return err
	}
	if err := sh.Run("go", "test", "./cmd/..."); err != nil {
		return err
	}
	return nil
}

func Bench() error {
	if err := sh.Run("for", "i", "in", "1", "2", "3;", "do", "go", "test", "-bench=.", "./internal/...;", "done", ">", "benchmark_results"); err != nil {
		return err
	}
	if err := sh.Run("benchstat", "benchmark_results"); err != nil {
		return err
	}
	return nil
}

func Bench_baseline() error {
	if err := sh.Run("for", "i", "in", "1", "2", "3;", "do", "go", "test", "-bench=.", "./internal/...;", "done", ">", "benchmark_results_baseline"); err != nil {
		return err
	}
	return nil
}

func Bench_compare() error {
	if err := sh.Run("for", "i", "in", "1", "2", "3;", "do", "go", "test", "-bench=.", "./internal/...;", "done", ">", "benchmark_results"); err != nil {
		return err
	}
	if err := sh.Run("benchstat", "-alpha", "0.15", "benchmark_results_baseline", "benchmark_results"); err != nil {
		return err
	}
	return nil
}

func Clean() error {
	if err := sh.Run("rm", "-f", "micro"); err != nil {
		return err
	}
	return nil
}
