package hwwatcher

import "github.com/kelseyhightower/envconfig"

type Config struct {
	CPULoadSourcePath string `split_words:"true"`
}

func MustNewConfig() Config {
	var cfg Config
	envconfig.MustProcess("hw", &cfg)

	return cfg
}
