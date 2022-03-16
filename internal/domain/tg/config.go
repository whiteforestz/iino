package tg

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	APIToken string `split_words:"true"`
	AdminID  int64  `split_words:"true"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("tg", &cfg); err != nil {
		return nil, fmt.Errorf("can't create cfg: %w", err)
	}

	return &cfg, nil
}
