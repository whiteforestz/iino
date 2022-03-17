package wgwatcher

import (
	"regexp"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Cmd         string   `split_words:"true"`
	CmdArgs     []string `split_words:"true"`
	ConfDirPath string   `split_words:"true"`
	ConfPattern string   `split_words:"true"`

	ConfPatternRe *regexp.Regexp
}

func MustNewConfig() Config {
	var cfg Config
	envconfig.MustProcess("wg", &cfg)

	cfg.ConfPatternRe = regexp.MustCompile(cfg.ConfPattern)

	return cfg
}
