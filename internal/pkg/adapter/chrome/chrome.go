package chromeDriver

import (
	"amazon-crawler/m/v2/internal/config"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type ChromeDriver struct {
	config            *config.Config
	ChromeContext     context.Context
	CancelContextFunc context.CancelFunc
}

func New(parentCtx context.Context, conf *config.Config) *ChromeDriver {
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
	return &ChromeDriver{
		config:            conf,
		ChromeContext:     chromeDpCtx,
		CancelContextFunc: chromeDpCancel,
	}
}

func (c ChromeDriver) AcceptCookies(url string) error {
	tasks := chromedp.Tasks{
		network.SetExtraHTTPHeaders(network.Headers(map[string]interface{}{"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36"})),
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Click(`#sp-cc-accept`, chromedp.NodeVisible),
		chromedp.Sleep(time.Second * 2),
	}

	err := chromedp.Run(c.ChromeContext, tasks)
	if err != nil {
		return err
	}
	return nil
}

func (c ChromeDriver) ScreenshotDetailPages(url string, quality int, selectedSize, sellerName *string) (*[]byte, error) {
	var imageBuf *[]byte

	tasks := chromedp.Tasks{
		chromedp.EmulateViewport(1920, 1080),
		network.SetExtraHTTPHeaders(network.Headers(map[string]interface{}{"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36", "Accept-Language": "de"})),
		chromedp.Navigate(fmt.Sprintf("%s%s", c.config.SearchBaseUrl, url)),
		chromedp.Sleep(2000 * time.Millisecond),
		chromedp.WaitVisible(`#navFooter`),

		chromedp.Click(`#native_dropdown_selected_size_name`),

		//chromedp.EvaluateAsDevTools(`selEl = document.getElementById("native_dropdown_selected_size_name"); selEl.options[1].selected = true; selEl.options[1].value;`, &result),
		chromedp.Sleep(2000 * time.Millisecond),

		chromedp.Value(`#native_dropdown_selected_size_name`, selectedSize, chromedp.ByID),
		//chromedp.WaitVisible(`#availability`, chromedp.ByID),
		//chromedp.Text(`div[tabular-attribute-name="Verkäufer"]`, sellerName, chromedp.ByQuery),
		chromedp.FullScreenshot(imageBuf, quality),
	}

	err := chromedp.Run(c.ChromeContext, tasks)
	if err != nil {
		return nil, err
	}
	return imageBuf, nil
}
