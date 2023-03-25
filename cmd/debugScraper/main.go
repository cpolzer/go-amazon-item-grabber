package main

import (
	"amazon-crawler/m/v2/internal/config"
	"amazon-crawler/m/v2/internal/pkg/adapter/scraper"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/sethvargo/go-envconfig"
)

func main() {
	parentCtx := context.Background()

	var conf config.Config
	if err := envconfig.Process(parentCtx, &conf); err != nil {
		log.Fatalf("Error loading config, err: %s", err)
	}
	err := setUpDirectories(conf)
	if err != nil {
		log.Fatalf("Error setting up required output directories '%s/%s'. Error: %s", conf.OutPath, conf.ScreenshotSubDir, err)
	}

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
	fmt.Printf("Found %d items: %v", len(knownASINs), knownASINs)

}

func setUpDirectories(conf config.Config) error {
	err := os.MkdirAll(fmt.Sprintf("%s/%s", conf.OutPath, conf.ScreenshotSubDir), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
