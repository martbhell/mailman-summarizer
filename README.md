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

Ideas
=====

License: MIT
