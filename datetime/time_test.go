package datetime

import (
	"time"

	"testing"
        "github.com/stretchr/testify/assert"
)

var (
	time0 = time.Time{}
	time1 = time.Now()
	time2 = time1.Add(time.Second * 10)
	time3 = time2.Add(time.Second * 10)
)

func TestMin(t *testing.T) {
	// empty case should panic
	assert.Panics(t, func(){ Min() }, "empty case should panic")

        cases := []struct {
                in []time.Time
		want time.Time
        }{
                {[]time.Time{time1}, time1},
                {[]time.Time{time1, time2, time3}, time1},
                {[]time.Time{time0, time2}, time0},
        }
        for i, c := range cases {
                got := Min(c.in...)
                assert.Equal(t, c.want, got, "case #%d should be equal", i)
        }
}

func TestMax(t *testing.T) {
	// empty case should panic
	assert.Panics(t, func(){ Max() }, "empty case should panic")

        cases := []struct {
                in []time.Time
		want time.Time
        }{
                {[]time.Time{time1}, time1},
                {[]time.Time{time1, time2, time3}, time3},
                {[]time.Time{time0, time2}, time2},
        }
        for i, c := range cases {
                got := Max(c.in...)
                assert.Equal(t, c.want, got, "case #%d should be equal", i)
        }
}

