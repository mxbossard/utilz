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

func TestGetSession(t *testing.T) {
	// outW := &strings.Builder{}
	// errW := &strings.Builder{}
	// outs := printz.NewOutputs(outW, errW)
	tmpDir := "/tmp/foo42"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir)
	require.NotNil(t, s)
	session := s.Session("bar", 42)
	assert.NotNil(t, session)
	assert.NoDirExists(t, tmpDir+"/printers_bar")

	err := session.Start(1000)
	require.NoError(t, err)
	assert.DirExists(t, tmpDir+"/printers_bar")
	matches, err := filepath.Glob(tmpDir + "/bar" + outFileNameSuffix)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	matches, err = filepath.Glob(tmpDir + "/bar" + errFileNameSuffix)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}

func TestGetSessionPrinter(t *testing.T) {
	// outW := &strings.Builder{}
	// errW := &strings.Builder{}
	// outs := printz.NewOutputs(outW, errW)
	tmpDir := "/tmp/foo42"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir)
	require.NotNil(t, s)
	session := s.Session("foo", 42)
	session.Start(1000)
	require.NotNil(t, session)
	prtr := session.Printer("bar", 10)
	assert.NotNil(t, prtr)
}

func TestAsyncPrint_Basic(t *testing.T) {
	tmpDir := "/tmp/foo42"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)
	require.NotNil(t, screen)

	expectedSession := "foo"
	expectedPrinter := "bar"
	expectedMessage := "baz"

	session := screen.Session(expectedSession, 42)
	require.NotNil(t, session)
	err := session.Start(1000)
	assert.NoError(t, err)
	prtr := session.Printer(expectedPrinter, 10)
	require.NotNil(t, prtr)

	sessionSerFilepath := filepath.Join(tmpDir, expectedSession+serializedExtension)
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
	assert.FileExists(t, sessionSerFilepath)
	assert.FileExists(t, sessionTmpOutFilepath)
	assert.FileExists(t, sessionTmpErrFilepath)
	assert.FileExists(t, printerTmpOutFilepath)
	assert.FileExists(t, printerTmpErrFilepath)

	assert.NotEmpty(t, func() string { s, _ := filez.ReadString(sessionSerFilepath); return s })
	ser, err := deserializeSession(sessionSerFilepath)
	require.NoError(t, err)
	assert.NotNil(t, ser)

	prtr.Out(expectedMessage)
	// assert.Empty(t, outW.String())
	// assert.Empty(t, errW.String())
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	err = prtr.Flush()
	assert.NoError(t, err)
	// assert.Empty(t, outW.String())
	// assert.Empty(t, errW.String())
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	err = session.Flush()
	assert.NoError(t, err)
	// assert.Empty(t, outW.String())
	// assert.Empty(t, errW.String())
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	err = screenTailer.Flush()
	assert.NoError(t, err)
	assert.NotEmpty(t, outW.String())
	assert.Equal(t, expectedMessage, outW.String())
	assert.Empty(t, errW.String())
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

}

func TestAsyncPrint_MultiplePrinters(t *testing.T) {

}

func TestAsyncPrint_MultipleSessions(t *testing.T) {

}
