build:
	go get -d ./cmd/micro
	go build -o micro ./cmd/micro

install:
	go get -d ./cmd/micro
	go install -o micro ./cmd/micro

runtime:
	go get -u github.com/jteeuwen/go-bindata/...
	go-bindata -o runtime.go data/
	mv runtime.go cmd/micro

test:
	go get -d ./cmd/micro
	go test ./cmd/micro

clean:
	rm -f micro
