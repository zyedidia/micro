cd ../..

tar czf "$1".tar.gz micro
zip -rq "$1".zip micro
mv "$1".tar.gz micro
mv "$1".zip micro
