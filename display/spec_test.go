package display

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mby.fr/utils/filez"
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
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	tmpDir := "/tmp/foo"
	s := NewAsyncScreen(outs, tmpDir)
	assert.NotNil(t, s)
	assert.DirExists(t, tmpDir)
}

func TestGetSession(t *testing.T) {
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	tmpDir := "/tmp/foo"
	s := NewAsyncScreen(outs, tmpDir)
	require.NotNil(t, s)
	session := s.Session("bar", 42)
	assert.NotNil(t, session)
	assert.FileExists(t, tmpDir+"/bar")
}

func TestGetSessionPrinter(t *testing.T) {
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	tmpDir := "/tmp/foo"
	s := NewAsyncScreen(outs, tmpDir)
	require.NotNil(t, s)
	session := s.Session("foo", 42)
	require.NotNil(t, session)
	prtr := session.Printer("bar")
	assert.NotNil(t, prtr)
}

func TestAsyncPrint(t *testing.T) {
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	tmpDir := "/tmp/foo"
	s := NewAsyncScreen(outs, tmpDir)
	require.NotNil(t, s)
	session := s.Session("foo", 42)
	require.NotNil(t, session)
	prtr := session.Printer("bar")
	require.NotNil(t, prtr)

	prtr.Out("baz")
	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	err := prtr.Flush()
	assert.NoError(t, err)
	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	err = session.Flush()
	assert.NoError(t, err)
	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	require.DirExists(t, tmpDir)
	sessionTmpFilepath := tmpDir + "/foo"
	require.FileExists(t, sessionTmpFilepath)
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpFilepath))

	err = s.Flush()
	assert.NoError(t, err)
	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

}
