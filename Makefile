.PHONY: runtime

VERSION := $(shell GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/build-version.go)
HASH := $(shell git rev-parse --short HEAD)
DATE := $(shell GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/build-date.go)
ADDITIONAL_GO_LINKER_FLAGS := $(shell GOOS=$(shell go env GOHOSTOS) \
	GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/info-plist.go "$(VERSION)")
GOBIN ?= $(shell go env GOPATH)/bin
GOVARS := -X github.com/zyedidia/micro/internal/util.Version=$(VERSION) -X github.com/zyedidia/micro/internal/util.CommitHash=$(HASH) -X 'github.com/zyedidia/micro/internal/util.CompileDate=$(DATE)' -X github.com/zyedidia/micro/internal/util.Debug=OFF

# Builds micro after checking dependencies but without updating the runtime
build:
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

# Builds the runtime
runtime:
	git submodule update --init
	go build -o tools/bindata ./tools/go-bindata
	tools/bindata -pkg config -nomemcopy -nometadata -o runtime.go runtime/...
	mv runtime.go internal/config
	gofmt -w internal/config/runtime.go

test:
	go test ./internal/...

clean:
	rm -f micro
