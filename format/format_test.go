package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"mby.fr/utils/ansi"
)

func TestUnformat(t *testing.T) {
	s := ""
	assert.Equal(t, s, Unformat(""))

	s0 := "foo"
	assert.Equal(t, s0, Unformat(s0))

	s1 := string(ansi.Black) + s0
	assert.Equal(t, s0, Unformat(s1))

	s1 = s0 + string(ansi.Black)
	assert.Equal(t, s0, Unformat(s1))

	s1 = string(ansi.Reset) + s0
	assert.Equal(t, s0, Unformat(s1))

	s1 = s0 + string(ansi.Reset)
	assert.Equal(t, s0, Unformat(s1))

	s1 = string(ansi.Black) + s0 + string(ansi.Reset)
	assert.Equal(t, s0, Unformat(s1))

	s1 = string(ansi.Reset) + s0 + string(ansi.Black)
	assert.Equal(t, s0, Unformat(s1))

	s1 = string(ansi.HilightBlack) + s0 + string(ansi.Reset)
	assert.Equal(t, s0, Unformat(s1))
}

func TestFormat(t *testing.T) {
	s := "foo"
	expectedRaw := s
	f := New(ansi.None, s)

	assert.Equal(t, expectedRaw, f.Raw())
	assert.Equal(t, expectedRaw, f.Format())
	assert.Equal(t, expectedRaw, f.String())

	f = New(ansi.Black, s)
	expectedFormatted := string(ansi.Black) + expectedRaw + string(ansi.Reset)
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
	s2 := string(ansi.Black) + s1 + string(ansi.Reset)
	f = New(ansi.Blue, s2)

	f.Squash(false)
	expectedRaw = s2
	expectedFormatted = string(ansi.Blue) + expectedRaw + string(ansi.Reset)
	assert.Equal(t, expectedRaw, f.Raw())
	assert.Equal(t, expectedFormatted, f.Format())
	assert.Equal(t, expectedFormatted, f.String())

	f.Squash(true)
	expectedRaw = s1
	expectedFormatted = string(ansi.Blue) + expectedRaw + string(ansi.Reset)
	assert.Equal(t, expectedRaw, f.Raw())
	assert.Equal(t, expectedFormatted, f.Format())
	assert.Equal(t, expectedFormatted, f.String())

}

func TestSprintf(t *testing.T) {
	expectedColor := ansi.Blue
	template := "foo %s"
	expectedResult := string(expectedColor) + "foo bar" + string(ansi.Reset)
	res := Sprintf(expectedColor, template, "bar")

	assert.Equal(t, expectedResult, res)
}
func TestSprintf_FormattedObjects(t *testing.T) {
	expectedColor := ansi.Blue
	template := "foo %s bar"

	expectedObjColor := ansi.Green
	expectedObjMsg := "baz"
	obj := New(expectedObjColor, expectedObjMsg)

	expectedResult := string(expectedColor) + "foo " + string(ansi.Reset) + string(expectedObjColor) + expectedObjMsg + string(ansi.Reset) + string(expectedColor) + " bar" + string(ansi.Reset)
	res := Sprintf(expectedColor, template, obj)

	assert.Equal(t, expectedResult, res)
}
