package formatz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/mxbossard/utilz/anzi"
)

func TestUnformat(t *testing.T) {
	s := ""
	assert.Equal(t, s, Unformat(""))

	s0 := "foo"
	assert.Equal(t, s0, Unformat(s0))

	s1 := string(anzi.Black) + s0
	assert.Equal(t, s0, Unformat(s1))

	s1 = s0 + string(anzi.Black)
	assert.Equal(t, s0, Unformat(s1))

	s1 = string(anzi.Reset) + s0
	assert.Equal(t, s0, Unformat(s1))

	s1 = s0 + string(anzi.Reset)
	assert.Equal(t, s0, Unformat(s1))

	s1 = string(anzi.Black) + s0 + string(anzi.Reset)
	assert.Equal(t, s0, Unformat(s1))

	s1 = string(anzi.Reset) + s0 + string(anzi.Black)
	assert.Equal(t, s0, Unformat(s1))

	s1 = string(anzi.HilightBlack) + s0 + string(anzi.Reset)
	assert.Equal(t, s0, Unformat(s1))
}

func TestFormat(t *testing.T) {
	s := "foo"
	expectedRaw := s
	f := New(anzi.None, s)

	assert.Equal(t, expectedRaw, f.Raw())
	assert.Equal(t, expectedRaw, f.Format())
	assert.Equal(t, expectedRaw, f.String())

	f = New(anzi.Black, s)
	expectedFormatted := string(anzi.Black) + expectedRaw + string(anzi.Reset)
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
	s2 := string(anzi.Black) + s1 + string(anzi.Reset)
	f = New(anzi.Blue, s2)

	f.Squash(false)
	expectedRaw = s2
	expectedFormatted = string(anzi.Blue) + expectedRaw + string(anzi.Reset)
	assert.Equal(t, expectedRaw, f.Raw())
	assert.Equal(t, expectedFormatted, f.Format())
	assert.Equal(t, expectedFormatted, f.String())

	f.Squash(true)
	expectedRaw = s1
	expectedFormatted = string(anzi.Blue) + expectedRaw + string(anzi.Reset)
	assert.Equal(t, expectedRaw, f.Raw())
	assert.Equal(t, expectedFormatted, f.Format())
	assert.Equal(t, expectedFormatted, f.String())

}

func TestSprintf(t *testing.T) {
	expectedColor := anzi.Blue
	template := "foo %s"
	expectedResult := string(expectedColor) + "foo bar" + string(anzi.Reset)
	res := Sprintf(expectedColor, template, "bar")

	assert.Equal(t, expectedResult, res)
}
func TestSprintf_FormattedObjects(t *testing.T) {
	expectedColor := anzi.Blue
	template := "foo %s bar"

	expectedObjColor := anzi.Green
	expectedObjMsg := "baz"
	obj := New(expectedObjColor, expectedObjMsg)

	expectedResult := string(expectedColor) + "foo " + string(anzi.Reset) + string(expectedObjColor) + expectedObjMsg + string(anzi.Reset) + string(expectedColor) + " bar" + string(anzi.Reset)
	res := Sprintf(expectedColor, template, obj)

	assert.Equal(t, expectedResult, res)
}
