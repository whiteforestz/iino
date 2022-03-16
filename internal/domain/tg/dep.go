package tg

import (
	"net/http"

	"github.com/whiteforestz/iino/internal/domain/sysmon"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type SysMonDomain interface {
	GetUsage() (*sysmon.Usage, error)
}
