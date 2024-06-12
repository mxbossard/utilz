package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"mby.fr/utils/ansi"
)

func TestText(t *testing.T) {
	title := "title"
	p := "lorem ipsum"

	f := Text(ansi.Blue)
	f.Fcatln(ansi.Red, title)
	f.Catln(p)

	expected := string(ansi.Red) + title + string(ansi.Reset) + "\n" + string(ansi.Blue) + p + string(ansi.Reset) + "\n"
	assert.Equal(t, expected, f.String())
}
