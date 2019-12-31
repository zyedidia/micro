# This script creates releases on Github for micro
# You must have the correct Github access token to run this script

# $1 is the title, $2 is the description

commitID=$(git rev-parse HEAD)
tag="v$1"

echo "Creating tag"
git tag $tag $commitID
git push --tags

echo "Creating new release"
github-release release \
    --user zyedidia \
    --repo micro \
    --tag $tag \
    --name "$1" \
    --description "$2" \

echo "Cross compiling binaries"
./cross-compile.sh $1
mv ../binaries .

echo "Uploading OSX binary"
github-release upload \
    --user zyedidia \
    --repo micro \
    --tag $tag \
    --name "micro-$1-osx.tar.gz" \
    --file binaries/micro-$1-osx.tar.gz

echo "Uploading Linux 64 binary"
github-release upload \
    --user zyedidia \
    --repo micro \
    --tag $tag \
    --name "micro-$1-linux64.tar.gz" \
    --file binaries/micro-$1-linux64.tar.gz

echo "Uploading Linux 64 static binary"
github-release upload \
    --user zyedidia \
    --repo micro \
    --tag $tag \
    --name "micro-$1-linux64-static.tar.gz" \
    --file binaries/micro-$1-linux64-static.tar.gz

echo "Uploading Linux 32 binary"
github-release upload \
    --user zyedidia \
    --repo micro \
    --tag $tag \
    --name "micro-$1-linux32.tar.gz" \
    --file binaries/micro-$1-linux32.tar.gz

echo "Uploading Linux Arm 32 binary"
github-release upload \
    --user zyedidia \
    --repo micro \
    --tag $tag \
    --name "micro-$1-linux-arm.tar.gz" \
    --file binaries/micro-$1-linux-arm.tar.gz

echo "Uploading Linux Arm 64 binary"
github-release upload \
    --user zyedidia \
    --repo micro \
    --tag $tag \
    --name "micro-$1-linux-arm64.tar.gz" \
    --file binaries/micro-$1-linux-arm64.tar.gz

echo "Uploading FreeBSD 64 binary"
github-release upload \
    --user zyedidia \
    --repo micro \
    --tag $tag \
    --name "micro-$1-freebsd64.tar.gz" \
    --file binaries/micro-$1-freebsd64.tar.gz

echo "Uploading FreeBSD 32 binary"
github-release upload \
    --user zyedidia \
    --repo micro \
    --tag $tag \
    --name "micro-$1-freebsd32.tar.gz" \
    --file binaries/micro-$1-freebsd32.tar.gz

echo "Uploading OpenBSD 64 binary"
github-release upload \
    --user zyedidia \
    --repo micro \
    --tag $tag \
    --name "micro-$1-openbsd64.tar.gz" \
    --file binaries/micro-$1-openbsd64.tar.gz

echo "Uploading OpenBSD 32 binary"
github-release upload \
    --user zyedidia \
    --repo micro \
    --tag $tag \
    --name "micro-$1-openbsd32.tar.gz" \
    --file binaries/micro-$1-openbsd32.tar.gz

echo "Uploading NetBSD 64 binary"
github-release upload \
    --user zyedidia \
    --repo micro \
    --tag $tag \
    --name "micro-$1-netbsd64.tar.gz" \
    --file binaries/micro-$1-netbsd64.tar.gz

echo "Uploading NetBSD 32 binary"
github-release upload \
    --user zyedidia \
    --repo micro \
    --tag $tag \
    --name "micro-$1-netbsd32.tar.gz" \
    --file binaries/micro-$1-netbsd32.tar.gz

echo "Uploading Windows 64 binary"
github-release upload \
    --user zyedidia \
    --repo micro \
    --tag $tag \
    --name "micro-$1-win64.zip" \
    --file binaries/micro-$1-win64.zip

echo "Uploading Windows 32 binary"
github-release upload \
    --user zyedidia \
    --repo micro \
    --tag $tag \
    --name "micro-$1-win32.zip" \
    --file binaries/micro-$1-win32.zip

# echo "Uploading vendored tarball"
# github-release upload \
#     --user zyedidia \
#     --repo micro \
#     --tag $tag \
#     --name "micro-$1-src.tar.gz" \
#     --file binaries/micro-$1-src.tar.gz
#
# echo "Uploading vendored zip"
# github-release upload \
#     --user zyedidia \
#     --repo micro \
#     --tag $tag \
#     --name "micro-$1-src.zip" \
#     --file binaries/micro-$1-src.zip
