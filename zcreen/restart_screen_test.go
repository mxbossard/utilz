package zcreen

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
	Disallow multiple Sink at same time => fileLock
	If a sink is inactive for a long period => relase the lock
	Add a ticker which update sink activity / lock
*/

func TestRestartAsyncScreen(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestRestartAsyncScreen"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir)
	assert.NotNil(t, s)
	assert.DirExists(t, tmpDir)

	require.NotPanics(t, func() {
		NewAsyncScreen(tmpDir)
	})
}
