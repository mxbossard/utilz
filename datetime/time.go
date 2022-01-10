package datetime

import (
	"time"
	"sort"
)

func Min(times ...time.Time) time.Time {
	sort.Slice(times, func(i, j int) bool {
		return times[i].Before(times[j])
	})
	return times[0]
}

func Max(times ...time.Time) time.Time {
	sort.Slice(times, func(i, j int) bool {
		return times[i].After(times[j])
	})
	return times[0]
}
