.PHONY: runtime

VERSION = $(shell GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/build-version.go)
HASH = $(shell git rev-parse --short HEAD)
DATE = $(shell GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/build-date.go)
ADDITIONAL_GO_LINKER_FLAGS = $(shell GOOS=$(shell go env GOHOSTOS) \
	GOARCH=$(shell go env GOHOSTARCH))
GOBIN ?= $(shell go env GOPATH)/bin
GOVARS = -X github.com/zyedidia/micro/v2/internal/util.Version=$(VERSION) -X github.com/zyedidia/micro/v2/internal/util.CommitHash=$(HASH) -X 'github.com/zyedidia/micro/v2/internal/util.CompileDate=$(DATE)'
DEBUGVAR = -X github.com/zyedidia/micro/v2/internal/util.Debug=ON
VSCODE_TESTS_BASE_URL = 'https://raw.githubusercontent.com/microsoft/vscode/e6a45f4242ebddb7aa9a229f85555e8a3bd987e2/src/vs/editor/test/common/model/'

# Builds micro after checking dependencies but without updating the runtime
build:
	go build -ldflags "-s -w $(GOVARS) $(ADDITIONAL_GO_LINKER_FLAGS)" ./cmd/micro

build-dbg:
	go build -ldflags "-s -w $(ADDITIONAL_GO_LINKER_FLAGS) $(DEBUGVAR)" ./cmd/micro

build-tags: fetch-tags
	go build -ldflags "-s -w $(GOVARS) $(ADDITIONAL_GO_LINKER_FLAGS)" ./cmd/micro

# Builds micro after building the runtime and checking dependencies
build-all: runtime build

# Builds micro without checking for dependencies
build-quick:
	go build -ldflags "-s -w $(GOVARS) $(ADDITIONAL_GO_LINKER_FLAGS)" ./cmd/micro

# Same as 'build' but installs to $GOBIN afterward
install:
	go install -ldflags "-s -w $(GOVARS) $(ADDITIONAL_GO_LINKER_FLAGS)" ./cmd/micro

# Same as 'build-all' but installs to $GOBIN afterward
install-all: runtime install

# Same as 'build-quick' but installs to $GOBIN afterward
install-quick:
	go install -ldflags "-s -w $(GOVARS) $(ADDITIONAL_GO_LINKER_FLAGS)"  ./cmd/micro

fetch-tags:
	git fetch --tags

# Builds the runtime
runtime:
	git submodule update --init
	go run runtime/syntax/make_headers.go runtime/syntax
	go build -o tools/bindata ./tools/go-bindata
	tools/bindata -pkg config -nomemcopy -nometadata -o runtime.go runtime/...
	mv runtime.go internal/config
	gofmt -w internal/config/runtime.go

testgen:
	mkdir -p tools/vscode-tests
	cd tools/vscode-tests && \
	curl --remote-name-all $(VSCODE_TESTS_BASE_URL){editableTextModelAuto,editableTextModel,model.line}.test.ts
	tsc tools/vscode-tests/*.ts > /dev/null; true
	go run tools/testgen.go tools/vscode-tests/*.js > buffer_generated_test.go
	mv buffer_generated_test.go internal/buffer
	gofmt -w internal/buffer/buffer_generated_test.go

test:
	go test ./internal/...

bench:
	for i in 1 2 3; do \
		go test -bench=. ./internal/...; \
	done > benchmark_results
	benchstat benchmark_results

bench-baseline:
	for i in 1 2 3; do \
		go test -bench=. ./internal/...; \
	done > benchmark_results_baseline

bench-compare:
	for i in 1 2 3; do \
		go test -bench=. ./internal/...; \
	done > benchmark_results
	benchstat benchmark_results_baseline benchmark_results

clean:
	rm -f micro
