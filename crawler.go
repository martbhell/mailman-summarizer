package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

func main() {

	// order matters between these!

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
			// http://lists.ceph.com/pipermail/ceph-users-ceph.com/2018-July/0123456.html
			parentthread = link
		}
		if strings.ContainsAny(lastlink, "0123456789") {
			if strings.Contains(lastlink, ".html") {
				if strings.Contains(e.Text, "RGW") {
					// only save last link for a thread
					// TODO: save all?
					// parentthread is at this point in time the full URL to the thread.html for this month
					parentthreadelementzero := strings.Split(parentthread, "/")[0]
					fulllinktothethread := "http://lists.ceph.com/pipermail/ceph-users-ceph.com/" + parentthreadelementzero + "/" + lastlink
					// with a \n at the end to make the output a bit more readable
					// fulllinktothethread := "http://lists.ceph.com/pipermail/ceph-users-ceph.com/" + parentthreadelementzero + "/" + lastlink + "\n"
					// maps has to be fully initialized or we get a runtime error - if it's nil and if so initialize it
					if data[parentthreadelementzero] == nil { data[parentthreadelementzero] = map[string]string{} }
					data[parentthreadelementzero][e.Text] = fulllinktothethread
				}
			}
		}
			// TODO: Only this and last year?
			if strings.Contains(link, "2018") {
				// Only Thread (from list of Months page)
				// https://stackoverflow.com/questions/45266784/go-test-string-contains-substring
				if strings.Contains(e.Text, "[ Thread ]") {
					c.Visit(e.Request.AbsoluteURL(link))
				}
			}
	})

	c.OnHTML("title", func(e *colly.HTMLElement) {
	    fmt.Println(e.Text)
	})

	// This piece adds dela so we are being nice on the Internet
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 1 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit("http://lists.ceph.com/pipermail/ceph-users-ceph.com/")

        // Data structure:
	// data = { "2018-November": { "thread1": "link1", "thread2": "link2", .. }, "2018-October": { "thread3": "link3", .. }, .. }
	fmt.Println(data)
	for o, _ := range data {
		// first key level is YYYY-month
		// o == 2018-November
		fmt.Println(o)
		for k, _ := range data[o] {
			// k == thread title
			fmt.Println(k)
			// data[o][k] == thread full URL
			fmt.Println(data[o][k])
		}
	}

}
