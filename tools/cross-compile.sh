cd ..

mkdir -p binaries
mkdir -p micro

cp LICENSE micro
cp README.md micro

# Mac
echo "OSX 64"
GOOS=darwin GOARCH=amd64 go build -o micro/micro ./cmd/micro
tar -czf micro-osx.tar.gz micro
mv micro-osx.tar.gz binaries

# Linux
echo "Linux 64"
GOOS=linux GOARCH=amd64 go build -o micro/micro ./cmd/micro
tar -czf micro-linux64.tar.gz micro
mv micro-linux64.tar.gz binaries
echo "Linux 32"
GOOS=linux GOARCH=386 go build -o micro/micro ./cmd/micro
tar -czf micro-linux32.tar.gz micro
mv micro-linux32.tar.gz binaries
echo "Linux arm"
GOOS=linux GOARCH=arm go build -o micro/micro ./cmd/micro
tar -czf micro-linux-arm.tar.gz micro
mv micro-linux-arm.tar.gz binaries

# NetBSD
echo "NetBSD 64"
GOOS=netbsd GOARCH=amd64 go build -o micro/micro ./cmd/micro
tar -czf micro-netbsd64.tar.gz micro
mv micro-netbsd64.tar.gz binaries
echo "NetBSD 32"
GOOS=netbsd GOARCH=386 go build -o micro/micro ./cmd/micro
tar -czf micro-netbsd32.tar.gz micro
mv micro-netbsd32.tar.gz binaries

# OpenBSD
echo "OpenBSD 64"
GOOS=openbsd GOARCH=amd64 go build -o micro/micro ./cmd/micro
tar -czf micro-openbsd64.tar.gz micro
mv micro-openbsd64.tar.gz binaries
echo "OpenBSD 32"
GOOS=openbsd GOARCH=386 go build -o micro/micro ./cmd/micro
tar -czf micro-openbsd32.tar.gz micro
mv micro-openbsd32.tar.gz binaries

# FreeBSD
echo "FreeBSD 64"
GOOS=freebsd GOARCH=amd64 go build -o micro/micro ./cmd/micro
tar -czf micro-freebsd64.tar.gz micro
mv micro-freebsd64.tar.gz binaries
echo "FreeBSD 32"
GOOS=freebsd GOARCH=386 go build -o micro/micro ./cmd/micro
tar -czf micro-freebsd32.tar.gz micro
mv micro-freebsd32.tar.gz binaries

rm micro/micro

# Windows
echo "Windows 64"
GOOS=windows GOARCH=amd64 go build -o micro/micro.exe ./cmd/micro
zip -r -q -T micro-win64.zip micro
mv micro-win64.zip binaries
echo "Windows 32"
GOOS=windows GOARCH=386 go build -o micro/micro.exe ./cmd/micro
zip -r -q -T micro-win32.zip micro
mv micro-win32.zip binaries

rm -rf micro
