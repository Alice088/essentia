package time

import "time"

func Seconds(sec int) time.Duration {
	return time.Duration(sec) * time.Second
}
