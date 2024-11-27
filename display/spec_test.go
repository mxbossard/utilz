package display

import (
	"os"
	"path/filepath"
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
	tmpDir := "/tmp/foo42"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(outs, tmpDir)
	assert.NotNil(t, s)
	assert.DirExists(t, tmpDir)
}

func TestGetSession(t *testing.T) {
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	tmpDir := "/tmp/foo42"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(outs, tmpDir)
	require.NotNil(t, s)
	session := s.Session("bar", 42)
	assert.NotNil(t, session)
	assert.DirExists(t, tmpDir+"/printers_bar")
	matches, err := filepath.Glob(tmpDir + "/bar" + outFileNameSuffix)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	matches, err = filepath.Glob(tmpDir + "/bar" + errFileNameSuffix)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}

func TestGetSessionPrinter(t *testing.T) {
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	tmpDir := "/tmp/foo42"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(outs, tmpDir)
	require.NotNil(t, s)
	session := s.Session("foo", 42)
	require.NotNil(t, session)
	prtr := session.Printer("bar", 10)
	assert.NotNil(t, prtr)
}

func TestAsyncPrint(t *testing.T) {
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	tmpDir := "/tmp/foo42"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(outs, tmpDir)
	require.NotNil(t, screen)

	expectedSession := "foo"
	expectedPrinter := "bar"
	expectedMessage := "baz"

	session := screen.Session(expectedSession, 42)
	require.NotNil(t, session)
	prtr := session.Printer(expectedPrinter, 10)
	require.NotNil(t, prtr)

	sessionDirTmpFilepath := filepath.Join(tmpDir, printerDirPrefix+expectedSession)
	sessionTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(tmpDir + "/" + expectedSession + outFileNameSuffix)
		return matches[0]
	}()
	sessionTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(tmpDir + "/" + expectedSession + errFileNameSuffix)
		return matches[0]
	}()
	printerTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedPrinter + outFileNameSuffix)
		return matches[0]
	}()
	printerTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedPrinter + errFileNameSuffix)
		return matches[0]
	}()

	require.DirExists(t, tmpDir)
	require.DirExists(t, sessionDirTmpFilepath)
	require.FileExists(t, sessionTmpOutFilepath)
	require.FileExists(t, sessionTmpErrFilepath)
	require.FileExists(t, printerTmpOutFilepath)
	require.FileExists(t, printerTmpErrFilepath)

	prtr.Out(expectedMessage)
	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	err := prtr.Flush()
	assert.NoError(t, err)
	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	err = session.Flush()
	assert.NoError(t, err)
	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	err = screen.Flush()
	assert.NoError(t, err)
	assert.NotEmpty(t, outW.String())
	assert.Equal(t, expectedMessage, outW.String())
	assert.Empty(t, errW.String())
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

}
