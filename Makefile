.PHONY: build build-all build-quick clean install install-all install-quick runtime runtime test 

VERSION := $(shell GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/build-version.go)
HASH := $(shell git rev-parse --short HEAD)
DATE := $(shell GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/build-date.go)
ADDITIONAL_GO_LINKER_FLAGS := $(shell GOOS=$(shell go env GOHOSTOS) \
	GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/info-plist.go "$(VERSION)")
GOBIN ?= $(shell go env GOPATH)/bin
.DEFAULT_TARGET := cmd/micro/micro

# Build micro without updating the runtime
build:
	cd cmd/micro && go build -ldflags "-s -w -X main.Version=$(VERSION) -X main.CommitHash=$(HASH) -X 'main.CompileDate=$(DATE)' $(ADDITIONAL_GO_LINKER_FLAGS)"

# Builds micro after building the runtime and checking dependencies
build-all: runtime build

# Builds micro without checking for dependencies
build-quick:
	cd cmd/micro && go build -ldflags "-s -w -X main.Version=$(VERSION) -X main.CommitHash=$(HASH) -X 'main.CompileDate=$(DATE)' $(ADDITIONAL_GO_LINKER_FLAGS)"

# Same as 'build' but installs to $GOBIN afterward
install:
	cd cmd/micro && go install -ldflags "-s -w -X main.Version=$(VERSION) -X main.CommitHash=$(HASH) -X 'main.CompileDate=$(DATE)' $(ADDITIONAL_GO_LINKER_FLAGS)"

# Same as 'build-all' but installs to $GOBIN afterward
install-all: runtime install

# Same as 'build-quick' but installs to $GOBIN afterward
install-quick:
	cd cmd/micro && go install -ldflags "-s -w -X main.Version=$(VERSION) -X main.CommitHash=$(HASH) -X 'main.CompileDate=$(DATE)' $(ADDITIONAL_GO_LINKER_FLAGS)"

# Builds the runtime
runtime:
	go get -u github.com/jteeuwen/go-bindata/...
	$(GOBIN)/go-bindata -nometadata -o runtime.go runtime/...
	mv runtime.go cmd/micro
	gofmt -w cmd/micro/runtime.go

test:
	cd cmd/micro && go test

clean:
	rm -f cmd/micro/micro
