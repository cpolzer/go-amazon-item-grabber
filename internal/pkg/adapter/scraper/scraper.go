package scraper

import (
	"amazon-crawler/m/v2/internal/config"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

type Scraper struct {
	asinRegex  *regexp.Regexp
	csvWriter  *csv.Writer
	reportFile *os.File
	conf       *config.Config
	collector  *colly.Collector
}

func New(ctx context.Context, conf *config.Config, reportFile *os.File, csvWriter *csv.Writer) *Scraper {
	return &Scraper{
		asinRegex:  regexp.MustCompile(`^.*[dp\/](.*)\/ref.*`),
		csvWriter:  csvWriter,
		reportFile: reportFile,
		conf:       conf,
		collector:  colly.NewCollector(colly.AllowedDomains(removeHttpPrefixFromUrl(conf.SearchBaseUrl))),
	}
}

func removeHttpPrefixFromUrl(url string) string {
	var regexUrl = regexp.MustCompile(`^https:\/\/|http:\/\/(.*)`)
	return regexUrl.ReplaceAllString(url, `$1`)
}

func (scr Scraper) GetSearchItemLinksAndAsin(searchTerm string) (map[string]string, error) {
	knownASINs := map[string]string{}
	prepareReportFile(scr.csvWriter)
	// Create a callback on the XPath query searching for the URLs
	scr.collector.OnHTML("a.a-link-normal", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		if strings.Contains(link, "/dp/") {
			s := scr.asinRegex.ReplaceAllString(link, `$1`)
			knownASINs[s] = link
			/// klick tshirt size if available, check seller name‚
		}

	})
	// Start the collector
	err := scr.collector.Visit(fmt.Sprintf("%s/s?k=%s", scr.conf.SearchBaseUrl, searchTerm))
	if err != nil {
		return nil, err
	}
	for key, el := range knownASINs {
		scr.csvWriter.Write([]string{
			key,
			el,
		})
	}

	return knownASINs, nil
}

func prepareReportFile(writer *csv.Writer) {
	writer.Write([]string{
		"ASIN",
		"URL",
	})
}
