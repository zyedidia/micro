#!/bin/sh

set -e

VERSION="$1"
if [ -z "$VERSION" ]; then
	VERSION="$(go run tools/build-version.go)"
fi

mkdir -p binaries
mkdir -p micro-$VERSION

cp LICENSE micro-$VERSION
cp README.md micro-$VERSION
cp LICENSE-THIRD-PARTY micro-$VERSION
cp assets/packaging/micro.1 micro-$VERSION
cp assets/packaging/micro.desktop micro-$VERSION
cp assets/micro-logo-mark.svg micro-$VERSION/micro.svg

create_artefact_generic()
{
	mv micro micro-$VERSION/
	tar -czf micro-$VERSION-$1.tar.gz micro-$VERSION
	sha256sum micro-$VERSION-$1.tar.gz > micro-$VERSION-$1.tar.gz.sha
	mv micro-$VERSION-$1.* binaries
	rm micro-$VERSION/micro
}

create_artefact_windows()
{
	mv micro.exe micro-$VERSION/
	zip -r -q -T micro-$VERSION-$1.zip micro-$VERSION
	sha256sum micro-$VERSION-$1.zip > micro-$VERSION-$1.zip.sha
	mv micro-$VERSION-$1.* binaries
	rm micro-$VERSION/micro.exe
}

# Mac
echo "OSX 64"
GOOS=darwin GOARCH=amd64 make build
create_artefact_generic "osx"

# Mac ARM64
echo "MacOS ARM64"
GOOS=darwin GOARCH=arm64 make build
create_artefact_generic "macos-arm64"

# Linux
echo "Linux 64"
GOOS=linux GOARCH=amd64 make build
if ./tools/package-deb.sh $VERSION; then
	sha256sum micro-$VERSION-amd64.deb > micro-$VERSION-amd64.deb.sha
	mv micro-$VERSION-amd64.* binaries
fi
create_artefact_generic "linux64"

echo "Linux 64 fully static (same as linux64)"
# It is kept for the next release only to support...
# https://github.com/benweissmann/getmic.ro/blob/f90870e948afab8be9ec40884050044b59ed5b7c/index.sh#L197-L204
# ...and allow a fluent switch via:
# https://github.com/benweissmann/getmic.ro/pull/40
GOOS=linux GOARCH=amd64 make build
create_artefact_generic "linux64-static"

echo "Linux 32"
GOOS=linux GOARCH=386 make build
create_artefact_generic "linux32"

echo "Linux ARM 32"
GOOS=linux GOARM=6 GOARCH=arm make build
create_artefact_generic "linux-arm"

echo "Linux ARM 64"
GOOS=linux GOARCH=arm64 make build
create_artefact_generic "linux-arm64"

# Solaris
echo "Solaris 64"
GOOS=solaris GOARCH=amd64 make build
create_artefact_generic "solaris64"

# Illumos
echo "Illumos 64"
GOOS=illumos GOARCH=amd64 make build
create_artefact_generic "illumos64"

# NetBSD
echo "NetBSD 64"
GOOS=netbsd GOARCH=amd64 make build
create_artefact_generic "netbsd64"

echo "NetBSD 32"
GOOS=netbsd GOARCH=386 make build
create_artefact_generic "netbsd32"

# OpenBSD
echo "OpenBSD 64"
GOOS=openbsd GOARCH=amd64 make build
create_artefact_generic "openbsd64"

echo "OpenBSD 32"
GOOS=openbsd GOARCH=386 make build
create_artefact_generic "openbsd32"

# FreeBSD
echo "FreeBSD 64"
GOOS=freebsd GOARCH=amd64 make build
create_artefact_generic "freebsd64"

echo "FreeBSD 32"
GOOS=freebsd GOARCH=386 make build
create_artefact_generic "freebsd32"

# Windows
echo "Windows 64"
GOOS=windows GOARCH=amd64 make build
create_artefact_windows "win64"

echo "Windows ARM 64"
GOOS=windows GOARCH=arm64 make build
create_artefact_windows "win-arm64"

echo "Windows 32"
GOOS=windows GOARCH=386 make build
create_artefact_windows "win32"

rm -rf micro-$VERSION
