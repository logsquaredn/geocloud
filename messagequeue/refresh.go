package messagequeue

import "time"

const tm5 = "5m"

var queueRefreshInterval time.Duration

func init() {
	queueRefreshInterval, _ = time.ParseDuration(tm5)
}
