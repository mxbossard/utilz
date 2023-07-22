package cmdz

import (
	//"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmdzMock(t *testing.T) {
	c := Cmd("echo", "foo")
	rc, err := c.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo\n", c.StdoutRecord())

	StartStringMock(t, func(c Vcmd) (int, string, string) {
		return 1, "bar", ""
	})
	rc, err = c.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 1, rc)
	assert.Equal(t, "bar", c.StdoutRecord())

	StopMock()

	rc, err = c.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo\n", c.StdoutRecord())

}
