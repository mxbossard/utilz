package zcreen

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mxbossard/utilz/filez"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession_Get(t *testing.T) {
	tmpDir := "/tmp/session_test_420"
	expectedSession := "bar101"
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)

	session, err := buildSession(expectedSession, 42, tmpDir)
	require.NoError(t, err)
	require.NotNil(t, session)
	assert.Implements(t, (*Session)(nil), session)
}

func TestSession_Start(t *testing.T) {
	tmpDir := "/tmp/session_test_430"
	expectedSession := "bar101"
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)

	expectedSessionDir := forgeSessionDirPath(tmpDir, expectedSession, 42)
	session, err := buildSession(expectedSession, 42, tmpDir)
	require.NoError(t, err)
	require.NotNil(t, session)
	assert.DirExists(t, session.TmpPath)
	assert.Equal(t, expectedSessionDir, session.TmpPath)

	err = session.Start(10 * time.Millisecond)
	require.NoError(t, err)
	assert.DirExists(t, session.TmpPath)

	// matches, err := filepath.Glob(expectedSessionDir + "/" + expectedSession + outFileNameSuffix + "*")
	// require.NoError(t, err)
	// require.Len(t, matches, 1)
	// matches, err = filepath.Glob(expectedSessionDir + "/" + expectedSession + errFileNameSuffix + "*")
	// require.NoError(t, err)
	// require.Len(t, matches, 1)
}

func priorizedPrinterFilenameMatches(printersDir, name, qualifier string) []string {
	//wildcardPath := fmt.Sprintf("*__%s%s-*", name, qualifier)
	printerDirWildcardPath := fmt.Sprintf("*__%s", name)
	printerFileWildcardPath := fmt.Sprintf("%s.*", qualifier)
	wildcardPath := filepath.Join(printersDir, printerDirWildcardPath, printerFileWildcardPath)
	// fmt.Printf("wildcardPath: %s\n", wildcardPath)
	matches, err := filepath.Glob(wildcardPath)
	if err != nil {
		panic(err)
	}
	return matches
}

func TestSession_GetPrinter(t *testing.T) {
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
	require.NotNil(t, prtr)

	expectedPrintersDirPAth := printersDirPath(session.TmpPath)

	prtr.Out("foo")
	prtr.Err("bar")

	// Until first print no printer file should exists
	matches := priorizedPrinterFilenameMatches(expectedPrintersDirPAth, expectedPrinter, outFileQualifier)
	assert.Empty(t, matches)

	matches = priorizedPrinterFilenameMatches(expectedPrintersDirPAth, expectedPrinter, errFileQualifier)
	assert.Empty(t, matches)

	prtr.Flush()

	// After flushing printers files should be created
	printerTmpOutFilepath := func() string {
		matches := priorizedPrinterFilenameMatches(expectedPrintersDirPAth, expectedPrinter, outFileQualifier)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpErrFilepath := func() string {
		matches := priorizedPrinterFilenameMatches(expectedPrintersDirPAth, expectedPrinter, errFileQualifier)
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
	expectedMessage1 := "msg1"
	expectedMessage2 := "msg2"
	sessionTimeout := 10 * time.Millisecond
	require.NoError(t, os.RemoveAll(tmpDir))
	os.MkdirAll(tmpDir, 0744)

	expectedSessionDir := forgeSessionDirPath(tmpDir, expectedSession, 42)
	session, err := buildSession(expectedSession, 42, tmpDir)
	require.NoError(t, err)
	require.NotNil(t, session)

	// Opening a printer in a not started session should not error
	_, err = session.Printer(expectedPrinter, 42)
	assert.NoError(t, err)

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

	sessionSerFilepath := sessionSerializedPath(expectedSessionDir)
	expectedPrintersDir := printersDirPath(expectedSessionDir)
	matches := priorizedPrinterFilenameMatches(expectedPrintersDir, expectedPrinter, outFileQualifier)
	assert.Empty(t, matches)
	matches = priorizedPrinterFilenameMatches(expectedPrintersDir, expectedPrinter, errFileQualifier)
	assert.Empty(t, matches)

	require.FileExists(t, sessionSerFilepath)

	assert.NotEmpty(t, func() string { s, _ := filez.ReadString(sessionSerFilepath); return s })
	ser, err := deserializeSession(sessionSerFilepath)
	require.NoError(t, err)
	assert.NotNil(t, ser)

	prtr10.Out(expectedMessage1)

	err = prtr10.Flush()
	assert.NoError(t, err)

	printerTmpOutFilepath := func() string {
		matches := priorizedPrinterFilenameMatches(expectedPrintersDir, expectedPrinter, outFileQualifier)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	matches = priorizedPrinterFilenameMatches(expectedPrintersDir, expectedPrinter, errFileQualifier)
	require.Empty(t, matches)

	assert.Equal(t, expectedMessage1, filez.ReadStringOrPanic(printerTmpOutFilepath))

	prtr10.Out(expectedMessage2)
	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage1+expectedMessage2, filez.ReadStringOrPanic(printerTmpOutFilepath))

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

	expectedSessionDir := forgeSessionDirPath(tmpDir, expectedSession, 42)
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

	sessionSerFilepath := sessionSerializedPath(expectedSessionDir)
	expectedPrintersDir := printersDirPath(expectedSessionDir)
	matches := priorizedPrinterFilenameMatches(expectedPrintersDir, expectedPrinter, outFileQualifier)
	assert.Empty(t, matches)
	matches = priorizedPrinterFilenameMatches(expectedPrintersDir, expectedPrinter, errFileQualifier)
	assert.Empty(t, matches)

	assert.FileExists(t, sessionSerFilepath)

	assert.NotEmpty(t, func() string { s, _ := filez.ReadString(sessionSerFilepath); return s })
	ser, err := deserializeSession(sessionSerFilepath)
	require.NoError(t, err)
	assert.NotNil(t, ser)

	prtr10.Out(expectedMessage)

	err = prtr10.Flush()
	assert.NoError(t, err)

	printerTmpOutFilepath := func() string {
		matches := priorizedPrinterFilenameMatches(expectedPrintersDir, expectedPrinter, outFileQualifier)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	matches = priorizedPrinterFilenameMatches(expectedPrintersDir, expectedPrinter, errFileQualifier)
	assert.Empty(t, matches)

	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))

	time.Sleep(sessionTimeout + extraTimeout + 10*time.Millisecond)
	// Opening a printer after session timeout should panic
	_, err = session.Printer("another", 42)
	assert.Error(t, err)

	assert.DirExists(t, session.TmpPath)

	// Clear & reOpen session
	err = session.clear()
	assert.NoError(t, err)

	require.NoDirExists(t, session.TmpPath)
	assert.NoFileExists(t, printerTmpOutFilepath)

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

	matches = priorizedPrinterFilenameMatches(expectedPrintersDir, expectedPrinter, outFileQualifier)
	require.Empty(t, matches)
	matches = priorizedPrinterFilenameMatches(expectedPrintersDir, expectedPrinter, errFileQualifier)
	assert.Empty(t, matches)

	assert.FileExists(t, sessionSerFilepath)

	assert.NotEmpty(t, func() string { s, _ := filez.ReadString(sessionSerFilepath); return s })
	ser, err = deserializeSession(sessionSerFilepath)
	require.NoError(t, err)
	assert.NotNil(t, ser)

	prtr10.Out(expectedMessage2)

	err = prtr10.Flush()
	assert.NoError(t, err)

	printerTmpOutFilepath = func() string {
		matches := priorizedPrinterFilenameMatches(expectedPrintersDir, expectedPrinter, outFileQualifier)
		require.NotEmpty(t, matches)
		require.Len(t, matches, 1)
		return matches[0]
	}()
	assert.Equal(t, expectedMessage2, filez.ReadStringOrPanic(printerTmpOutFilepath))

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage2, filez.ReadStringOrPanic(printerTmpOutFilepath))

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

	// sessionTmpOutFilepath := session.tmpOutName
	// assert.FileExists(t, sessionTmpOutFilepath)
	sessionNotifierOutFilepath := session.notifier.outFilepath
	sessionNotifierErrFilepath := session.notifier.errFilepath

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

	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierOutFilepath)
	})
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	session.NotifyPrinter().Out("notif2,")

	// First print on not first printer => should not write
	prtr20a.Out("20a-1,")
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,notif2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	prtr20a.Out("20a-2,")
	prtr10a.Out("10a-1,")

	session.NotifyPrinter().Out("notif3,")

	// First flush, nothing is Closed => first printers should be written only
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,notif2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,notif2,"+"10a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,notif3,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	prtr15a.Out("15a-1,")
	prtr30a.Out("30a-1,")
	prtr10a.Out("10a-2,")
	prtr20c.Out("20c-1,")
	prtr20b.Out("20b-1,")

	// Nothing is Closed => first printers should be written only
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,notif2,"+"10a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,notif3,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,notif3,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	// Close a printer which is not first => nothing more should be written
	session.ClosePrinter(expectedPrinter20a, "msg")

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,notif3,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	assert.Panics(t, func() {
		prtr20a.Out("20a-3,")
	})

	prtr10a.Out("10a-3,")

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,10a-3,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,notif3,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	// Close first printer => should write next printers
	session.ClosePrinter(expectedPrinter10a, "msg")

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,10a-3,"+"15a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,notif3,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	session.ClosePrinter(expectedPrinter15a, "msg")
	session.ClosePrinter(expectedPrinter20c, "msg")

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,10a-3,"+"15a-1,"+"20a-1,20a-2,20a-3,20b-1,20c-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,notif3,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	prtr20b.Out("20b-2,")

	// Close a printer which is not first => should not write anything
	session.ClosePrinter(expectedPrinter30a, "msg")

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,10a-3,"+"15a-1,"+"20a-1,20a-2,20a-3,20b-1,20c-1,20b-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,notif3,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	// Close last printer => should write everything
	session.ClosePrinter(expectedPrinter20b, "msg")

	session.NotifyPrinter().Out("notif4,")

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,10a-3,"+"15a-1,"+"20a-1,20a-2,20a-3,20b-1,20c-1,20b-2,"+"30a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,notif3,notif4,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	time.Sleep(2 * time.Millisecond)
	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,10a-3,"+"15a-1,"+"20a-1,20a-2,20a-3,20b-1,20c-1,20b-2,"+"30a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,notif3,notif4,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	err = session.End("msg")
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,notif2,"+"10a-1,10a-2,10a-3,"+"15a-1,"+"20a-1,20a-2,20a-3,20b-1,20c-1,20b-2,"+"30a-1,"+"notif3,notif4,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,notif3,notif4,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})
}

func TestSession_EmptyPrinter(t *testing.T) {
	tmpDir := "/tmp/session_test_470"
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

	// sessionTmpOutFilepath := session.tmpOutName
	// assert.FileExists(t, sessionTmpOutFilepath)

	sessionNotifierOutFilepath := session.notifier.outFilepath
	sessionNotifierErrFilepath := session.notifier.errFilepath

	prtr10a, err := session.Printer(expectedPrinter10a, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10a)
	prtr20a, err := session.Printer(expectedPrinter20a, 20)
	require.NoError(t, err)
	require.NotNil(t, prtr20a)

	// Never print on 10a and close 10a
	// Consolidated session should contains what was printed on 20a
	_ = prtr10a

	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// err = session.Flush()
	// assert.NoError(t, err)
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))

	// First print on not first printer => should not write
	prtr20a.Out("20a-1,")
	session.NotifyPrinter().Out("notif1,")

	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})
	// prtr10a.Out("")

	session.ClosePrinter(expectedPrinter10a, "empty")
	session.ClosePrinter(expectedPrinter20a, "notEmpty")

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,20a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})
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

	// sessionTmpOutFilepath := session.tmpOutName
	// assert.FileExists(t, sessionTmpOutFilepath)

	sessionNotifierOutFilepath := session.notifier.outFilepath
	sessionNotifierErrFilepath := session.notifier.errFilepath

	prtr10a, err := session.Printer(expectedPrinter10a, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10a)
	prtr20a, err := session.Printer(expectedPrinter20a, 20)
	require.NoError(t, err)
	require.NotNil(t, prtr20a)

	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	prtr10a.Out("10a1,")
	prtr20a.Out("20a1,")

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,10a1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	session.NotifyPrinter().Out("notif2,")

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,10a1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	session.ClosePrinter(expectedPrinter20a, "msg")

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,10a1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	session.ClosePrinter(expectedPrinter10a, "msg")

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,10a1,20a1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,10a1,20a1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	time.Sleep(2 * time.Millisecond)
	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,10a1,20a1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})

	time.Sleep(sessionTimeout + extraTimeout)

	err = session.Flush()
	assert.NoError(t, err)
	// session flush does not consolidate session tmp files anymore
	// assert.Equal(t, "notif1,10a1,20a1,notif2,notifTimeout,", filez.ReadStringOrPanic(sessionTmpOutFilepath))
	// assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, "notif1,notif2,", filez.ReadStringOrPanic(sessionNotifierOutFilepath))
	assert.Panics(t, func() {
		filez.ReadStringOrPanic(sessionNotifierErrFilepath)
	})
}
