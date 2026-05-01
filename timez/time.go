package timez

import (
	"sort"
	"time"
)

const YYYYMMDD = "2006-01-02"

func Min(times ...time.Time) time.Time {
	if len(times) == 0 {
		panic("datetime: no time supplied")
	}
	sort.Slice(times, func(i, j int) bool {
		return times[i].Before(times[j])
	})
	return times[0]
}

func Max(times ...time.Time) time.Time {
	if len(times) == 0 {
		panic("datetime: no time supplied")
	}
	sort.Slice(times, func(i, j int) bool {
		return times[i].After(times[j])
	})
	return times[0]
}

func ParseOrPanic(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

// Parse a date with format "2026-12-21"
func ParseYyyyMmDd(value string) time.Time {
	return ParseOrPanic(YYYYMMDD, value)
}
