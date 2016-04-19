.PHONY: runtime

VERSION = "$(shell git rev-parse --short HEAD)"

build: runtime
	go get -d ./cmd/micro
	go build -ldflags "-X main.Version=$(VERSION)" -o micro ./cmd/micro

install: build runtime
	mv micro $(GOPATH)/bin

runtime:
	go get -u github.com/jteeuwen/go-bindata/...
	$(GOPATH)/bin/go-bindata -o runtime.go runtime/...
	mv runtime.go cmd/micro

test:
	go get -d ./cmd/micro
	go test ./cmd/micro

clean:
	rm -f micro
