.PHONY: runtime

VERSION = $(shell go run tools/build-version.go)
HASH = $(shell git rev-parse --short HEAD)
DATE = $(shell go run tools/build-date.go)
ADDITIONAL_GO_LINKER_FLAGS = $(shell go run tools/info-plist.go "$(VERSION)")

GOBIN ?= $(GOPATH)/bin

# Builds micro after checking dependencies but without updating the runtime
build: deps
	go build -ldflags "-s -w -X main.Version=$(VERSION) -X main.CommitHash=$(HASH) -X 'main.CompileDate=$(DATE)' $(ADDITIONAL_GO_LINKER_FLAGS)" ./cmd/micro

# Builds micro after building the runtime and checking dependencies
build-all: runtime build

# Builds micro without checking for dependencies
build-quick:
	go build -ldflags "-s -w -X main.Version=$(VERSION) -X main.CommitHash=$(HASH) -X 'main.CompileDate=$(DATE)' $(ADDITIONAL_GO_LINKER_FLAGS)" ./cmd/micro

# Same as 'build' but installs to $GOBIN afterward
install: deps
	go install -ldflags "-s -w -X main.Version=$(VERSION) -X main.CommitHash=$(HASH) -X 'main.CompileDate=$(DATE)' $(ADDITIONAL_GO_LINKER_FLAGS)" ./cmd/micro

# Same as 'build-all' but installs to $GOBIN afterward
install-all: runtime install

# Same as 'build-quick' but installs to $GOBIN afterward
install-quick:
	go install -ldflags "-s -w -X main.Version=$(VERSION) -X main.CommitHash=$(HASH) -X 'main.CompileDate=$(DATE)' $(ADDITIONAL_GO_LINKER_FLAGS)"  ./cmd/micro

# Checks for dependencies
deps:
	go get -d ./cmd/micro

update:
	git pull
	go get -u -d ./cmd/micro

# Builds the runtime
runtime:
	go get -u github.com/jteeuwen/go-bindata/...
	$(GOBIN)/go-bindata -nometadata -o runtime.go runtime/...
	mv runtime.go cmd/micro

test:
	go get -d ./cmd/micro
	go test ./cmd/micro

clean:
	rm -f micro
