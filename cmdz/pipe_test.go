package cmdz

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicPipe(t *testing.T) {
	left := Sh(`echo foo`)
	stdout := strings.Builder{}
	left.SetStdout(&stdout)
	right := Sh(`sed -e 's/o/a/'`)
	p := basicPipe{
		feeder: left,
	}

	rc, err := p.Pipe(right).BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "fao\n", p.StdoutRecord())
	assert.Equal(t, "", stdout.String())

	rc, err = left.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo\n", left.StdoutRecord())
	assert.Equal(t, "foo\n", stdout.String())
}

func TestBasicPipe_PipeFail(t *testing.T) {
	left := Sh(`true`)
	right := Sh(`echo foo`)
	p := basicPipe{
		feeder: left,
	}

	rc, err := p.Pipe(right).BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo\n", p.StdoutRecord())

	left = Sh(`false`)
	right = Sh(`echo foo`)
	p = basicPipe{
		feeder: left,
	}
	rc, err = p.Pipe(right).BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo\n", p.StdoutRecord())

	left = Sh(`false`)
	right = Sh(`echo foo`)
	p = basicPipe{
		feeder: left,
	}
	rc, err = p.PipeFail(right).BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 1, rc)
	assert.Equal(t, "", p.StdoutRecord())
}
