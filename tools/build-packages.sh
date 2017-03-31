#Builds all packages we support

version=$1
if [ "$1" == "" ] 
   then
     version=$(go run build-version.go | tr "-" ".")
fi
echo "Building packages for Version '$version'"
echo "Compiling."
./compile-linux.sh $version

#Build the debs
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
echo "Starting deb build process"
PKGPATH="../packages/deb"
rm -fr $PKGPATH
mkdir -p $PKGPATH/amd64/DEBIAN/
mkdir -p $PKGPATH/i386/DEBIAN/
mkdir -p $PKGPATH/arm/DEBIAN/

getControl "amd64" "$version" > $PKGPATH/amd64/DEBIAN/control
tar -xzf "../binaries/micro-$version-linux64.tar.gz" "micro-$version/micro"
mkdir -p $PKGPATH/amd64/usr/local/bin/
mv "micro-$version/micro" "$PKGPATH/amd64/usr/local/bin/"
  
getControl "i386" "$version" > $PKGPATH/i386/DEBIAN/control
tar -xzf "../binaries/micro-$version-linux32.tar.gz" "micro-$version/micro"
mkdir -p $PKGPATH/i386/usr/local/bin/
mv "micro-$version/micro" "$PKGPATH/i386/usr/local/bin/"
        
getControl "arm" "$version" > $PKGPATH/arm/DEBIAN/control
tar -xzf "../binaries/micro-$version-linux-arm.tar.gz" "micro-$version/micro"
mkdir -p $PKGPATH/arm/usr/local/bin
mv "micro-$version/micro" "$PKGPATH/arm/usr/local/bin"
        
rm -rf "micro-$version"
        
installFiles $PKGPATH "amd64"
installFiles $PKGPATH "i386"
installFiles $PKGPATH "arm"

echo "Building debs"        
dpkg -b "$PKGPATH/amd64/" "../packages/micro-$version-amd64.deb"
dpkg -b "$PKGPATH/i386/" "../packages/micro-$version-i386.deb"
dpkg -b "$PKGPATH/arm/" "../packages/micro-$version-arm.deb"
#Build the RPMS
echo "Starting RPM build process"
PKGPATH="../packages/rpm"

rm -rf $PKGPATH
mkdir -p $PKGPATH

versionsplit=$(echo $version | tr "." "\n")
version=""
i=0
for string in $versionsplit
do
	if (("$i" < "2")) 
	then
		version=$(echo $version$string.)
	fi
	if (("$i" == "2")) 
	then
		version=$(echo $version$string)
	fi
	if (("$i" == "3")) 
	then
		dev=$(echo $dev$string.)
	fi
	if (("$i"=="4"))
	then
		dev=$(echo $dev$string)
	fi
	let "i+=1"
done
echo "Building the RPM packages"
rpmbuild -bs micro.spec --define "_sourcedir $(pwd)/../../binaries/" --define "_rpmdir $(pwd)/../"
rpmbuild -bb micro.spec --define "_sourcedir $(pwd)/../../binaries/" --define "_rpmdir $(pwd)/../" --target x86_64
rpmbuild -bb micro.spec --define "_sourcedir $(pwd)/../../binaries/" --define "_rpmdir $(pwd)/../" --target i686
rpmbuild -bb micro.spec --define "_sourcedir $(pwd)/../../binaries/" --define "_rpmdir $(pwd)/../" --target armv7l

cd ..

mv x86_64/micro-$version-1.$dev.x86_64.rpm ./
mv i686/micro-$version-1.$dev.i686.rpm ./
mv armv7l/micro-$version-1.$dev.armv7l.rpm ./

echo "Cleaning up."
rm -rf x86_64
rm -rf i686
rm -rf armv7l
rm -rf rpm
rm -rf deb

echo "Your packages should be ready now. Thank you, have a nice day. :)"
