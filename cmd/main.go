package main

import (
	"amazon-crawler/m/v2/internal/config"
	chromeDriver "amazon-crawler/m/v2/internal/pkg/adapter/chrome"
	"amazon-crawler/m/v2/internal/pkg/adapter/scraper"
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/sethvargo/go-envconfig"
)

const (
	out_path        = "../out"
	image_out_dir   = "images"
	report_filename = "data.csv"
	target_base_uri = "https://www.amazon.de"
)

func main() {
	parentCtx := context.Background()
	var regexItemName = regexp.MustCompile(`^(.*)\/[dp\/].*`)

	var conf config.Config
	if err := envconfig.Process(parentCtx, &conf); err != nil {
		log.Fatalf("Error loading config, err: %s", err)
	}

	chromeDriver := chromeDriver.New(parentCtx, &conf)
	defer chromeDriver.CancelContextFunc()

	file, err := os.Create(fmt.Sprintf("%s/%s", conf.OutPath, conf.ReportFileName))
	if err != nil {
		log.Fatalf("Could not create file, err: %q", err)
	}
	writer := csv.NewWriter(file)
	writer.Comma = ';'
	defer file.Close()
	defer writer.Flush()

	//fmt.Println("All known URLs:")
	//for _, url := range knownUrls {
	//	fmt.Println("\t", url)
	//}
	htmlScraper := scraper.New(parentCtx, &conf, file, writer)
	knownASINs, err := htmlScraper.GetSearchItemLinksAndAsin("normahl+shirt")
	if err != nil {
		log.Fatalf("Eror scrapeing our target - err: %q", err)
	}

	for asin, url := range knownASINs {
		fmt.Printf("ASIN: %s | url: %s\n", asin, url)

		itemName := regexItemName.ReplaceAllString(url, `$1`)
		fileName := fmt.Sprintf("%s/%s/%s-%s.png", out_path, image_out_dir, itemName, asin)
		var selectedOption, sellerName *string

		imageBuf, err := chromeDriver.ScreenshotDetailPages(url, 90, selectedOption, sellerName)
		if err != nil {
			log.Fatalf("Eror taein screenshot from target ASIN %s and URL %s- err: %q", asin, url, err)
		}

		if err := ioutil.WriteFile(fileName, *imageBuf, 9544); err != nil {
			log.Fatal(err)
		}
	}

}
