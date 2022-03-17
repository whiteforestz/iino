package persistor

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	RootPath string `split_words:"true"`
}

func MustNewConfig() Config {
	var cfg Config
	envconfig.MustProcess("persistor", &cfg)

	return cfg
}
