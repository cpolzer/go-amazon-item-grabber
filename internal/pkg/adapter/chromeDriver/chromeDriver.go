package chromeDriver

type ChromeDriver interface {
	AcceptCookies(url string) error
	ScreenshotPage(uri string, quality int) (*[]byte, error)
}
