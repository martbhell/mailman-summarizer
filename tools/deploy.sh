#/bin/bash
BUCKET=$1

if [ x"$BUCKET" == "x" ]; then
	echo "needs buckets"
	exit 1
fi

echo "TEST: deploying to BUCKET: $BUCKET"
gsutil stat gs://$BUCKET/feed.xml

echo "TEST: fetching feed.xml before compiling and putting it in /tmp/feed2.xml"
wget https://storage.googleapis.com/ceph-rgw-users/feed.xml -O /tmp/feed2.xml

# https://stackoverflow.com/questions/28081486/how-can-i-go-run-a-project-with-multiple-files-in-the-main-package
# go run all except *_test.go files (handy when we get around to add testing :)
shopt -s extglob
go run !(*_test).go -rss -topic GW,S3,keystone,civet > feed.xml
echo "TEST: Showing diff between old and new feed.xml"
diff -u /tmp/feed2.xml feed.xml
echo "TEST: listing dir"
ls -l
echo "DEPLOY: uploading feed.xml publicly to $BUCKET"
gsutil cp -a public-read feed.xml gs://$BUCKET/feed.xml

