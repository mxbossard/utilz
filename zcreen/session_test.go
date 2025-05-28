package zcreen

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mxbossard/utilz/filez"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSession(t *testing.T) {
	tmpDir := "/tmp/session_test_420"
	expectedSession := "bar101"
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)

	session, err := buildSession(expectedSession, 42, tmpDir)
	require.NoError(t, err)
	require.NotNil(t, session)
	assert.Implements(t, (*Session)(nil), session)
}

func TestSessionStart(t *testing.T) {
	tmpDir := "/tmp/session_test_430"
	expectedSession := "bar101"
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)
	expectedSessionDir := sessionDirPath(tmpDir, expectedSession)

	session, err := buildSession(expectedSession, 42, tmpDir)
	require.NoError(t, err)
	require.NotNil(t, session)
	assert.NoDirExists(t, session.TmpPath)

	err = session.Start(10 * time.Millisecond)
	require.NoError(t, err)
	assert.DirExists(t, session.TmpPath)
	matches, err := filepath.Glob(expectedSessionDir + "/" + expectedSession + outFileNameSuffix)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	matches, err = filepath.Glob(expectedSessionDir + "/" + expectedSession + errFileNameSuffix)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}

func TestSessionGetPrinter(t *testing.T) {
	tmpDir := "/tmp/session_test_440"
	expectedSession := "bar201"
	expectedPrinter := "baz"
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)

	session, err := buildSession(expectedSession, 42, tmpDir)
	require.NoError(t, err)
	require.NotNil(t, session)

	session.Start(10 * time.Millisecond)
	require.NotNil(t, session)
	prtr, err := session.Printer(expectedPrinter, 10)
	require.NoError(t, err)
	assert.NotNil(t, prtr)

	expectedPrintersDirPAth := printersDirPath(session.TmpPath)

	printerTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(expectedPrintersDirPAth + "/" + expectedPrinter + outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(expectedPrintersDirPAth + "/" + expectedPrinter + errFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	assert.FileExists(t, printerTmpOutFilepath)
	assert.FileExists(t, printerTmpErrFilepath)
}

func TestSession_FileLayer(t *testing.T) {
	tmpDir := "/tmp/session_test_450"
	expectedSession := "bar301"
	expectedPrinter := "baz"
	expectedMessage := "msg"
	sessionTimeout := 10 * time.Millisecond
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)

	session, err := buildSession(expectedSession, 42, tmpDir)
	require.NoError(t, err)
	require.NotNil(t, session)

	// Opening a printer in a not started session should panic
	_, err = session.Printer(expectedPrinter, 42)
	assert.Error(t, err)

	err = session.Start(sessionTimeout)
	assert.NoError(t, err)
	prtr10, err := session.Printer(expectedPrinter, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10)

	require.DirExists(t, tmpDir)
	require.DirExists(t, session.TmpPath)

	// ReOpening a printer should not panic
	assert.NotPanics(t, func() {
		session.Printer(expectedPrinter, 42)
	})

	sessionSerFilepath := filepath.Join(tmpDir, expectedSession+serializedExtension)
	expectedSessionDir := sessionDirPath(tmpDir, expectedSession)
	expectedPrintersDir := printersDirPath(expectedSessionDir)
	sessionTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(expectedSessionDir + "/" + expectedSession + outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	sessionTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(expectedSessionDir + "/" + expectedSession + errFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(expectedPrintersDir + "/" + expectedPrinter + outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(expectedPrintersDir + "/" + expectedPrinter + errFileNameSuffix)
		require.NotEmpty(t, matches)
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

	time.Sleep(sessionTimeout + 10*time.Millisecond + extraTimeout)
	// Opening a printer after session timeout should panic
	_, err = session.Printer("another", 42)
	assert.Error(t, err)
}

func TestSession_ReOpen(t *testing.T) {
	tmpDir := "/tmp/session_test_451"
	expectedSession := "bar302"
	expectedPrinter := "baz"
	expectedMessage := "msg"
	expectedMessage2 := "pouf"
	sessionTimeout := 10 * time.Millisecond
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)

	session, err := buildSession(expectedSession, 42, tmpDir)
	require.NoError(t, err)
	require.NotNil(t, session)

	err = session.Start(sessionTimeout)
	assert.NoError(t, err)
	prtr10, err := session.Printer(expectedPrinter, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10)

	require.DirExists(t, tmpDir)
	require.DirExists(t, session.TmpPath)

	// ReOpening a printer should not panic
	assert.NotPanics(t, func() {
		session.Printer(expectedPrinter, 42)
	})

	sessionSerFilepath := filepath.Join(tmpDir, expectedSession+serializedExtension)
	expectedSessionDir := sessionDirPath(tmpDir, expectedSession)
	expectedPrintersDir := printersDirPath(expectedSessionDir)
	sessionTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(expectedSessionDir + "/" + expectedSession + outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	sessionTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(expectedSessionDir + "/" + expectedSession + errFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(expectedPrintersDir + "/" + expectedPrinter + outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(expectedPrintersDir + "/" + expectedPrinter + errFileNameSuffix)
		require.NotEmpty(t, matches)
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

	time.Sleep(sessionTimeout + extraTimeout + 10*time.Millisecond)
	// Opening a printer after session timeout should panic
	_, err = session.Printer("another", 42)
	assert.Error(t, err)

	// Clear & reOpen session
	err = session.clear()
	assert.NoError(t, err)

	require.NoDirExists(t, session.TmpPath)
	assert.NoFileExists(t, sessionTmpOutFilepath)
	assert.NoFileExists(t, sessionTmpErrFilepath)
	assert.NoFileExists(t, printerTmpOutFilepath)
	assert.NoFileExists(t, printerTmpErrFilepath)

	session, err = buildSession(expectedSession, 42, tmpDir)
	require.NoError(t, err)
	require.NotNil(t, session)

	err = session.Start(sessionTimeout)
	assert.NoError(t, err)
	prtr10, err = session.Printer(expectedPrinter, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10)

	require.DirExists(t, tmpDir)
	require.DirExists(t, session.TmpPath)

	// ReOpening a printer should not panic
	assert.NotPanics(t, func() {
		session.Printer(expectedPrinter, 42)
	})

	sessionSerFilepath = filepath.Join(tmpDir, expectedSession+serializedExtension)
	sessionTmpOutFilepath = func() string {
		matches, _ := filepath.Glob(expectedSessionDir + "/" + expectedSession + outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	sessionTmpErrFilepath = func() string {
		matches, _ := filepath.Glob(expectedSessionDir + "/" + expectedSession + errFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpOutFilepath = func() string {
		matches, _ := filepath.Glob(expectedPrintersDir + "/" + expectedPrinter + outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpErrFilepath = func() string {
		matches, _ := filepath.Glob(expectedPrintersDir + "/" + expectedPrinter + errFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()

	assert.FileExists(t, sessionSerFilepath)
	assert.FileExists(t, sessionTmpOutFilepath)
	assert.FileExists(t, sessionTmpErrFilepath)
	assert.FileExists(t, printerTmpOutFilepath)
	assert.FileExists(t, printerTmpErrFilepath)

	assert.NotEmpty(t, func() string { s, _ := filez.ReadString(sessionSerFilepath); return s })
	ser, err = deserializeSession(sessionSerFilepath)
	require.NoError(t, err)
	assert.NotNil(t, ser)

	prtr10.Out(expectedMessage2)
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	err = prtr10.Flush()
	assert.NoError(t, err)
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedMessage2, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage2, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedMessage2, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	time.Sleep(sessionTimeout + extraTimeout + 10*time.Millisecond)
	// Opening a printer after session timeout should panic
	_, err = session.Printer("another", 42)
	assert.Error(t, err)
}

func TestSession_MultiplePrinters(t *testing.T) {
	tmpDir := "/tmp/session_test_460"
	expectedSession := "bar401"
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)

	expectedPrinter10a := "bar10a"
	expectedPrinter15a := "bar15a"
	expectedPrinter20a := "bar20a"
	expectedPrinter20b := "bar20b"
	expectedPrinter20c := "bar20c"
	expectedPrinter30a := "bar30a"

	session, err := buildSession(expectedSession, 42, tmpDir)
	require.NoError(t, err)
	err = session.Start(10 * time.Millisecond)
	assert.NoError(t, err)

	session.NotifyPrinter().Out("notif1,")

	sessionTmpOutFilepath := session.TmpOutName
	assert.FileExists(t, sessionTmpOutFilepath)

	prtr10a, err := session.Printer(expectedPrinter10a, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10a)
	prtr15a, err := session.Printer(expectedPrinter15a, 15)
	require.NoError(t, err)
	require.NotNil(t, prtr15a)
	prtr20a, err := session.Printer(expectedPrinter20a, 20)
	require.NoError(t, err)
	require.NotNil(t, prtr20a)
	prtr20b, err := session.Printer(expectedPrinter20b, 20)
	require.NoError(t, err)
	require.NotNil(t, prtr20b)
	prtr20c, err := session.Printer(expectedPrinter20c, 20)
	require.NoError(t, err)
	require.NotNil(t, prtr20c)
	prtr30a, err := session.Printer(expectedPrinter30a, 30)
	require.NoError(t, err)
	require.NotNil(t, prtr30a)

	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	session.NotifyPrinter().Out("notif2,")

	// First print on not first printer => should not write
	prtr20a.Out("20a-1,")
	assert.Equal(t, "notif1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	prtr20a.Out("20a-2,")
	prtr10a.Out("10a-1,")

	session.NotifyPrinter().Out("notif3,")

	// First flush, nothing is Closed => first printers should be written only
	assert.Equal(t, "notif1,notif2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"10a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	prtr15a.Out("15a-1,")
	prtr30a.Out("30a-1,")
	prtr10a.Out("10a-2,")
	prtr20c.Out("20c-1,")
	prtr20b.Out("20b-1,")

	// Nothing is Closed => first printers should be written only
	assert.Equal(t, "notif1,notif2,"+"10a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	// Close a printer which is not first => nothing more should be written
	session.ClosePrinter(expectedPrinter20a, "msg")

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	prtr20a.Out("20a-3,")
	prtr10a.Out("10a-3,")

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,10a-3,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	// Close first printer => should write next printers
	session.ClosePrinter(expectedPrinter10a, "msg")

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,10a-3,"+"15a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	session.ClosePrinter(expectedPrinter15a, "msg")
	session.ClosePrinter(expectedPrinter20c, "msg")

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,10a-3,"+"15a-1,"+"20a-1,20a-2,20a-3,20b-1,20c-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	prtr20b.Out("20b-2,")

	// Close a printer which is not first => should not write anything
	session.ClosePrinter(expectedPrinter30a, "msg")

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,10a-3,"+"15a-1,"+"20a-1,20a-2,20a-3,20b-1,20c-1,20b-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	// Close last printer => should write everything
	session.ClosePrinter(expectedPrinter20b, "msg")

	session.NotifyPrinter().Out("notif4,")

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,10a-3,"+"15a-1,"+"20a-1,20a-2,20a-3,20b-1,20c-1,20b-2,"+"30a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	time.Sleep(2 * time.Millisecond)
	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,10a-3,"+"15a-1,"+"20a-1,20a-2,20a-3,20b-1,20c-1,20b-2,"+"30a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = session.End("msg")
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,10a-3,"+"15a-1,"+"20a-1,20a-2,20a-3,20b-1,20c-1,20b-2,"+"30a-1,"+"notif3,notif4,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

}

func TestSession_EmptyPrinter(t *testing.T) {
	tmpDir := "/tmp/session_test_460"
	expectedSession := "bar401"
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)

	expectedPrinter10a := "bar10a"
	expectedPrinter20a := "bar20a"

	session, err := buildSession(expectedSession, 42, tmpDir)
	require.NoError(t, err)
	require.NotNil(t, session)
	err = session.Start(10 * time.Millisecond)
	assert.NoError(t, err)

	sessionTmpOutFilepath := session.TmpOutName
	assert.FileExists(t, sessionTmpOutFilepath)

	prtr10a, err := session.Printer(expectedPrinter10a, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10a)
	prtr20a, err := session.Printer(expectedPrinter20a, 20)
	require.NoError(t, err)
	require.NotNil(t, prtr20a)

	// Never print on 10a and close 10a
	// Consolidated session should contains what was printed on 20a
	_ = prtr10a

	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	err = session.Flush()
	assert.NoError(t, err)
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))

	// First print on not first printer => should not write
	prtr20a.Out("20a-1,")
	session.NotifyPrinter().Out("notif1,")

	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	// prtr10a.Out("")

	session.ClosePrinter(expectedPrinter10a, "empty")
	session.ClosePrinter(expectedPrinter20a, "notEmpty")

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,20a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
}

func TestSession_Timeout(t *testing.T) {
	tmpDir := "/tmp/session_test_560"
	expectedSession := "bar501"
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)

	expectedPrinter10a := "bar10a"
	expectedPrinter20a := "bar20a"
	sessionTimeout := 10 * time.Millisecond

	session, err := buildSession(expectedSession, 42, tmpDir)
	require.NoError(t, err)
	err = session.Start(sessionTimeout, func(s Session) {
		s.NotifyPrinter().Out("notifTimeout,")
	})
	assert.NoError(t, err)

	session.NotifyPrinter().Out("notif1,")

	sessionTmpOutFilepath := session.TmpOutName
	assert.FileExists(t, sessionTmpOutFilepath)

	prtr10a, err := session.Printer(expectedPrinter10a, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10a)
	prtr20a, err := session.Printer(expectedPrinter20a, 20)
	require.NoError(t, err)
	require.NotNil(t, prtr20a)

	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	prtr10a.Out("10a1,")
	prtr20a.Out("20a1,")

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,10a1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	session.NotifyPrinter().Out("notif2,")

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,10a1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	session.ClosePrinter(expectedPrinter20a, "msg")

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,10a1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	session.ClosePrinter(expectedPrinter10a, "msg")

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,10a1,20a1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,10a1,20a1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	time.Sleep(2 * time.Millisecond)
	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,10a1,20a1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	time.Sleep(sessionTimeout + extraTimeout)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,10a1,20a1,notif2,notifTimeout,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
}
