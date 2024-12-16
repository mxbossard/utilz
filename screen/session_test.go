package screen

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mby.fr/utils/filez"
)

func TestGetSession(t *testing.T) {
	tmpDir := "/tmp/foo42"
	expectedSession := "bar101"
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)

	session := buildSession(expectedSession, 42, tmpDir)
	assert.NotNil(t, session)
	assert.Implements(t, (*Session)(nil), session)
}

func TestSessionStart(t *testing.T) {
	tmpDir := "/tmp/foo42"
	expectedSession := "bar101"
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)

	session := buildSession(expectedSession, 42, tmpDir)
	assert.NotNil(t, session)
	assert.NoDirExists(t, session.TmpPath)

	err := session.Start(10 * time.Millisecond)
	require.NoError(t, err)
	assert.DirExists(t, session.TmpPath)
	matches, err := filepath.Glob(tmpDir + "/bar101" + outFileNameSuffix)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	matches, err = filepath.Glob(tmpDir + "/bar101" + errFileNameSuffix)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}

func TestSessionGetPrinter(t *testing.T) {
	tmpDir := "/tmp/foo42"
	expectedSession := "bar201"
	expectedPrinter := "baz"
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)

	session := buildSession(expectedSession, 42, tmpDir)
	assert.NotNil(t, session)

	session.Start(10 * time.Millisecond)
	require.NotNil(t, session)
	prtr := session.Printer(expectedPrinter, 10)
	assert.NotNil(t, prtr)

	printerTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(session.TmpPath + "/" + expectedPrinter + outFileNameSuffix)
		return matches[0]
	}()
	printerTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(session.TmpPath + "/" + expectedPrinter + errFileNameSuffix)
		return matches[0]
	}()
	assert.FileExists(t, printerTmpOutFilepath)
	assert.FileExists(t, printerTmpErrFilepath)
}

func TestSession_Basic(t *testing.T) {
	tmpDir := "/tmp/foo42"
	expectedSession := "bar301"
	expectedPrinter := "baz"
	expectedMessage := "msg"
	sessionTimeout := 10 * time.Millisecond
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)

	session := buildSession(expectedSession, 42, tmpDir)
	assert.NotNil(t, session)

	// Opening a printer in a not started session should panic
	assert.Panics(t, func() {
		session.Printer(expectedPrinter, 42)
	})

	err := session.Start(sessionTimeout)
	assert.NoError(t, err)
	prtr10 := session.Printer(expectedPrinter, 10)
	require.NotNil(t, prtr10)

	require.DirExists(t, tmpDir)
	require.DirExists(t, session.TmpPath)

	// ReOpening a printer should panic
	assert.Panics(t, func() {
		session.Printer(expectedPrinter, 42)
	})

	sessionSerFilepath := filepath.Join(tmpDir, expectedSession+serializedExtension)
	sessionTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(tmpDir + "/" + expectedSession + outFileNameSuffix)
		return matches[0]
	}()
	sessionTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(tmpDir + "/" + expectedSession + errFileNameSuffix)
		return matches[0]
	}()
	printerTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(session.TmpPath + "/" + expectedPrinter + outFileNameSuffix)
		return matches[0]
	}()
	printerTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(session.TmpPath + "/" + expectedPrinter + errFileNameSuffix)
		return matches[0]
	}()

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
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	err = prtr10.Flush()
	assert.NoError(t, err)
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	time.Sleep(sessionTimeout + 2*time.Millisecond)
	// Opening a printer after session timeout should panic
	assert.Panics(t, func() {
		session.Printer("another", 42)
	})
}

func TestSession_MultiplePrinters(t *testing.T) {
	tmpDir := "/tmp/foo42"
	expectedSession := "bar401"
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)

	expectedPrinter10a := "bar10a"
	expectedPrinter15a := "bar15a"
	expectedPrinter20a := "bar20a"
	expectedPrinter20b := "bar20b"
	expectedPrinter20c := "bar20c"
	expectedPrinter30a := "bar30a"

	session := buildSession(expectedSession, 42, tmpDir)
	err := session.Start(10 * time.Millisecond)
	assert.NoError(t, err)

	sessionTmpOutFilepath := session.TmpOutName
	assert.FileExists(t, sessionTmpOutFilepath)

	prtr10a := session.Printer(expectedPrinter10a, 10)
	prtr15a := session.Printer(expectedPrinter15a, 15)
	prtr20a := session.Printer(expectedPrinter20a, 20)
	prtr20b := session.Printer(expectedPrinter20b, 20)
	prtr20c := session.Printer(expectedPrinter20c, 20)
	prtr30a := session.Printer(expectedPrinter30a, 30)

	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	err = session.Flush()
	assert.NoError(t, err)
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))

	// First print on not first printer => should not write
	prtr20a.Out("20a-1,")
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	err = session.Flush()
	assert.NoError(t, err)
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))

	prtr20a.Out("20a-2,")
	prtr10a.Out("10a-1,")

	// First flush, nothing is Closed => first printers should be written only
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	prtr15a.Out("15a-1,")
	prtr30a.Out("30a-1,")
	prtr10a.Out("10a-2,")
	prtr20c.Out("20c-1,")
	prtr20b.Out("20b-1,")

	// Nothing is Closed => first printers should be written only
	assert.Equal(t, "10a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	// Close a printer which is not first => nothing more should be written
	session.ClosePrinter(expectedPrinter20a)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	prtr20a.Out("20a-3,")
	prtr10a.Out("10a-3,")

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,10a-3,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	// Close first printer => should write next printers
	session.ClosePrinter(expectedPrinter10a)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,10a-3,"+"15a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	session.ClosePrinter(expectedPrinter15a)
	session.ClosePrinter(expectedPrinter20c)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,10a-3,"+"15a-1,"+"20a-1,20a-2,20a-3,20b-1,20c-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	prtr20b.Out("20b-2,")

	// Close a printer which is not first => should not write anything
	session.ClosePrinter(expectedPrinter30a)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,10a-3,"+"15a-1,"+"20a-1,20a-2,20a-3,20b-1,20c-1,20b-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	// Close last printer => should write everything
	session.ClosePrinter(expectedPrinter20b)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,10a-3,"+"15a-1,"+"20a-1,20a-2,20a-3,20b-1,20c-1,20b-2,"+"30a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

}
