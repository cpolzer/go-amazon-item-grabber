package config

type Config struct {
	Debug bool `env:"DEBUG,default=false"`
}
