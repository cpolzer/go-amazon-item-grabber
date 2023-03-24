package main

import (
	"amazon-crawler/m/v2/internal/config"
	"amazon-crawler/m/v2/internal/pkg/adapter/chromeDriver/chromeDp"
	"amazon-crawler/m/v2/internal/pkg/adapter/scraper"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/sethvargo/go-envconfig"
)

func main() {
	parentCtx := context.Background()
	var regexItemName = regexp.MustCompile(`^(.*)\/[dp\/].*`)
	var conf config.Config
	if err := envconfig.Process(parentCtx, &conf); err != nil {
		log.Fatalf("Error loading config, err: %s", err)
	}
	err := setUpDirectories(conf)
	if err != nil {
		log.Fatalf("Error setting up required output directories '%s/%s'. Error: %s", conf.OutPath, conf.ScreenshotSubDir, err)
	}
	chromeDriver, chromeDriverCancelFunc := chromeDp.New(parentCtx, &conf)
	defer chromeDriverCancelFunc()

	file, err := os.OpenFile(fmt.Sprintf("%s/%s", conf.OutPath, conf.ReportFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Could not create file, err: %q", err)
	}
	writer := csv.NewWriter(file)
	writer.Comma = ';'
	defer writer.Flush()

	htmlScraper := scraper.New(parentCtx, &conf, file, writer)
	knownASINs, err := htmlScraper.GetSearchItemLinksAndAsin("normahl+shirt")
	if err != nil {
		log.Fatalf("Error scrapeing our target - err: %q", err)
	}

	err = chromeDriver.AcceptCookies(conf.SearchBaseUrl)
	if err != nil {
		log.Fatalf("Error accepting the cookies - err: %q", err)
	}

	for asin, url := range knownASINs {
		fmt.Printf("ASIN: %s | url: %s\n", asin, url)

		itemName := regexItemName.ReplaceAllString(url, `$1`)
		fileName := fmt.Sprintf("%s/%s/%s-%s.png", conf.OutPath, conf.ScreenshotSubDir, itemName, asin)

		if err != nil {
			log.Fatalf("Error taking screenshot from target ASIN %s and URL %s- err: %q", asin, url, err)
		}

		imageBuf, err := chromeDriver.ScreenshotPage(url, 90)
		if err != nil {
			log.Fatalf("Error taking screenshot from target ASIN %s and URL %s- err: %q", asin, url, err)
		}
		_, err = os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Could not create image file, err: %q", err)
		}
		if err := os.WriteFile(fileName, *imageBuf, 9544); err != nil {
			log.Fatal(err)
		}
	}
	handleGracefulShutdown(chromeDriverCancelFunc)
}

func handleGracefulShutdown(cancelChromeDriver context.CancelFunc) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel

	cancelChromeDriver()
}

func setUpDirectories(conf config.Config) error {
	err := os.MkdirAll(fmt.Sprintf("%s/%s", conf.OutPath, conf.ScreenshotSubDir), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
