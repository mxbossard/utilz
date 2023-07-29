package cmdz

import (
	//"fmt"
	//"io"

	//"log"
	//"os/exec"

	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mby.fr/utils/inout"
)

func TestInputProcesser(t *testing.T) {
	// No input processer
	c := Cmd("cat")
	r := strings.NewReader("foo")
	c.Input(r)
	rc, err := c.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo", c.StdinRecord())
	assert.Equal(t, "foo", c.StdoutRecord())

	// No input processer
	p := Proccessed(Cmd("cat"))
	r = strings.NewReader("foo")
	p.Input(r)
	rc, err = p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo", p.StdinRecord())
	assert.Equal(t, "foo", p.StdoutRecord())

	prefixProcessor := inout.StringLineProcesser(func(in string) (out string, err error) {
		return "PREFIX" + in, nil
	})

	// Simple prefix input processer
	p = Proccessed(Cmd("cat"))
	r = strings.NewReader("foo")
	p.ProcessIn(prefixProcessor)
	p.Input(r)
	rc, err = p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "PREFIXfoo", p.StdinRecord())
	assert.Equal(t, "PREFIXfoo", p.StdoutRecord())

	// Reverse config order
	p = Proccessed(Cmd("cat"))
	r = strings.NewReader("foo")
	p.Input(r)
	p.ProcessIn(prefixProcessor)
	rc, err = p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "PREFIXfoo", p.StdinRecord())
	assert.Equal(t, "PREFIXfoo", p.StdoutRecord())
}
