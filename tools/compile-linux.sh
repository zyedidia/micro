cd ..

mkdir -p binaries
mkdir -p micro-$1

cp LICENSE micro-$1
cp README.md micro-$1
cp LICENSE-THIRD-PARTY micro-$1

HASH="$(git rev-parse --short HEAD)"
VERSION="$(go run tools/build-version.go)"
DATE="$(go run tools/build-date.go)"
ADDITIONAL_GO_LINKER_FLAGS="$(go run tools/info-plist.go $VERSION)"

echo "Linux 64"
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.Version=$1 -X main.CommitHash=$HASH -X 'main.CompileDate=$DATE'" -o micro-$1/micro ./cmd/micro
tar -czf micro-$1-linux64.tar.gz micro-$1
mv micro-$1-linux64.tar.gz binaries
echo "Linux 32"
GOOS=linux GOARCH=386 go build -ldflags "-s -w -X main.Version=$1 -X main.CommitHash=$HASH -X 'main.CompileDate=$DATE'" -o micro-$1/micro ./cmd/micro
tar -czf micro-$1-linux32.tar.gz micro-$1
mv micro-$1-linux32.tar.gz binaries
echo "Linux arm 32"
GOOS=linux GOARM=6 GOARCH=arm go build -ldflags "-s -w -X main.Version=$1 -X main.CommitHash=$HASH -X 'main.CompileDate=$DATE'" -o micro-$1/micro ./cmd/micro
tar -czf micro-$1-linux-arm.tar.gz micro-$1
mv micro-$1-linux-arm.tar.gz binaries
echo "Linux arm 64"
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.Version=$1 -X main.CommitHash=$HASH -X 'main.CompileDate=$DATE'" -o micro-$1/micro ./cmd/micro
tar -czf micro-$1-linux-arm64.tar.gz micro-$1
mv micro-$1-linux-arm64.tar.gz binaries

rm -rf micro-$1
