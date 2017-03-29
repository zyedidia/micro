# Builds two .deb packages, for x86 (i386) and x86_64 (amd64)
# These packages are the bare minimum, which means that they can be installed
# But they do not feature everything yet.
# This does not mean that the editor itself is affected.

function getControl() {
echo Section: editors
echo Package: micro
echo Version: $2
echo Priority: extra
echo Maintainer: \"Zachary Yedidia\" \<zyedidia@gmail.com\>
echo Standards-Version: 3.9.8
echo Homepage: https://micro-editor.github.io/
echo Architecture: $1
echo "Description: A modern and intuitive terminal-based text editor"
echo " This package contains a modern alternative to other terminal-based"
echo " Editors. It is easy to Use, highly customizable via themes and plugins"
echo " and it supports mouse input"
}

function installFiles() {
  TO="$1/$2/usr/share/doc/micro/"
  mkdir -p $TO
  mkdir -p "$1/$2/usr/share/man/man1/"
  mkdir -p "$1/$2/usr/share/applications/"
  mkdir -p "$1/$2/usr/share/icons/"
  cp ../LICENSE $TO
  cp ../LICENSE-THIRD-PARTY $TO
  cp ../README.md $TO
  gzip -c ../assets/packaging/micro.1 > $1/$2/usr/share/man/man1/micro.1.gz
  cp ../assets/packaging/micro.desktop $1/$2/usr/share/applications/
  cp ../assets/logo.svg $1/$2/usr/share/icons/micro.svg
}

version=$1
if [ "$1" == "" ]
then
  version=$(go run build-version.go)
fi
echo "Building packages for Version '$version'"
echo "Running Cross-Compile"
./cross-compile.sh $version

echo "Beginning package build process"

PKGPATH="../packages/deb"

rm -fr $PKGPATH
mkdir -p $PKGPATH/amd64/DEBIAN/
mkdir -p $PKGPATH/i386/DEBIAN/

getControl "amd64" "$version" > $PKGPATH/amd64/DEBIAN/control
tar -xzf "../binaries/micro-$version-linux64.tar.gz" "micro-$version/micro"
mkdir -p $PKGPATH/amd64/usr/local/bin/
mv "micro-$version/micro" "$PKGPATH/amd64/usr/local/bin/"

getControl "i386" "$version" > $PKGPATH/i386/DEBIAN/control
tar -xzf "../binaries/micro-$version-linux32.tar.gz" "micro-$version/micro"
mkdir -p $PKGPATH/i386/usr/local/bin/
mv "micro-$version/micro" "$PKGPATH/i386/usr/local/bin/"

rm -rf "micro-$version"

installFiles $PKGPATH "amd64"
installFiles $PKGPATH "i386"

dpkg -b "$PKGPATH/amd64/" "../packages/micro-$version-amd64.deb"
dpkg -b "$PKGPATH/i386/" "../packages/micro-$version-i386.deb"
