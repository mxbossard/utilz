package formatz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/mxbossard/utilz/anzi"
)

func TestText(t *testing.T) {
	title := "title"
	p := "lorem ipsum"

	f := Text(anzi.Blue)
	f.Fcatln(anzi.Red, title)
	f.Catln(p)

	expected := string(anzi.Red) + title + string(anzi.Reset) + "\n" + string(anzi.Blue) + p + string(anzi.Reset) + "\n"
	assert.Equal(t, expected, f.String())
}
