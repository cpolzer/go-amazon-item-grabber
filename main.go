package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

func main() {
	var re = regexp.MustCompile(`^.*[dp\/](.*)\/ref.*`)

	fName := "data.csv"
	file, err := os.Create(fName)

	if err != nil {
		log.Fatalf("Could not create file, err: %q", err)
		return
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	writer.Comma = ';'
	defer writer.Flush()
	writer.Write([]string{
		"ASIN",
		"URL",
	})
	// Array containing all the known URLs in a sitemap
	knownUrls := []string{}
	knownASINs := map[string]string{}
	// Create a Collector specifically for Shopify
	c := colly.NewCollector(colly.AllowedDomains("www.amazon.de"))

	// Create a callback on the XPath query searching for the URLs
	c.OnHTML("a.a-link-normal", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		if strings.Contains(link, "/dp/") {
			knownUrls = append(knownUrls, link)
			s := re.ReplaceAllString(link, `$1`)
			knownASINs[s] = link
			/// klick tshirt size if available, check seller name‚
		}

	})
	// Start the collector
	c.Visit("https://www.amazon.de/s?k=normahl+shirt")

	//fmt.Println("All known URLs:")
	//for _, url := range knownUrls {
	//	fmt.Println("\t", url)
	//}

	fmt.Println("All known ASINs:")
	for key, el := range knownASINs {
		fmt.Printf("ASIN: %s | url: %s\n", key, el)
		writer.Write([]string{
			key,
			el,
		})
	}
	fmt.Println("Collected", len(knownASINs), "ASINs")
}
