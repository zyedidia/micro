.PHONY: runtime build generate build-quick

VERSION = $(shell GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/build-version.go)
HASH = $(shell git rev-parse --short HEAD)
DATE = $(shell GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/build-date.go)
ADDITIONAL_GO_LINKER_FLAGS = $(shell GOOS=$(shell go env GOHOSTOS) \
	GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/info-plist.go "$(shell go env GOOS)" "$(VERSION)")
GOBIN ?= $(shell go env GOPATH)/bin
GOVARS = -X github.com/zyedidia/micro/v2/internal/util.Version=$(VERSION) -X github.com/zyedidia/micro/v2/internal/util.CommitHash=$(HASH) -X 'github.com/zyedidia/micro/v2/internal/util.CompileDate=$(DATE)'
DEBUGVAR = -X github.com/zyedidia/micro/v2/internal/util.Debug=ON
VSCODE_TESTS_BASE_URL = 'https://raw.githubusercontent.com/microsoft/vscode/e6a45f4242ebddb7aa9a229f85555e8a3bd987e2/src/vs/editor/test/common/model/'

build: generate build-quick

build-quick:
	go build -trimpath -ldflags "-s -w $(GOVARS) $(ADDITIONAL_GO_LINKER_FLAGS)" ./cmd/micro

build-dbg:
	go build -trimpath -ldflags "-s -w $(ADDITIONAL_GO_LINKER_FLAGS) $(DEBUGVAR)" ./cmd/micro

build-tags: fetch-tags generate
	go build -trimpath -ldflags "-s -w $(GOVARS) $(ADDITIONAL_GO_LINKER_FLAGS)" ./cmd/micro

build-all: build

install: generate
	go install -ldflags "-s -w $(GOVARS) $(ADDITIONAL_GO_LINKER_FLAGS)" ./cmd/micro

install-all: install

fetch-tags:
	git fetch --tags

generate:
	GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) go generate ./runtime

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
	go test ./cmd/...

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
	benchstat -alpha 0.15 benchmark_results_baseline benchmark_results

clean:
	rm -f micro
