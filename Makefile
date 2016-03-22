build: syn-files
	go get -d ./src
	go build -o micro ./src

install: syn-files build
	mv micro $(GOBIN)

syn-files:
	mkdir -p ~/.micro/syntax
	cp syntax_files/* ~/.micro/syntax
