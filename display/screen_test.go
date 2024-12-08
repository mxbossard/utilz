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

func TestScreenGetSession(t *testing.T) {
	// outW := &strings.Builder{}
	// errW := &strings.Builder{}
	// outs := printz.NewOutputs(outW, errW)
	tmpDir := "/tmp/foo42"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir)
	require.NotNil(t, s)
	session := s.Session("bar", 42)
	assert.NotNil(t, session)
	assert.DirExists(t, tmpDir)
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

func TestScreenGetPrinter(t *testing.T) {
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

func TestAsyncScreen_Basic(t *testing.T) {
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
	prtr10 := session.Printer(expectedPrinter, 10)
	require.NotNil(t, prtr10)

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

	prtr10.Out(expectedMessage)
	// assert.Empty(t, outW.String())
	// assert.Empty(t, errW.String())
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	err = prtr10.Flush()
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

func TestAsyncScreen_MultiplePrinters(t *testing.T) {
	tmpDir := "/tmp/foo42"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	expectedSession := "foo"
	expectedPrinter10a := "bar10a"
	expectedPrinter20a := "bar20a"
	expectedPrinter20b := "bar20b"
	expectedPrinter30a := "bar30a"

	session := screen.Session(expectedSession, 42)
	err := session.Start(1000)
	assert.NoError(t, err)

	sessionTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(tmpDir + "/" + expectedSession + outFileNameSuffix)
		return matches[0]
	}()
	assert.FileExists(t, sessionTmpOutFilepath)

	prtr10a := session.Printer(expectedPrinter10a, 10)
	prtr20a := session.Printer(expectedPrinter20a, 20)
	prtr20b := session.Printer(expectedPrinter20b, 20)
	prtr30a := session.Printer(expectedPrinter30a, 30)

	prtr20a.Out("20a-1,")
	prtr20a.Out("20a-2,")
	prtr10a.Out("10a-1,")
	prtr30a.Out("30a-1,")
	prtr10a.Out("10a-2,")
	prtr20b.Out("20b-1,")
	prtr20a.Out("20a-3,")

	// First flush, nothing is Closed => first printers should be written only
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	err = screenTailer.Flush()
	assert.NoError(t, err)

	assert.Equal(t, "10a-1,10a-2,", outW.String())
	assert.Empty(t, errW.String())

	// Close a printer which is not first => nothing more should be written
	session.Close(expectedPrinter20a)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = screenTailer.Flush()
	assert.NoError(t, err)

	assert.Equal(t, "10a-1,10a-2,", outW.String())
	assert.Empty(t, errW.String())

	prtr10a.Out("10a-3,")
	prtr20b.Out("20b-2,")

	// Close first printer => should write next printers
	session.Close(expectedPrinter10a)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = screenTailer.Flush()
	assert.NoError(t, err)

	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,", outW.String())
	assert.Empty(t, errW.String())

	// Close a printer which is not first => should not write anything
	session.Close(expectedPrinter30a)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = screenTailer.Flush()
	assert.NoError(t, err)

	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,", outW.String())
	assert.Empty(t, errW.String())

	// Close last printer => should write everything
	session.Close(expectedPrinter20b)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,"+"30a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = screenTailer.Flush()
	assert.NoError(t, err)

	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,"+"30a-1,", outW.String())
	assert.Empty(t, errW.String())
}

func TestAsyncScreen_MultipleSessions(t *testing.T) {
	// TODO
}
