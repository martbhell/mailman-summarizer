package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"flag" // for CLI parsing
	"log" // nice logger

	"github.com/PuerkitoBio/goquery" // for scraping
	"github.com/gocolly/colly" // for scraping
	"github.com/gorilla/feeds" // making RSS
)

func main() {

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
				if strings.Contains(e.Text, "RGW") {
					// only save last link for a thread
					// TODO: save all?
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
					if data[datakey] == nil { data[datakey] = map[string]string{} }
					data[datakey][e.Text] = fulllinktothethread
				}
			}
		}
			// TODO: Only this and last year?
			if strings.Contains(link, "2018") {
				// Only Thread (from list of Months page): http://lists.ceph.com/pipermail/ceph-users-ceph.com/2018-November/thread.html
				// https://stackoverflow.com/questions/45266784/go-test-string-contains-substring
				if strings.Contains(e.Text, "[ Thread ]") {
					c.Visit(e.Request.AbsoluteURL(link))
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

	c.Visit("http://lists.ceph.com/pipermail/ceph-users-ceph.com/")


	//////////////// 
	// TODO: split out into another class/method/function?
	// 02 Loop over the data 

        // Data structure:
	// data = { "2018-November": { "thread1": "link1", "thread2": "link2", .. }, "2018-October": { "thread3": "link3", .. }, .. }
//	fmt.Println(data)


	// https://stackoverflow.com/questions/1841443/iterating-over-all-the-keys-of-a-map
	// https://stackoverflow.com/questions/23330781/sort-go-map-values-by-keys
	// First iterate over keys and put them in a list and then sort them
	keys := make([]string, 0, len(data))
	for l, _ := range data {
		keys = append(keys, l)
	}
	sort.Strings(keys)
	// debug:
	// 	fmt.Println(keys)
	// 	for n, _ := range keys { fmt.Println(data[keys[n]]) }
	// Now we have a sorted list called keys. It's sorted on the Thread Names. Would be nicer with sorted on the URL and the date..
	// data structure:
	// data = { 2018-November: { thread1: link1, thread2: link2, .. }, 2018-October: { thread3: link3, .. }, .. }

	// CLI parsing: https://gobyexample.com/command-line-flags
	rss := flag.Bool("rss", false, "Set if you want RSS output instead of HTML")
	flag.Parse()

	// 03 Make HTML
	// bool vs *bool
	if *rss == false {
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
	}

	// 04 Make RSS
	if *rss == true {
		// https://github.com/gorilla/feeds
		// http://www.gorillatoolkit.org/pkg/feeds
		now := time.Now()
		// &feeds.Feed{} == ??
		feed := &feeds.Feed{
		      Title:       "mailman-summarizer",
		      Link:        &feeds.Link{Href: "https://guldmyr.com/blog"},
		      Description: "discussion about tech",
		      Author:      &feeds.Author{Name: "Johan Guldmyr", Email: "martbhell+mailman@gmail.com"},
		      Created:     now,
		}

		for o, _ := range keys {
			fmt.Println(keys[o])
			fmt.Println(feed)
			// keys is a sorted list of keys of data
			// o == 0,1,2 etc (num of elements)
			// keys[o] == "2018-11-01 00:00:00 +0000 UTC" etc, each month

		        thelinks := ""
			for k, _ := range data[keys[o]] {
				thelinks = thelinks + "<a href='" + data[keys[o]][k] + "'>" + k + "</a><br>"
				// k == thread title
				// data[o][k] == thread full URL
			}
			// Do an Add() instead of defining Items every Month
			feed.Add(&feeds.Item{

				// TODO: Created/Updated could be set to 1st of each month for previous months
				//  	and time.Now() for current month. Maybe this would update the RSS feed?
                                    Title:       "CEPH Threads for " + keys[o],
                                    Link:        &feeds.Link{Href: "https://guldmyr.com/"},
                                    Description: thelinks,
				    Author:      &feeds.Author{Name: "CEPH Community", Email: "http://lists.ceph.com/pipermail/ceph-users-ceph.com/"},
                                    Created:     now,
                                },

			)
		}

		atom, err := feed.ToAtom()
		if err != nil {
		    log.Fatal(err)
		}

		rss, err := feed.ToRss()
		if err != nil {
		    log.Fatal(err)
		}

		json, err := feed.ToJSON()
		if err != nil {
		    log.Fatal(err)
		}

		fmt.Println(atom, "\n", rss, "\n", json)
	}

}
