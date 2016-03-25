cd ..

mkdir -p binaries
mkdir -p micro/bin
cp -r runtime micro/

echo 'mkdir -p ~/.micro && cp -r runtime/* ~/.micro' >> micro/install.sh
chmod +x micro/install.sh

# Mac
echo "OSX 64"
GOOS=darwin GOARCH=amd64 go build -o micro/bin/micro ./src
tar -czf micro-osx.tar.gz micro
mv micro-osx.tar.gz binaries

# Linux
echo "Linux 64"
GOOS=linux GOARCH=amd64 go build -o micro/bin/micro ./src
tar -czf micro-linux64.tar.gz micro
mv micro-linux64.tar.gz binaries
echo "Linux 32"
GOOS=linux GOARCH=386 go build -o micro/bin/micro ./src
tar -czf micro-linux32.tar.gz micro
mv micro-linux32.tar.gz binaries
echo "Linux arm"
GOOS=linux GOARCH=arm go build -o micro/bin/micro ./src
tar -czf micro-linux-arm.tar.gz micro
mv micro-linux-arm.tar.gz binaries

rm -rf micro

# No windows building right now
# echo 'move runtime %HOMEPATH%\.micro' >> micro/install.bat
# chmod +x micro/install.bat
# Windows
# echo "Windows 64"
# GOOS=windows GOARCH=amd64 go build -o micro/bin/micro.exe ./src
# zip -r -q -T micro-win64.zip micro
# mv micro-win64.zip binaries
# echo "Windows 32"
# GOOS=windows GOARCH=386 go build -o micro/bin/micro.exe ./src
# zip -r -q -T micro-win32.zip micro
# mv micro-win32.zip binaries
