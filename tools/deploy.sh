#/bin/bash
BUCKET=$1

if [ x"$BUCKET" == "x" ]; then
	echo "needs buckets"
	exit 1
fi

echo "deploying to BUCKET: $BUCKET"
gsutil stat gs://$BUCKET/feed.xml
# https://stackoverflow.com/questions/28081486/how-can-i-go-run-a-project-with-multiple-files-in-the-main-package
# go run all except *_test.go files (handy when we get around to add testing :)
shopt -s extglob
go run !(*_test).go -rss -topic GW,S3,keystone > feed.xml
ls -l
gsutil cp -a public-read feed.xml gs://$BUCKET/feed.xml

