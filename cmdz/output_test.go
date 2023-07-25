package cmdz

import (
	//"fmt"
	//"io"

	"fmt"

	//"log"
	//"os/exec"
	"strings"
	"testing"

	"mby.fr/utils/collections"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputString(t *testing.T) {
	e := OutputCmd("echo", "-n", "foo")
	o, err := e.OutputString()
	require.NoError(t, err)
	assert.Equal(t, "foo", o)
	assert.Equal(t, "foo", e.StdoutRecord())
	assert.Equal(t, "", e.StderrRecord())

	e = OutputCmd("/bin/sh", "-c", ">&2 echo foo; echo bar")
	o, err = e.OutputString()
	require.NoError(t, err)
	assert.Equal(t, "bar\n", o)
	assert.Equal(t, "bar\n", e.StdoutRecord())
	assert.Equal(t, "foo\n", e.StderrRecord())

	f := OutputCmd("/bin/false")
	_, err = f.OutputString()
	require.Error(t, err)
}

func TestCombinedOutputString(t *testing.T) {
	e := OutputCmd("echo", "-n", "foo")
	o, err := e.CombinedOutputString()
	require.NoError(t, err)
	assert.Equal(t, "foo", o)
	assert.Equal(t, "foo", e.StdoutRecord())
	assert.Equal(t, "", e.StderrRecord())

	e = OutputCmd("/bin/sh", "-c", ">&2 echo foo; sleep 0.1; echo bar")
	o, err = e.CombinedOutputString()
	require.NoError(t, err)
	assert.Equal(t, "foo\nbar\n", o)
	//assert.Equal(t, "bar\n", e.StdoutRecord())
	//assert.Equal(t, "foo\n", e.StderrRecord())
	assert.Equal(t, "foo\nbar\n", e.StdoutRecord())
	assert.Equal(t, "", e.StderrRecord())

	f := OutputCmd("/bin/false")
	_, err = f.CombinedOutputString()
	require.Error(t, err)
}

var trimNewLineProcesser = func(rc int, stdout, stderr string) (string, error) {
	out := strings.ReplaceAll(stdout, "\n", "")
	return out, nil
}

var appendLineStringProcesser = func(rc int, stdout, stderr string) (string, error) {
	lines := strings.Split(stdout, "\n")
	appended := collections.Map(lines, func(in string) string {
		return in + "END"
	})
	out := strings.Join(appended, "\n")
	return out, nil
}

var errorProcesser = func(rc int, stdout, stderr string) (string, error) {
	return "", fmt.Errorf("Error")
}

func TestStringProcess_OutputString(t *testing.T) {
	echo := OutputCmd("/bin/echo", "-n", "foo")
	p := echo.OutStringProcess(appendLineStringProcesser)
	o, err := p.OutputString()
	require.NoError(t, err)
	assert.Equal(t, "fooEND", o)

	echo = OutputCmd("/bin/echo", "foo")
	p = echo.OutStringProcess(appendLineStringProcesser)
	o, err = p.OutputString()
	require.NoError(t, err)
	assert.Equal(t, "fooEND\nEND", o)

	echo = OutputCmd("/bin/echo", "foo")
	p = echo.OutStringProcess(trimNewLineProcesser, appendLineStringProcesser)
	o, err = p.OutputString()
	require.NoError(t, err)
	require.Equal(t, "fooEND", o)

	echo = OutputCmd("/bin/echo", "foo")
	p = echo.OutStringProcess(errorProcesser, trimNewLineProcesser)
	o, err = p.OutputString()
	require.Error(t, err)
	assert.Equal(t, "", o)
}
