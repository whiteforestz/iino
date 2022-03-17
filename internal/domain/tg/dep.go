package tg

import (
	"net/http"

	"github.com/whiteforestz/iino/internal/domain/hwwatcher"
	"github.com/whiteforestz/iino/internal/domain/wgwatcher"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HWWatcherDomain interface {
	GetUsage() (*hwwatcher.Usage, error)
}

type WGWatcherDomain interface {
	GetUsage() (*wgwatcher.Usage, error)
}
