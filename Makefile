build:
	go get -d ./cmd/micro
	go build -o micro ./cmd/micro

install: build
	mv micro $(GOPATH)/bin

runtime:
	go get -u github.com/jteeuwen/go-bindata/...
	go-bindata -o runtime.go data/
	mv runtime.go cmd/micro

test:
	go get -d ./cmd/micro
	go test ./cmd/micro

clean:
	rm -f micro
