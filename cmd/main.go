package main

import (
	"amazon-crawler/m/v2/internal/config"
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/sethvargo/go-envconfig"

	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
)

const (
	out_path        = "../out"
	image_out_dir   = "images"
	report_filename = "data.csv"
	target_base_uri = "https://www.amazon.de"
)

func main() {
	parentCtx := context.Background()
	var regexAsin = regexp.MustCompile(`^.*[dp\/](.*)\/ref.*`)
	var regexItemName = regexp.MustCompile(`^(.*)\/[dp\/].*`)

	var conf config.Config
	if err := envconfig.Process(parentCtx, &conf); err != nil {
		log.Fatal(err)
	}
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.NoFirstRun,

		chromedp.NoDefaultBrowserCheck,
		chromedp.WindowSize(1920, 1080),
	)
	if conf.Debug {
		// hint from https://devmarkpro.com/chromedp-get-started
		opts = append(
			chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", false),
			chromedp.NoFirstRun,
			chromedp.NoDefaultBrowserCheck,
			chromedp.WindowSize(1920, 1080),
		)

	}
	parentCtx, _ = chromedp.NewExecAllocator(parentCtx, opts...)
	chromeDpCtx, chromeDpCancel := chromedp.NewContext(parentCtx, chromedp.WithDebugf(log.Printf))
	defer chromeDpCancel()
	if err := chromedp.Run(chromeDpCtx, acceptCookies(target_base_uri)); err != nil {
		log.Fatal(err)
	}
	file, writer, err := prepareReportFile(out_path, report_filename)

	if err != nil {
		panic("Exit.")
	}

	defer file.Close()
	defer writer.Flush()

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
			s := regexAsin.ReplaceAllString(link, `$1`)
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

		itemName := regexItemName.ReplaceAllString(el, `$1`)
		fileName := fmt.Sprintf("%s/%s/%s-%s.png", out_path, image_out_dir, itemName, key)
		screenShotUri := fmt.Sprintf("%s%s", target_base_uri, el)
		sizeOptionValue := fmt.Sprintf("1,%s", key)
		selectedOption := ""
		sellerName := ""

		fmt.Printf("Storing screenshot to %s | from url: %s\n | and the size selector is %s\n", fileName, screenShotUri, sizeOptionValue)
		var imageBuf []byte
		tasks, result := screenshotDetailPages(screenShotUri, 90, &sizeOptionValue, &selectedOption, &sellerName, &imageBuf)
		fmt.Printf("Result is  %v", result)
		err := chromedp.Run(chromeDpCtx, tasks)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Selected t-shirt size %s", sizeOptionValue)
		if err := ioutil.WriteFile(fileName, imageBuf, 9544); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Collected", len(knownASINs), "ASINs")

}

func acceptCookies(url string) chromedp.Tasks {
	return chromedp.Tasks{
		network.SetExtraHTTPHeaders(network.Headers(map[string]interface{}{"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36"})),
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Click(`#sp-cc-accept`, chromedp.NodeVisible),
		chromedp.Sleep(time.Second * 2),
	}
}

func screenshotDetailPages(url string, quality int, sizeOptionValue *string, selectedSize, sellerName *string, res *[]byte) (chromedp.Tasks, interface{}) {
	var result interface{}
	tasks := chromedp.Tasks{
		chromedp.EmulateViewport(1920, 1080),
		network.SetExtraHTTPHeaders(network.Headers(map[string]interface{}{"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36", "Accept-Language": "de"})),
		chromedp.Navigate(url),
		chromedp.Sleep(2000 * time.Millisecond),
		chromedp.WaitVisible(`#navFooter`),

		chromedp.Click(`#native_dropdown_selected_size_name`),

		//chromedp.EvaluateAsDevTools(`selEl = document.getElementById("native_dropdown_selected_size_name"); selEl.options[1].selected = true; selEl.options[1].value;`, &result),
		chromedp.Sleep(2000 * time.Millisecond),

		chromedp.Value(`#native_dropdown_selected_size_name`, selectedSize, chromedp.ByID),
		//chromedp.WaitVisible(`#availability`, chromedp.ByID),
		//chromedp.Text(`div[tabular-attribute-name="Verkäufer"]`, sellerName, chromedp.ByQuery),
		chromedp.FullScreenshot(res, quality),
	}
	return tasks, result
}

func prepareReportFile(outPath, fileName string) (*os.File, *csv.Writer, error) {
	file, err := os.Create(fmt.Sprintf("%s/%s", outPath, fileName))

	if err != nil {
		log.Fatalf("Could not create file, err: %q", err)
		return nil, nil, err
	}

	writer := csv.NewWriter(file)
	writer.Comma = ';'
	writer.Write([]string{
		"ASIN",
		"URL",
	})
	return file, writer, nil
}
