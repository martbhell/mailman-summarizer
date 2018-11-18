package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains("lists.ceph.com"),
	)

//	listoflinks := []

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

	// Callback for links on scraped pages
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// Extract the linked URL from the anchor tag
		link := e.Attr("href")
		linksplit := (strings.Split(link, "/"))
		lastlink := linksplit[0]
		if strings.ContainsAny(lastlink, "0123456789") {
			if strings.Contains(lastlink, ".html") {
				if strings.Contains(e.Text, "RGW") {
					fmt.Println(lastlink)
					fmt.Println(e.Text)
				}
			}
		}
			// Only if from 2018
			// https://gobyexample.com/if-else
			if strings.Contains(link, "2018") {
				// Only Thread (from list of Months page) and ceph-users (from thread page)
				// https://stackoverflow.com/questions/45266784/go-test-string-contains-substring
				if strings.Contains(e.Text, "[ceph-users]"); strings.Contains(e.Text, "[ Thread ]") {
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
}
