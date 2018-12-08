package main

import (

	"fmt"
	"time"
	"github.com/gorilla/feeds" // making RSS
	"log"
	"strconv"

)

func makeRSS(keys []string, data map[string]map[string]string, arss *bool, ajson *bool, aatom *bool) {

	// keys = list of months
	// data = { "2018-November": { "thread1": "link1", "thread2": "link2", .. }, "2018-October": { "thread3": "link3", .. }, .. }
	// then some flags to control output

	// 03 Make HTML
	// bool vs *bool vs &bool
	if *arss == false && *ajson == false && *aatom == false {
		for o, _ := range keys {
			// keys is a sorted list of keys of data
			// o == 0,1,2 etc (index of element)
			// keys[o] == "2018-11-01 00:00:00 +0000 UTC" etc, each month
			fmt.Println("<h1>" + keys[o] + "</h1>")
			for k, _ := range data[keys[o]] {
				aHREF := "<a href='" + data[keys[o]][k] + "'>" + k + "</a><br>"
				// k == thread title
				// data[o][k] == thread full URL
				fmt.Print(aHREF)
			}
		}
	} else {

	// 04 Make RSS, ATOM or JSON
		// https://github.com/gorilla/feeds
		// http://www.gorillatoolkit.org/pkg/feeds
		now := time.Now()
		// &feeds.Feed{} == ??
		feed := &feeds.Feed{
		      Title:       "CEPH-users GW Threads",
		      Link:        &feeds.Link{Href: "http://lists.ceph.com/pipermail/ceph-users-ceph.com/"},
		      Description: "Threads from ceph-users CEPH mailing lists with GW in the title. Generated with https://github.com/martbhell/mailman-summarizer",
		      Created:     now,
		}

		for o, _ := range keys {
			// keys is a sorted list of keys of data
			// o == 0,1,2 etc (num of elements)
			// keys[o] == "2018-11-01 00:00:00 +0000 UTC" etc, each month
			// earlier we turned it into the above string, for feeds Created/Updated fields it needs to be time.Time again..
			dateofthreadsinTime, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", keys[o])
			// for the current month we set the PubDate field to when the script was run
			updatedfield := dateofthreadsinTime
			if dateofthreadsinTime.Month() == now.Month() && dateofthreadsinTime.Year() == now.Year() {
				updatedfield = now
			}
			// We make a string hex representation of updatedfield since epoch
			//   This is used as a GUID for the item in the RSS feed to make the w3c feed validator happy
			//   TODO: best would probably be to just link to the thread.html for that month in the mailing list web archive
			guid := strconv.FormatInt(updatedfield.Unix(), 16)

			// thelinks is some HTML with the threads we want to display
		        thelinks := ""
			for k, _ := range data[keys[o]] {
				thelinks = thelinks + "<a href='" + data[keys[o]][k] + "'>" + k + "</a><br>"
				// k == thread title
				// data[o][k] == thread full URL
			}
			// Do an Add() instead of defining Items every Month
			feed.Add(&feeds.Item{
                                    Title:       "CEPH GW Threads for " + keys[o],
				    Link:        &feeds.Link{Href: "https://storage.googleapis.com/ceph-rgw-users/feed.xml?guid=" + guid},
				    Id:		 "https://storage.googleapis.com/ceph-rgw-users/feed.xml?guid=" + guid,
                                    Description: thelinks,
				    Author:      &feeds.Author{Name: "ceph-users@lists.ceph.com (CEPH Users Mailing List)", Email: "http://lists.ceph.com/pipermail/ceph-users-ceph.com/"},
				    Updated:     updatedfield,
                                },

			)
		}

		atom, err := feed.ToAtom()
		if err != nil {
		    log.Fatal(err)
		}
		// aatom, arss and ajson are CLI arguments to the executable
		if *aatom == true { fmt.Println(atom) }

		rss, err := feed.ToRss()
		if err != nil {
		    log.Fatal(err)
		}
		if *arss == true { fmt.Println(rss) }

		json, err := feed.ToJSON()
		if err != nil {
		    log.Fatal(err)
		}
		if *ajson == true { fmt.Println(json) }

	}

}
