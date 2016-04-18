build: syn-files
	go get -d ./cmd/micro
	go build -o micro ./cmd/micro

install: syn-files build
	mv micro $(GOBIN)

syn-files:
	mkdir -p ~/.micro
	cp -r runtime/* ~/.micro

test:
	go get -d ./cmd/micro
	go test ./cmd/micro

clean:
	rm -f micro
