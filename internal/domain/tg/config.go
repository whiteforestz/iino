package tg

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	APIToken string `split_words:"true"`
	AdminID  int64  `split_words:"true"`
}

func MustNewConfig() Config {
	var cfg Config
	envconfig.MustProcess("tg", &cfg)

	return cfg
}
