#/bin/bash
BUCKET=$1

if [ x"$BUCKET" == "x" ]; then
	echo "needs buckets"
	exit 1
fi

echo "deploying to BUCKET: $BUCKET"
gsutil stat gs://$BUCKET/LICENSE
cd .
echo "PWD: $PWD"
find .
./mailman-summarizer -rss > feed.xml
ls -l
gsutil cp -a public-read feed.xml gs://$BUCKET/feed.xml

