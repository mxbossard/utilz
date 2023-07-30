package cmdz

import (
	//"fmt"
	//"io"

	//"log"
	//"os/exec"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputString(t *testing.T) {
	e := OutputtedCmd("echo", "-n", "foo")
	o, err := e.OutputString()
	require.NoError(t, err)
	assert.Equal(t, "foo", o)
	assert.Equal(t, "foo", e.StdoutRecord())
	assert.Equal(t, "", e.StderrRecord())

	e = OutputtedCmd("/bin/sh", "-c", ">&2 echo foo; echo bar")
	o, err = e.OutputString()
	require.NoError(t, err)
	assert.Equal(t, "bar\n", o)
	assert.Equal(t, "bar\n", e.StdoutRecord())
	assert.Equal(t, "foo\n", e.StderrRecord())

	f := OutputtedCmd("/bin/false")
	_, err = f.OutputString()
	require.Error(t, err)
}

func TestCombinedOutputString(t *testing.T) {
	e := OutputtedCmd("echo", "-n", "foo")
	o, err := e.CombinedOutputString()
	require.NoError(t, err)
	assert.Equal(t, "foo", o)
	assert.Equal(t, "foo", e.StdoutRecord())
	assert.Equal(t, "", e.StderrRecord())

	e = OutputtedCmd("/bin/sh", "-c", ">&2 echo foo; sleep 0.1; echo bar")
	o, err = e.CombinedOutputString()
	require.NoError(t, err)
	assert.Equal(t, "foo\nbar\n", o)
	//assert.Equal(t, "bar\n", e.StdoutRecord())
	//assert.Equal(t, "foo\n", e.StderrRecord())
	assert.Equal(t, "foo\nbar\n", e.StdoutRecord())
	assert.Equal(t, "", e.StderrRecord())

	f := OutputtedCmd("/bin/false")
	_, err = f.CombinedOutputString()
	require.Error(t, err)
}
