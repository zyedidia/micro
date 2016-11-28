cd ../cmd/micro

govendor init
govendor add +e

cd ../../..

tar czf "$1".tar.gz micro
zip -rq "$1".zip micro
mv "$1".tar.gz micro
mv "$1".zip micro

cd micro/cmd/micro
rm -rf vendor
