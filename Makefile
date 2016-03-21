build: syn-files
	make build -C src
	mv src/micro .

install: syn-files
	make install -C src

syn-files:
	mkdir -p ~/.micro/syntax
	cp syntax_files/* ~/.micro/syntax
