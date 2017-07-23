.PHONY: runtime

# detect GOPATH if not set
ifndef $(GOPATH)
    $(info GOPATH is not set, autodetecting..)
    TESTPATH := $(dir $(abspath ../../..))
    DIRS := bin pkg src
    # create a ; separated line of tests and pass it to shell
    MISSING_DIRS := $(shell $(foreach entry,$(DIRS),test -d "$(TESTPATH)$(entry)" || echo "$(entry)";))
    ifeq ($(MISSING_DIRS),)
        $(info Found GOPATH: $(TESTPATH))
        export GOPATH := $(TESTPATH)
    else
        $(info ..missing dirs "$(MISSING_DIRS)" in "$(TESTDIR)")
        $(info GOPATH autodetection failed)
    endif
endif

VERSION := $(shell GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/build-version.go)
HASH := $(shell git rev-parse --short HEAD)
DATE := $(shell GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/build-date.go)
ADDITIONAL_GO_LINKER_FLAGS := $(shell GOOS=$(shell go env GOHOSTOS) \
	GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/info-plist.go "$(VERSION)")
GOBIN ?= $(shell go env GOPATH)/bin

# Builds micro after checking dependencies but without updating the runtime
build: update
	go build -ldflags "-s -w -X main.Version=$(VERSION) -X main.CommitHash=$(HASH) -X 'main.CompileDate=$(DATE)' $(ADDITIONAL_GO_LINKER_FLAGS)" ./cmd/micro

# Builds micro after building the runtime and checking dependencies
build-all: runtime build

# Builds micro without checking for dependencies
build-quick:
	go build -ldflags "-s -w -X main.Version=$(VERSION) -X main.CommitHash=$(HASH) -X 'main.CompileDate=$(DATE)' $(ADDITIONAL_GO_LINKER_FLAGS)" ./cmd/micro

# Same as 'build' but installs to $GOBIN afterward
install: update
	go install -ldflags "-s -w -X main.Version=$(VERSION) -X main.CommitHash=$(HASH) -X 'main.CompileDate=$(DATE)' $(ADDITIONAL_GO_LINKER_FLAGS)" ./cmd/micro

# Same as 'build-all' but installs to $GOBIN afterward
install-all: runtime install

# Same as 'build-quick' but installs to $GOBIN afterward
install-quick:
	go install -ldflags "-s -w -X main.Version=$(VERSION) -X main.CommitHash=$(HASH) -X 'main.CompileDate=$(DATE)' $(ADDITIONAL_GO_LINKER_FLAGS)"  ./cmd/micro

update:
	git pull
	git submodule update --init

# Builds the runtime
runtime:
	go get -u github.com/jteeuwen/go-bindata/...
	$(GOBIN)/go-bindata -nometadata -o runtime.go runtime/...
	mv runtime.go cmd/micro

test:
	go test ./cmd/micro

clean:
	rm -f micro
