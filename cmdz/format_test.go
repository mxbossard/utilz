package cmdz

import (
	//"fmt"
	//"io"

	//"log"
	//"os/exec"

	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormat(t *testing.T) {
	c := Cmd("echo", "-n", "foo")
	barFormatter := func(rc int, stdout, stderr []byte, err error) (string, error) {
		return "bar", nil
	}
	f := Formatted[string, error](c, barFormatter)
	o, err := f.Format()
	require.NoError(t, err)
	assert.Equal(t, "bar", o)

	errFormatter := func(rc int, stdout, stderr []byte, err error) (string, error) {
		return "", fmt.Errorf("errFromFormatter")
	}
	f = Formatted[string, error](c, errFormatter)
	o, err = f.Format()
	require.Error(t, err)
	assert.ErrorContains(t, err, "errFromFormatter")
	assert.Equal(t, "", o)

	f = FormattedCmd(barFormatter, "/bin/false")
	o, err = f.Format()
	require.NoError(t, err)
	assert.Equal(t, "bar", o)

	f = FormattedCmd(errFormatter, "/bin/false")
	o, err = f.Format()
	require.Error(t, err)
	assert.ErrorContains(t, err, "errFromFormatter")
	assert.Equal(t, "", o)
}
