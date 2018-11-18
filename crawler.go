package main

import (
	"fmt"
	"strings"
	"time"
	"sort"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

func main() {

	// order matters between these!

	c := colly.NewCollector(
		colly.AllowedDomains("lists.ceph.com"),
	)

	m := make(map[string]string)

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
			// Here we store the parentthread in a variable. Later used to make the fullinktothethread
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
					m[e.Text] = fulllinktothethread
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

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 1 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit("http://lists.ceph.com/pipermail/ceph-users-ceph.com/")

	// At this point we have a map (m) of the good stuff. Now to make a summary?
	// fmt.Println(m)
	// https://stackoverflow.com/questions/1841443/iterating-over-all-the-keys-of-a-map
	// https://stackoverflow.com/questions/23330781/sort-go-map-values-by-keys
	// First iterate over keys and put them in a list and then sort them
	keys := make([]string, 0, len(m))
	for l, _ := range m {
		keys = append(keys, l)
	}
	sort.Strings(keys)
	// Now we have a sorted list called keys. It's sorted on the Thread Names. Would be nicer with sorted on the URL and the date..
	// fmt.Println(keys)
	// Could this data structure be better perhaps?
	// data = { 2018-November: { thread1: link1, thread2: link2, .. }, 2018-October: { thread3: link3, .. }, .. }
	for k, _ := range m {
		aHREF := "<a href='" + m[k] + "'>"
		fmt.Println(aHREF + k + "</a><br>")
	}

}
