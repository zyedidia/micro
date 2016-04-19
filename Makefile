NAME=micro
RELEASE:=$(shell git rev-parse --verify --short HEAD)

.PHONY: runtime

build: runtime
	go get -d ./cmd/micro
	go build -o micro ./cmd/micro

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

cross:
	echo "This will take a while."
	date
	go get -u github.com/jteeuwen/go-bindata/...
	mkdir bin || true
	GOOS=windows GOARCH=386 $(GOPATH)/bin/go-bindata -o runtime.go runtime/...
	mv runtime.go cmd/micro
	GOOS=windows GOARCH=386 go build -v -o binaries/${NAME}-${RELEASE}-WIN32.exe ./cmd/micro
	GOOS=windows GOARCH=amd64 $(GOPATH)/bin/go-bindata -o runtime.go runtime/...
	mv runtime.go cmd/micro
	GOOS=windows GOARCH=amd64 go build -v -o binaries/${NAME}-${RELEASE}-WIN64.exe ./cmd/micro
	GOOS=darwin GOARCH=386 $(GOPATH)/bin/go-bindata -o runtime.go runtime/...
	mv runtime.go cmd/micro
	GOOS=darwin GOARCH=386 go build -v -o binaries/${NAME}-${RELEASE}-OSX-x86 ./cmd/micro
	GOOS=darwin GOARCH=amd64 $(GOPATH)/bin/go-bindata -o runtime.go runtime/...
	mv runtime.go cmd/micro
	GOOS=darwin GOARCH=amd64 go build -v -o binaries/${NAME}-${RELEASE}-OSX-amd64 ./cmd/micro
	GOOS=linux GOARCH=amd64 $(GOPATH)/bin/go-bindata -o runtime.go runtime/...
	mv runtime.go cmd/micro
	GOOS=linux GOARCH=amd64 go build -v -o binaries/${NAME}-${RELEASE}-linux-amd64 ./cmd/micro
	GOOS=linux GOARCH=386 $(GOPATH)/bin/go-bindata -o runtime.go runtime/...
	mv runtime.go cmd/micro
	GOOS=linux GOARCH=386 go build -v -o binaries/${NAME}-${RELEASE}-linux-x86 ./cmd/micro
	GOOS=freebsd GOARCH=amd64 $(GOPATH)/bin/go-bindata -o runtime.go runtime/...
	mv runtime.go cmd/micro
	GOOS=freebsd GOARCH=amd64 go build -v -o binaries/${NAME}-${RELEASE}-freebsd-amd64 ./cmd/micro
	GOOS=freebsd GOARCH=386 $(GOPATH)/bin/go-bindata -o runtime.go runtime/...
	mv runtime.go cmd/micro
	GOOS=freebsd GOARCH=386 go build -v -o binaries/${NAME}-${RELEASE}-freebsd-x86 ./cmd/micro
	GOOS=openbsd GOARCH=amd64 $(GOPATH)/bin/go-bindata -o runtime.go runtime/...
	mv runtime.go cmd/micro
	GOOS=openbsd GOARCH=amd64 go build -v -o binaries/${NAME}-${RELEASE}-openbsd-amd64 ./cmd/micro
	GOOS=openbsd GOARCH=386 $(GOPATH)/bin/go-bindata -o runtime.go runtime/...
	mv runtime.go cmd/micro
	GOOS=openbsd GOARCH=386 go build -v -o binaries/${NAME}-${RELEASE}-openbsd-x86 ./cmd/micro
	GOOS=netbsd GOARCH=amd64 $(GOPATH)/bin/go-bindata -o runtime.go runtime/...
	mv runtime.go cmd/micro
	GOOS=netbsd GOARCH=amd64 go build -v -o binaries/${NAME}-${RELEASE}-netbsd-amd64 ./cmd/micro
	GOOS=netbsd GOARCH=386 $(GOPATH)/bin/go-bindata -o runtime.go runtime/...
	mv runtime.go cmd/micro
	GOOS=netbsd GOARCH=386 go build -v -o binaries/${NAME}-${RELEASE}-netbsd-x86	./cmd/micro
	date
	echo "Now run: make pkg"
pkg:
	mkdir pkg-bin || true
	for i in $(shell ls binaries); do echo $$i; mkdir micro-${RELEASE};cp binaries/$$i micro-${RELEASE}; cp README.md micro-${RELEASE}; cp LICENSE micro-${RELEASE}; zip -r pkg-bin/$$i.zip micro-${RELEASE}; rm -Rf micro-${RELEASE}; done;
