#This script builds four rpm packages
#One for x86 (i386) and x86_64 (amd64) and arm (armv7l)
#and one containing the source tarball
version=$1
if [ "$1" == "" ] 
   then
     version=$(go run build-version.go | tr "-" ".")
fi
echo "Building packages for Version '$version'"
echo "Compiling."
./compile-linux.sh $version

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
echo "Starting the packaging process"
#Generate the spec file
cat micro.spec | sed s/"dev.126"/"$dev"/ | sed s/"Version: 1.1.5"/"Version: $version"/ | sed s/"-Version: 1.1.5"/"-Version: $version"/ | sed s/"DATE"/"$(date +%F\ %H:%m)"/ | sed s/"rdieter1@localhost.localdomain"/"$USER@$HOSTNAME"/ | tee > $PKGPATH/micro.spec

cd $PKGPATH

rpmbuild -bs micro.spec --define "_sourcedir $(pwd)/../../binaries/" --define "_rpmdir $(pwd)/../"
rpmbuild -bb micro.spec --define "_sourcedir $(pwd)/../../binaries/" --define "_rpmdir $(pwd)/../" --target x86_64
rpmbuild -bb micro.spec --define "_sourcedir $(pwd)/../../binaries/" --define "_rpmdir $(pwd)/../" --target i686
rpmbuild -bb micro.spec --define "_sourcedir $(pwd)/../../binaries/" --define "_rpmdir $(pwd)/../" --target armv7l

cd ..

mv x86_64/micro-$version-1.$dev.x86_64.rpm ./
mv i686/micro-$version-1.$dev.i686.rpm ./
mv armv7l/micro-$version-1.$dev.armv7l.rpm ./

rm -rf x86_64
rm -rf i686
rm -rf armv7l
