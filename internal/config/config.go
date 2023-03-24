package config

type Config struct {
	Debug            bool   `env:"DEBUG,default=false"`
	OutPath          string `env:"OUT_PATH,default=../out"`
	ScreenshotSubDir string `env:"SCREENSHOT_SUB_PATH,default=screenshots"`
	ReportFileName   string `env:"REPORT_FILENAME,default=data.csv"`
	SearchBaseUrl    string `env:"SEARCH_BASE_URL,default=https://www.amazon.de"`
}
