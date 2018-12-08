package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"flag" // for CLI parsing
	"log" // nice logger
	"strconv" // convert Int to String

	"github.com/PuerkitoBio/goquery" // for scraping
	"github.com/gocolly/colly" // for scraping
	"github.com/gorilla/feeds" // making RSS
)


func main() {

        // CLI parsing: https://gobyexample.com/command-line-flags
        arss := flag.Bool("rss", false, "Set if you want RSS output instead of HTML")
        ajson := flag.Bool("json", false, "Set if you want JSON output instead of HTML")
        aatom := flag.Bool("atom", false, "Set if you want Atom output instead of HTML")
	var topic string
	flag.StringVar(&topic, "topic", "GW", "a comma separated list of strings which the thread topic must contain")
        flag.Parse()

	topicsplit := (strings.Split(topic, ","))

	// order matters between these!
	// 01 first we scrape a website

	c := colly.NewCollector(
		colly.AllowedDomains("lists.ceph.com"),
	)
	// https://stackoverflow.com/questions/44305617/nested-maps-in-golang
	// in python this would look like data["November-2018"]["ceph-users title"] = "http://link.to.thread"
	data := make(map[string]map[string]string)

	// Callback for when a scraped page contains an article element
	c.OnHTML("article", func(e *colly.HTMLElement) {
		isEmojiPage := false

		// Extract meta tags from the document
		e.DOM.ParentsUntil("~").Find("meta").Each(func(_ int, s *goquery.Selection) {
			// Search for og:type meta tags
			if property, _ := s.Attr("property"); strings.EqualFold(property, "og:type") {
				content, _ := s.Attr("content")

				// Emoji pages have "article" as their og:type
				isEmojiPage = strings.EqualFold(content, "article")
			}
		})

		if isEmojiPage {
			// Find the emoji page title
			fmt.Println("Emoji: ", e.DOM.Find("h1").Text())
			// Grab all the text from the emoji's description
			fmt.Println("Description: ", e.DOM.Find(".description").Find("p").Text())
		}
	})

	parentthread := ""
	// Callback for links on scraped pages
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// Extract the linked URL from the anchor tag
		link := e.Attr("href")
		linksplit := (strings.Split(link, "/"))
		lastlink := linksplit[0]
		if strings.Contains(link, "thread.html") {
			// http://lists.ceph.com/pipermail/ceph-users-ceph.com/2018-November/thread.html
			// Here we store the parentthread in a variable. Later used to make the fullinktothethread
			// it should be something like "2018-November"
			parentthread = link
		}
		if strings.ContainsAny(lastlink, "0123456789") {
			if strings.Contains(lastlink, ".html") {
				// topic is an argument, defaults to "GW"
				for a, _ := range topicsplit {
					// note how when we range, "a" is the number of the element, so 0, 1, etc
					if strings.Contains(e.Text, topicsplit[a]) {
						// parentthread is at this point in time the full URL to the thread.html for this month
						// we split out yearmonth so we get: "2018-November"
					        yearmonth := (strings.Split(parentthread, "/")[0])
						// we get two values from time.Parse(). The _ is where we put the second value
						// parsedmonth should look like: 2018-11-01 00:00:00 +0000 UTC
						parsedmonth, _ := time.Parse("2006-January", yearmonth)

						fulllinktothethread := "http://lists.ceph.com/pipermail/ceph-users-ceph.com/" + yearmonth + "/" + lastlink
						// key in the map of maps is the string of parsedmonth
						datakey := parsedmonth.String()
						// maps has to be fully initialized or we get a runtime error - if it's nil and if so initialize it
						// only save the link if it's empty (we should link to the first e-mail in the thread)
						if data[datakey] == nil { data[datakey] = map[string]string{} }
						if data[datakey][e.Text] == "" {
							data[datakey][e.Text] = fulllinktothethread
						}
					}
				}
			}
		}
			// Because this script was made in 2018 wanted to also have 2017 in the entries.
			//  But don't want to delete entries when a year changes.
			//  so we loop over the years and only visit the last x years starting from 2017
			// golang does not have a range() like php/python apparently https://stackoverflow.com/questions/39868029/how-to-generate-a-sequence-of-numbers-in-golang
			thisyear := time.Now().Year()
			// apparently (above stackoverflow) this way is only good for a small range.. is this small?
			listofyears := []int{2017, thisyear}
			for q, _ := range listofyears {
				yearstring := strconv.Itoa(listofyears[q])
				if strings.Contains(link, yearstring) {
					// Only Thread (from list of Months page): http://lists.ceph.com/pipermail/ceph-users-ceph.com/2018-November/thread.html
					// https://stackoverflow.com/questions/45266784/go-test-string-contains-substring
					if strings.Contains(e.Text, "[ Thread ]") {
						c.Visit(e.Request.AbsoluteURL(link))
					}
				}
			}
	})

	c.OnHTML("title", func(e *colly.HTMLElement) {
	    // fmt.Println(e.Text)
	    // used to print the HTML <title>
	})

	// This piece adds delay so we are being nice on the Internet
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 1 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
//		fmt.Println("Visiting", r.URL.String())
// from demo, used to print which URL our scraper visits
	})

	// TODO: make an argument
	c.Visit("http://lists.ceph.com/pipermail/ceph-users-ceph.com/")


	//////////////// 
	// TODO: split out into another class/method/function/file?
	// 02 Loop over the data 

        // Data structure:
	// data = { "2018-November": { "thread1": "link1", "thread2": "link2", .. }, "2018-October": { "thread3": "link3", .. }, .. }
//	fmt.Println(data)


	// https://stackoverflow.com/questions/1841443/iterating-over-all-the-keys-of-a-map
	// https://stackoverflow.com/questions/23330781/sort-go-map-values-by-keys
	// First iterate over the keys (2018-November) in the data map, put them in a list and then sort them
	keys := make([]string, 0, len(data))
	for l, _ := range data {
		keys = append(keys, l)
	}
	sort.Strings(keys)

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
