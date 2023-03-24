package chromeDp

import (
	"amazon-crawler/m/v2/internal/config"
	"amazon-crawler/m/v2/internal/pkg/adapter/chromeDriver"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type driver struct {
	config        *config.Config
	ChromeContext context.Context
}

func New(parentCtx context.Context, conf *config.Config) (chromeDriver.ChromeDriver, context.CancelFunc) {
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
	return &driver{
		config:        conf,
		ChromeContext: chromeDpCtx,
	}, chromeDpCancel
}

func (c driver) AcceptCookies(url string) error {
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

func (c driver) ScreenshotPage(uri string, quality int) (*[]byte, error) {
	imageBuf := []byte{}
	uri = fmt.Sprintf("%s%s&psc=1", c.config.SearchBaseUrl, uri)
	tasks := chromedp.Tasks{
		chromedp.EmulateViewport(1920, 1080),
		network.SetExtraHTTPHeaders(map[string]interface{}{"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36", "Accept-Language": "de"}),
		chromedp.Navigate(uri),
		chromedp.Sleep(2000 * time.Millisecond),
		chromedp.WaitVisible(`#navFooter`),
		chromedp.Click(`#sp-cc-accept`, chromedp.NodeVisible),
		chromedp.Sleep(time.Second * 2),
		chromedp.WaitVisible(`#availability`, chromedp.ByID),
		chromedp.FullScreenshot(&imageBuf, quality),
	}

	err := chromedp.Run(c.ChromeContext, tasks)
	if err != nil {
		return nil, err
	}
	return &imageBuf, nil
}
