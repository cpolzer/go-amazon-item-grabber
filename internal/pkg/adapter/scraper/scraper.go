package scraper

import (
	"amazon-crawler/m/v2/internal/config"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

type Scraper struct {
	asinRegex  *regexp.Regexp
	csvWriter  *csv.Writer
	reportFile *os.File
	conf       *config.Config
	collector  *colly.Collector
}

func New(_ context.Context, conf *config.Config, reportFile *os.File, csvWriter *csv.Writer) *Scraper {
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

	scr.collector.OnHTML(`div[data-component-type="s-search-result"]`, func(e *colly.HTMLElement) {
		asin := e.Attr("data-asin")
		fmt.Printf("Found asin! %s\n", asin)

		e.ForEach("div.sg-col-inner > div.s-widget-container > div.s-card-container > div.a-section > div.a-section", func(_ int, el *colly.HTMLElement) {

			domItems := el.DOM.Children()

			selection := domItems.Find("div.a-row > span.a-color-base.a-text-bold")
			if selection.Nodes != nil && len(selection.Nodes) > 0 {
				html, err := goquery.OuterHtml(selection)
				if err != nil {
					fmt.Printf("Error grabbing outer html. %v\n", err)
				}
				if strings.Contains(html, "Amazon Merch on Demand") {
					selection = domItems.Find("a.a-link-normal")
					if selection.Nodes != nil && len(selection.Nodes) > 0 {
						for idx, node := range selection.Nodes {
							element := colly.NewHTMLElementFromSelectionNode(el.Response, selection, node, idx)
							link := element.Attr("href")
							if strings.Contains(link, "/dp/") {
								knownASINs[asin] = link
							}
						}
					}
				}
			}

		})
	})

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
