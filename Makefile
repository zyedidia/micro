.PHONY: runtime

VERSION = $(shell git describe --tags --abbrev=0)
HASH = $(shell git rev-parse --short HEAD)
DATE = $(shell python -c 'import time; print(time.strftime("%B %d, %Y"))')

# Builds micro after checking dependencies but without updating the runtime
build: deps tcell
	go build -ldflags "-X main.Version=$(VERSION) -X main.CommitHash=$(HASH) -X 'main.CompileDate=$(DATE)'" -o micro ./cmd/micro

# Builds micro after building the runtime and checking dependencies
build-all: runtime build

# Builds micro without checking for dependencies
build-quick:
	go build -ldflags "-X main.Version=$(VERSION) -X main.CommitHash=$(HASH) -X 'main.CompileDate=$(DATE)'" -o micro ./cmd/micro

# Same as 'build' but installs to $GOPATH/bin afterward
install: build
	mv micro $(GOPATH)/bin

# Same as 'build-all' but installs to $GOPATH/bin afterward
install-all: runtime install

# Same as 'build-quick' but installs to $GOPATH/bin afterward
install-quick: build-quick
	mv micro $(GOPATH)/bin

# Updates tcell
tcell:
	git -C $(GOPATH)/src/github.com/zyedidia/tcell pull

# Checks for dependencies
deps:
	go get -d ./cmd/micro

# Builds the runtime
runtime:
	go get -u github.com/jteeuwen/go-bindata/...
	$(GOPATH)/bin/go-bindata -nometadata -o runtime.go runtime/...
	mv runtime.go cmd/micro

test:
	go get -d ./cmd/micro
	go test ./cmd/micro

clean:
	rm -f micro
