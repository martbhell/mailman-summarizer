package main

import (
	"sort"
	"strings" // for doing strings.Contains()
	"time"
	"flag" // for CLI parsing
	"strconv" // convert Int to String

	"github.com/gocolly/colly" // for scraping
)


func main() {

	// 1. parses a website 2. at the end calls makeRSS() which prints stuff to stdout

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
					// TODO: case insensitive comparing: https://stackoverflow.com/questions/24836044/case-insensitive-string-search-in-golang
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
		// fmt.Println("Visiting", r.URL.String())
		// from demo, used to print which URL our scraper visits. Leaving as a hint for what one could do.
	})

	// TODO: make an argument
	c.Visit("http://lists.ceph.com/pipermail/ceph-users-ceph.com/")


	// 02 Loop over the data 

        // Data structure:
	// data = { "2018-November": { "thread1": "link1", "thread2": "link2", .. }, "2018-October": { "thread3": "link3", .. }, .. }

	// https://stackoverflow.com/questions/1841443/iterating-over-all-the-keys-of-a-map
	// https://stackoverflow.com/questions/23330781/sort-go-map-values-by-keys
	// First iterate over the keys (2018-November) in the data map, put them in a list and then sort them
	keys := make([]string, 0, len(data))
	for l, _ := range data {
		keys = append(keys, l)
	}
	sort.Strings(keys)

	makeRSS(keys, data, topic, arss, ajson, aatom)

}
