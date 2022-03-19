package tglistener

import (
	"fmt"
	"time"
)

const (
	defaultTimeFormat = "2006-01-02 15:04"

	activityStatusOffline         = "offline"
	activityStatusOnline          = "online"
	activityStatusOnlineThreshold = 2 * time.Minute

	memorySizeThreshold = 1000
)

var (
	memoryUnitSlugAccessor = map[int64]string{
		0: "B",
		1: "KiB",
		2: "MiB",
		3: "GiB",
		4: "TiB",
		5: "PiB",
		6: "EiB",
		7: "ZiB",
		8: "YiB",
	}
)

func formatActivityStatus(nowUnix, latestHandshakeUnix int64) string {
	if (nowUnix - latestHandshakeUnix) < int64(activityStatusOnlineThreshold.Seconds()) {
		return activityStatusOnline
	}

	return activityStatusOffline
}

func formatLatestActivity(latestHandshakeUnix int64) string {
	return time.Unix(latestHandshakeUnix, 0).Format(defaultTimeFormat)
}

func formatMemorySize(bytes int64) string {
	var (
		order int64
		n     = float64(bytes)
	)

	for n > memorySizeThreshold {
		n /= 1024
		order++
	}

	slug, found := memoryUnitSlugAccessor[order]
	if !found {
		return "unknown 🗿"
	}

	return fmt.Sprintf("%.2f %s", n, slug)
}
