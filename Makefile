.PHONY: runtime

VERSION = "$(shell git rev-parse --short HEAD)"

deps:
	go get -d ./cmd/micro

build: deps runtime tcell
	go build -ldflags "-X main.Version=$(VERSION)" -o micro ./cmd/micro

install: build
	mv micro $(GOPATH)/bin

tcell:
	cd $(GOPATH)/src/github.com/gdamore/tcell
	git pull

runtime:
	go get -u github.com/jteeuwen/go-bindata/...
	$(GOPATH)/bin/go-bindata -o runtime.go runtime/...
	mv runtime.go cmd/micro

test:
	go get -d ./cmd/micro
	go test ./cmd/micro

clean:
	rm -f micro
