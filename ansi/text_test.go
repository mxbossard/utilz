package ansi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestText(t *testing.T) {
	title := "title"
	p := "lorem ipsum"

	f := Text(Blue)
	f.Fcatln(Red, title)
	f.Catln(p)

	expected := string(Red) + title + string(Reset) + "\n" + string(Blue) + p + string(Reset) + "\n"
	assert.Equal(t, expected, f.String())
}
