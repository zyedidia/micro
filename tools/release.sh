# This script creates releases on Github for micro
# You must have the correct Github access token to run this script

# $1 is the title, $2 is the description

commitID=$(git rev-parse HEAD)
tag="v$1"

echo "Creating tag"
git tag $tag $commitID
hub push --tags

NL=$'\n'

echo "Cross compiling binaries"
./cross-compile.sh $1
mv ../binaries .

echo "Creating new release"
hub release create $tag \
    --message "$1${NL}${NL}$2" \
    --attach "binaries/micro-$1-osx.tar.gz" \
    --attach "binaries/micro-$1-macos-arm64.tar.gz" \
    --attach "binaries/micro-$1-linux64.tar.gz" \
    --attach "binaries/micro-$1-linux64-static.tar.gz" \
    --attach "binaries/micro-$1-amd64.deb" \
    --attach "binaries/micro-$1-linux32.tar.gz" \
    --attach "binaries/micro-$1-linux-arm.tar.gz" \
    --attach "binaries/micro-$1-linux-arm64.tar.gz" \
    --attach "binaries/micro-$1-freebsd64.tar.gz" \
    --attach "binaries/micro-$1-freebsd32.tar.gz" \
    --attach "binaries/micro-$1-openbsd64.tar.gz" \
    --attach "binaries/micro-$1-openbsd32.tar.gz" \
    --attach "binaries/micro-$1-netbsd64.tar.gz" \
    --attach "binaries/micro-$1-netbsd32.tar.gz" \
    --attach "binaries/micro-$1-win64.zip" \
    --attach "binaries/micro-$1-win32.zip"
