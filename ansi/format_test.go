package ansi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnformat(t *testing.T) {
	s := ""
	assert.Equal(t, s, Unformat(""))

	s0 := "foo"
	assert.Equal(t, s0, Unformat(s0))

	s1 := string(Black) + s0
	assert.Equal(t, s0, Unformat(s1))

	s1 = s0 + string(Black)
	assert.Equal(t, s0, Unformat(s1))

	s1 = string(Reset) + s0
	assert.Equal(t, s0, Unformat(s1))

	s1 = s0 + string(Reset)
	assert.Equal(t, s0, Unformat(s1))

	s1 = string(Black) + s0 + string(Reset)
	assert.Equal(t, s0, Unformat(s1))

	s1 = string(Reset) + s0 + string(Black)
	assert.Equal(t, s0, Unformat(s1))

	s1 = string(HilightBlack) + s0 + string(Reset)
	assert.Equal(t, s0, Unformat(s1))
}

func TestFormat(t *testing.T) {
	s := "foo"
	expectedRaw := s
	f := Format(None, s)

	assert.Equal(t, expectedRaw, f.Raw())
	assert.Equal(t, expectedRaw, f.Format())
	assert.Equal(t, expectedRaw, f.String())

	f = Format(Black, s)
	expectedFormatted := string(Black) + expectedRaw + string(Reset)
	assert.Equal(t, expectedRaw, f.Raw())
	assert.Equal(t, expectedFormatted, f.Format())
	assert.Equal(t, expectedFormatted, f.String())

	f.Disable()
	assert.Equal(t, expectedRaw, f.Raw())
	assert.Equal(t, expectedFormatted, f.Format())
	assert.Equal(t, expectedRaw, f.String())

	f.Enable()
	assert.Equal(t, expectedRaw, f.Raw())
	assert.Equal(t, expectedFormatted, f.Format())
	assert.Equal(t, expectedFormatted, f.String())

	s1 := "bar"
	s2 := string(Black) + s1 + string(Reset)
	f = Format(Blue, s2)

	f.Squash(false)
	expectedRaw = s2
	expectedFormatted = string(Blue) + expectedRaw + string(Reset)
	assert.Equal(t, expectedRaw, f.Raw())
	assert.Equal(t, expectedFormatted, f.Format())
	assert.Equal(t, expectedFormatted, f.String())

	f.Squash(true)
	expectedRaw = s1
	expectedFormatted = string(Blue) + expectedRaw + string(Reset)
	assert.Equal(t, expectedRaw, f.Raw())
	assert.Equal(t, expectedFormatted, f.Format())
	assert.Equal(t, expectedFormatted, f.String())

}
