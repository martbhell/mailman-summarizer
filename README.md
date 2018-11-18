[![Build Status](https://travis-ci.com/martbhell/mailman-summarizer.svg?branch=master)](https://travis-ci.com/martbhell/mailman-summarizer)

mailman-summarizer
==========

General Idea:
---------

 - http://lists.ceph.com/pipermail/ceph-users-ceph.com/
 - Find last month's thread archive like http://lists.ceph.com/pipermail/ceph-users-ceph.com/2018-November/thread.html
 - Find threads with one or more patterns in the subject
 - Return something like:
   - An RSS feed (could take an existing RSS feed as an input and add another element to it)
   - An e-mail
   - To stdout
   - JSON

Usage:
-------

installing go:

 export GOPATH=$HOME/go

dependencies:

 go get -u github.com/PuerkitoBio/goquery
 go get -u github.com/gocolly/colly/...

building:

 go build

running: 

 ./mailman-summarizer


Sources
=====

 - https://benjamincongdon.me/blog/2018/03/01/Scraping-the-Web-in-Golang-with-Colly-and-Goquery/
 - https://github.com/bcongdon/colly-example

https://github.com/PuerkitoBio/goquery
https://github.com/gocolly/colly
