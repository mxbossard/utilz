package display

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mby.fr/utils/printz"
)

func TestGetScreen(t *testing.T) {
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	s := NewScreen(outs)
	assert.NotNil(t, s)
}

func TestGetAsyncScreen(t *testing.T) {
	// outW := &strings.Builder{}
	// errW := &strings.Builder{}
	// outs := printz.NewOutputs(outW, errW)
	tmpDir := "/tmp/foo42"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir)
	assert.NotNil(t, s)
	assert.DirExists(t, tmpDir)

	require.Panics(t, func() {
		NewAsyncScreen(tmpDir)
	})
}

func TestGetReadOnlyAsyncScreen(t *testing.T) {
	tmpDir := "/tmp/foo42"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir)
	assert.NotNil(t, s)
	assert.DirExists(t, tmpDir)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	s2 := NewAsyncScreenTailer(outs, tmpDir)
	assert.NotNil(t, s2)

	require.Panics(t, func() {
		NewAsyncScreenTailer(outs, "/tmp/notExistingDir")
	})

}
