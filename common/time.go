package common

import (
	"time"
)

func NowInMilli() int64 {
	return time.Now().UnixMilli()
}

func ToSeconds(duration time.Duration) int {
	const nanosecondMultiplier = time.Nanosecond * time.Microsecond * time.Millisecond
	return int(duration / nanosecondMultiplier)
}
