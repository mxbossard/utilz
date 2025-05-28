package zcreen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/printz"
	"github.com/mxbossard/utilz/zlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// test context initialization here
	zlog.ColoredConfig()
	//zlog.SetLogLevelThreshold(zlog.LevelPerf)
	zlog.PerfTimerStartAsTrace(false)
	os.Exit(m.Run())
}

func TestGetScreen(t *testing.T) {
	t.Skip()
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	s := NewScreen(outs)
	assert.NotNil(t, s)
	assert.Implements(t, (*Sink)(nil), s)
}

func TestGetTailer(t *testing.T) {
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	tmpDir := "/tmp/utilz.zcreen.foo40a"
	require.NoError(t, os.RemoveAll(tmpDir))
	err := os.MkdirAll(tmpDir, 0755)
	require.NoError(t, err)
	s := NewAsyncScreenTailer(outs, tmpDir)
	assert.NotNil(t, s)
	assert.Implements(t, (*Tailer)(nil), s)
}

func TestGetAsyncScreen(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo40b"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir)
	assert.NotNil(t, s)
	assert.DirExists(t, tmpDir)

	require.Panics(t, func() {
		NewAsyncScreen(tmpDir)
	})
}

func TestGetReadOnlyAsyncScreen(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo40c"
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
	tmpDir := "/tmp/utilz.zcreen.foo1001"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir)
	require.NotNil(t, s)
	session, err := s.Session("bar1001", 42)
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.DirExists(t, tmpDir)
	assert.NoDirExists(t, tmpDir+"/"+sessionDirPrefix+"bar1001")

	err = session.Start(100 * time.Millisecond)
	require.NoError(t, err)
	expectedSessionDir := filepath.Join(tmpDir, sessionDirPrefix+"bar1001")
	assert.DirExists(t, expectedSessionDir)
	matches, err := filepath.Glob(expectedSessionDir + "/bar1001" + outFileNameSuffix)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	matches, err = filepath.Glob(expectedSessionDir + "/bar1001" + errFileNameSuffix)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}

func TestScreenGetPrinter(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo2001"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir)
	require.NotNil(t, s)
	session, err := s.Session("foo2001", 42)
	require.NoError(t, err)
	require.NotNil(t, session)
	session.Start(100 * time.Millisecond)
	prtr, err := session.Printer("bar", 10)
	require.NoError(t, err)
	assert.NotNil(t, prtr)
}

func TestAsyncScreen_BasicOut(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo3001"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)
	require.NotNil(t, screen)

	expectedSession := "foo3001"
	expectedPrinter := "bar"
	expectedMessage := "baz"

	session, err := screen.Session(expectedSession, 42)
	require.NoError(t, err)
	require.NotNil(t, session)
	err = session.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtr10, err := session.Printer(expectedPrinter, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10)

	sessionSerFilepath := filepath.Join(tmpDir, expectedSession+serializedExtension)
	sessionDirTmpFilepath := filepath.Join(tmpDir, sessionDirPrefix+expectedSession)
	printersDirTmpFilepath := printersDirPath(sessionDirTmpFilepath)
	sessionTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedSession + outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	sessionTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedSession + errFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(printersDirTmpFilepath + "/" + expectedPrinter + outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(printersDirTmpFilepath + "/" + expectedPrinter + errFileNameSuffix)
		require.NotEmpty(t, matches)
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

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	//err = screenTailer.flushAll0()
	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.NotEmpty(t, outW.String())
	assert.Equal(t, expectedMessage, outW.String())
	assert.Empty(t, errW.String())
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))
}

func TestAsyncScreen_ClearSession(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo3002"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)
	require.NotNil(t, screen)

	expectedSession := "foo3002"
	expectedPrinter := "bar"
	expectedMessage := "baz"
	expectedMessage2 := "pif"

	session, err := screen.Session(expectedSession, 42)
	require.NoError(t, err)
	require.NotNil(t, session)
	err = session.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtr10, err := session.Printer(expectedPrinter, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10)

	prtr10.Out(expectedMessage)
	err = prtr10.Flush()
	assert.NoError(t, err)

	err = session.Flush()
	assert.NoError(t, err)

	err = session.End("msg")
	assert.NoError(t, err)

	// First tailing should tail printed message
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, outW.String())

	// Second tailing should not tail anything
	outW.Reset()
	errW.Reset()
	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Empty(t, outW.String())

	// New tailer should tail printed message
	outW.Reset()
	errW.Reset()
	screenTailer = NewAsyncScreenTailer(outs, tmpDir)
	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, outW.String())

	// After cleared session new tailer should not tail printed message
	outW.Reset()
	errW.Reset()
	// err = screen.ClearSession(expectedSession)
	// assert.NoError(t, err)
	screenTailer = NewAsyncScreenTailer(outs, tmpDir)
	err = screenTailer.ClearSession(expectedSession)
	assert.NoError(t, err)

	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Empty(t, outW.String())

	// After cleared session a session reopening should be possible
	err = screen.Resync()
	assert.NoError(t, err)

	session, err = screen.Session(expectedSession, 42)
	require.NoError(t, err)
	require.NotNil(t, session)
	err = session.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtr10, err = session.Printer(expectedPrinter, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10)

	prtr10.Out(expectedMessage2)
	err = prtr10.Flush()
	assert.NoError(t, err)

	err = session.Flush()
	assert.NoError(t, err)

	err = session.End("msg")
	assert.NoError(t, err)

	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage2, outW.String())
}

func TestAsyncScreen_BasicOutAndErr(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo3003"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)
	require.NotNil(t, screen)

	expectedSession := "foo3003"
	expectedPrinter := "bar"
	expectedOutMessage := "baz"
	expectedErrMessage := "err"

	session, err := screen.Session(expectedSession, 42)
	require.NoError(t, err)
	require.NotNil(t, session)
	err = session.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtr10, err := session.Printer(expectedPrinter, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10)

	sessionSerFilepath := filepath.Join(tmpDir, expectedSession+serializedExtension)
	sessionDirTmpFilepath := filepath.Join(tmpDir, sessionDirPrefix+expectedSession)
	printersDirTmpFilepath := printersDirPath(sessionDirTmpFilepath)
	sessionTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedSession + outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	sessionTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedSession + errFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(printersDirTmpFilepath + "/" + expectedPrinter + outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpErrFilepath := func() string {
		matches, _ := filepath.Glob(printersDirTmpFilepath + "/" + expectedPrinter + errFileNameSuffix)
		require.NotEmpty(t, matches)
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

	prtr10.Out(expectedOutMessage)
	prtr10.Err(expectedErrMessage)
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))

	err = prtr10.Flush()
	assert.NoError(t, err)
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedOutMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Equal(t, expectedErrMessage, filez.ReadStringOrPanic(printerTmpErrFilepath))

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, expectedOutMessage, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, expectedErrMessage, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedOutMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Equal(t, expectedErrMessage, filez.ReadStringOrPanic(printerTmpErrFilepath))

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.NotEmpty(t, outW.String())
	assert.NotEmpty(t, errW.String())
	assert.Equal(t, expectedOutMessage, outW.String())
	assert.Equal(t, expectedErrMessage, errW.String())
	assert.Equal(t, expectedOutMessage, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Equal(t, expectedErrMessage, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedOutMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Equal(t, expectedErrMessage, filez.ReadStringOrPanic(printerTmpErrFilepath))
}

func TestAsyncScreen_MultiplePrinters(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo4001"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	expectedSession := "foo4001"
	expectedPrinter10a := "bar10a"
	expectedPrinter20a := "bar20a"
	expectedPrinter20b := "bar20b"
	expectedPrinter30a := "bar30a"

	session, err := screen.Session(expectedSession, 42)
	require.NoError(t, err)
	err = session.Start(100 * time.Millisecond)
	assert.NoError(t, err)

	sessionTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirPath(tmpDir, expectedSession) + "/" + expectedSession + outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	assert.FileExists(t, sessionTmpOutFilepath)

	prtr10a, err := session.Printer(expectedPrinter10a, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10a)
	prtr20a, err := session.Printer(expectedPrinter20a, 20)
	require.NoError(t, err)
	require.NotNil(t, prtr20a)
	prtr20b, err := session.Printer(expectedPrinter20b, 20)
	require.NoError(t, err)
	require.NotNil(t, prtr20b)
	prtr30a, err := session.Printer(expectedPrinter30a, 30)
	require.NoError(t, err)
	require.NotNil(t, prtr30a)

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

	err = screenTailer.tailAll()
	assert.NoError(t, err)

	assert.Equal(t, "10a-1,10a-2,", outW.String())
	assert.Empty(t, errW.String())

	// Close a printer which is not first => nothing more should be written
	err = session.ClosePrinter(expectedPrinter20a, "msg")
	assert.NoError(t, err)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = screenTailer.tailAll()
	assert.NoError(t, err)

	assert.Equal(t, "10a-1,10a-2,", outW.String())
	assert.Empty(t, errW.String())

	prtr10a.Out("10a-3,")
	prtr20b.Out("20b-2,")

	// Close first printer => should write next printers
	err = session.ClosePrinter(expectedPrinter10a, "msg")
	assert.NoError(t, err)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = screenTailer.tailAll()
	assert.NoError(t, err)

	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,", outW.String())
	assert.Empty(t, errW.String())

	// Close a printer which is not first => should not write anything
	err = session.ClosePrinter(expectedPrinter30a, "msg")
	assert.NoError(t, err)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = screenTailer.tailAll()
	assert.NoError(t, err)

	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,", outW.String())
	assert.Empty(t, errW.String())

	// Close last printer => should write everything
	err = session.ClosePrinter(expectedPrinter20b, "msg")
	assert.NoError(t, err)

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,"+"30a-1,", filez.ReadStringOrPanic(sessionTmpOutFilepath))

	err = screenTailer.tailAll()
	assert.NoError(t, err)

	assert.Equal(t, "10a-1,10a-2,10a-3,"+"20a-1,20a-2,20a-3,20b-1,20b-2,"+"30a-1,", outW.String())
	assert.Empty(t, errW.String())
}

func TestAsyncScreen_MultipleSessions(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo5001c"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	expectedSessionA := "barA5001"
	expectedPrinterA10a := "barA10a"
	expectedPrinterA20a := "barA20a"
	expectedSessionB := "barB5001"
	expectedPrinterB10a := "barB10a"
	expectedPrinterB30a := "barB30a"
	expectedSessionC := "barC5001"
	expectedPrinterC10a := "barC10a"
	expectedPrinterC30a := "barC30a"
	expectedPrinterC40a := "barC40a"

	sessionA, err := screen.Session(expectedSessionA, 12)
	require.NoError(t, err)
	require.NotNil(t, sessionA)
	err = sessionA.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtrA10a, err := sessionA.Printer(expectedPrinterA10a, 10)
	require.NoError(t, err)
	require.NotNil(t, prtrA10a)
	prtrA20a, err := sessionA.Printer(expectedPrinterA20a, 20)
	require.NoError(t, err)
	require.NotNil(t, prtrA20a)

	sessionB, err := screen.Session(expectedSessionB, 42)
	require.NoError(t, err)
	require.NotNil(t, sessionB)
	err = sessionB.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtrB10a, err := sessionB.Printer(expectedPrinterB10a, 10)
	require.NoError(t, err)
	require.NotNil(t, prtrB10a)
	prtrB30a, err := sessionB.Printer(expectedPrinterB30a, 30)
	require.NoError(t, err)
	require.NotNil(t, prtrB30a)

	sessionC, err := screen.Session(expectedSessionC, 42)
	require.NoError(t, err)
	require.NotNil(t, sessionC)
	err = sessionC.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtrC10a, err := sessionC.Printer(expectedPrinterC10a, 10)
	require.NoError(t, err)
	require.NotNil(t, prtrC10a)
	prtrC30a, err := sessionC.Printer(expectedPrinterC30a, 30)
	require.NoError(t, err)
	require.NotNil(t, prtrC30a)
	prtrC40a, err := sessionC.Printer(expectedPrinterC40a, 40)
	require.NoError(t, err)
	require.NotNil(t, prtrC40a)

	assert.Empty(t, filez.ReadStringOrPanic(sessionA.TmpOutName))
	assert.Empty(t, filez.ReadStringOrPanic(sessionB.TmpOutName))

	prtrB10a.Out("B10a1,")
	prtrB30a.Out("B30a1,")
	prtrA10a.Out("A10a1,")
	prtrB30a.Out("B30a2,")
	err = sessionB.ClosePrinter(expectedPrinterB30a, "msg")
	assert.NoError(t, err)
	prtrB10a.Out("B10a2,")
	err = sessionB.ClosePrinter(expectedPrinterB10a, "msg")
	assert.NoError(t, err)
	err = sessionB.End("msg")
	assert.NoError(t, err)
	err = sessionB.End("msg")
	assert.NoError(t, err)
	prtrA20a.Out("A20a1,")
	prtrA20a.Out("A20a2,")
	err = sessionA.ClosePrinter(expectedPrinterA20a, "msg")
	assert.NoError(t, err)
	prtrA10a.Out("A10a2,")
	err = sessionA.ClosePrinter(expectedPrinterA10a, "msg")
	assert.NoError(t, err)

	prtrC10a.Out("C10a1,")
	prtrC30a.Out("C30a1,")
	prtrC40a.Out("C40a1,")

	assert.Empty(t, outW.String())
	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Empty(t, outW.String())

	assert.Empty(t, filez.ReadStringOrPanic(sessionA.TmpOutName))
	assert.NotEmpty(t, filez.ReadStringOrPanic(sessionB.TmpOutName))

	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, "", outW.String())

	err = sessionA.Flush()
	assert.NoError(t, err)

	assert.NotEmpty(t, filez.ReadStringOrPanic(sessionA.TmpOutName))

	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,", outW.String())

	err = sessionA.End("msg")
	assert.NoError(t, err)

	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,B10a1,B10a2,B30a1,B30a2,", outW.String())

	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,B10a1,B10a2,B30a1,B30a2,", outW.String())

	assert.Equal(t, "", filez.ReadStringOrPanic(sessionC.TmpOutName))
	err = sessionC.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "C10a1,", filez.ReadStringOrPanic(sessionC.TmpOutName))
	assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,"+"B10a1,B10a2,B30a1,B30a2,", outW.String())

	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,"+"B10a1,B10a2,B30a1,B30a2,"+"C10a1,", outW.String())

	err = sessionC.ClosePrinter(expectedPrinterC10a, "msg")
	assert.NoError(t, err)
	err = sessionC.Flush()
	assert.NoError(t, err)

	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,"+"B10a1,B10a2,B30a1,B30a2,"+"C10a1,C30a1,", outW.String())

	err = sessionC.End("msg")
	assert.NoError(t, err)
	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,"+"B10a1,B10a2,B30a1,B30a2,"+"C10a1,C30a1,C40a1,", outW.String())
}

func TestAsyncScreen_Notifications(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo6001"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	expectedSessionA := "barA6001"
	expectedPrinterA10a := "barA10a"
	expectedPrinterA20a := "barA20a"
	expectedSessionB := "barB6001"
	expectedPrinterB10a := "barB10a"
	expectedPrinterB30a := "barB30a"
	expectedSessionC := "barC6001"
	expectedPrinterC10a := "barC20a"
	expectedPrinterC30a := "barC30a"
	expectedPrinterC40a := "barC40a"

	sessionA, err := screen.Session(expectedSessionA, 12)
	require.NoError(t, err)
	require.NotNil(t, sessionA)
	err = sessionA.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtrA10a, err := sessionA.Printer(expectedPrinterA10a, 10)
	require.NoError(t, err)
	require.NotNil(t, prtrA10a)
	prtrA20a, err := sessionA.Printer(expectedPrinterA20a, 20)
	require.NoError(t, err)
	require.NotNil(t, prtrA20a)

	sessionB, err := screen.Session(expectedSessionB, 42)
	require.NoError(t, err)
	require.NotNil(t, sessionB)
	err = sessionB.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtrB10a, err := sessionB.Printer(expectedPrinterB10a, 10)
	require.NoError(t, err)
	require.NotNil(t, prtrB10a)
	prtrB30a, err := sessionB.Printer(expectedPrinterB30a, 30)
	require.NoError(t, err)
	require.NotNil(t, prtrB30a)

	sessionC, err := screen.Session(expectedSessionC, 42)
	require.NoError(t, err)
	require.NotNil(t, sessionC)
	err = sessionC.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtrC10a, err := sessionC.Printer(expectedPrinterC10a, 10)
	require.NoError(t, err)
	require.NotNil(t, prtrC10a)
	prtrC30a, err := sessionC.Printer(expectedPrinterC30a, 30)
	require.NoError(t, err)
	require.NotNil(t, prtrC30a)
	prtrC40a, err := sessionC.Printer(expectedPrinterC40a, 40)
	require.NoError(t, err)
	require.NotNil(t, prtrC40a)

	assert.Empty(t, filez.ReadStringOrPanic(sessionA.TmpOutName))
	assert.Empty(t, filez.ReadStringOrPanic(sessionB.TmpOutName))

	screen.NotifyPrinter().Out("notif1,")

	prtrB10a.Out("B10a1,")
	prtrB30a.Out("B30a1,")
	prtrA10a.Out("A10a1,")
	prtrB30a.Out("B30a2,")
	err = sessionB.ClosePrinter(expectedPrinterB30a, "msg")
	assert.NoError(t, err)
	prtrB10a.Out("B10a2,")
	err = sessionB.ClosePrinter(expectedPrinterB10a, "msg")
	assert.NoError(t, err)
	err = sessionB.End("msg")
	assert.NoError(t, err)
	prtrA20a.Out("A20a1,")
	prtrA20a.Out("A20a2,")
	err = sessionA.ClosePrinter(expectedPrinterA20a, "msg")
	assert.NoError(t, err)
	prtrA10a.Out("A10a2,")
	err = sessionA.ClosePrinter(expectedPrinterA10a, "msg")
	assert.NoError(t, err)

	screen.NotifyPrinter().Out("notif2,")
	err = screen.NotifyPrinter().Flush()
	assert.NoError(t, err)
	screen.NotifyPrinter().Out("notif3,")

	prtrC10a.Out("C10a1,")
	prtrC30a.Out("C30a1,")
	prtrC40a.Out("C40a1,")

	// Nothing should be flush yet
	assert.Empty(t, outW.String())

	err = screenTailer.tailNotifications() // should tail flushed notifs
	assert.NoError(t, err)

	_, err = screenTailer.tailNext() // should tail flushed notifs if no session opened
	assert.NoError(t, err)

	screen.NotifyPrinter().Out("notif4,")
	err = screen.NotifyPrinter().Flush()
	assert.NoError(t, err)

	// Only first 2 notifications are tailed. SessionB is flushed, SessionA is elected but not ended.
	assert.Equal(t, "notif1,notif2,", outW.String())

	assert.Empty(t, filez.ReadStringOrPanic(sessionA.TmpOutName))
	assert.NotEmpty(t, filez.ReadStringOrPanic(sessionB.TmpOutName))

	// First 4 notifications are flushed. SessionB is flushed, SessionA is elected but not ended.
	assert.Equal(t, "notif1,notif2,", outW.String())

	err = sessionA.Flush()
	assert.NoError(t, err)

	assert.NotEmpty(t, filez.ReadStringOrPanic(sessionA.TmpOutName))

	err = screenTailer.tailAll()
	assert.NoError(t, err)

	// First 4 notifications are flushed. SessionA is flushed but not ended, SessionB is ended.
	assert.Equal(t, "notif1,notif2,"+"A10a1,A10a2,A20a1,A20a2,", outW.String())

	err = sessionA.End("msg")
	assert.NoError(t, err)

	err = screenTailer.tailAll()
	assert.NoError(t, err)

	// First 4 notifications are flushed. SessionA & SessionB are ended. SessionC not flushed.
	assert.Equal(t, "notif1,notif2,"+"A10a1,A10a2,A20a1,A20a2,"+"notif3,notif4,"+"B10a1,B10a2,B30a1,B30a2,", outW.String())

	err = screenTailer.tailAll()
	assert.NoError(t, err)

	// SessionC still not flushed.
	assert.Equal(t, "notif1,notif2,"+"A10a1,A10a2,A20a1,A20a2,"+"notif3,notif4,"+"B10a1,B10a2,B30a1,B30a2,", outW.String())

	assert.Equal(t, "", filez.ReadStringOrPanic(sessionC.TmpOutName))
	err = sessionC.Flush()
	assert.NoError(t, err)
	assert.Equal(t, "C10a1,", filez.ReadStringOrPanic(sessionC.TmpOutName))

	// SessionC is flushed, but tailer not flushed
	assert.Equal(t, "notif1,notif2,"+"A10a1,A10a2,A20a1,A20a2,"+"notif3,notif4,"+"B10a1,B10a2,B30a1,B30a2,", outW.String())

	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"A10a1,A10a2,A20a1,A20a2,"+"notif3,notif4,"+"B10a1,B10a2,B30a1,B30a2,"+"C10a1,", outW.String())

	screen.NotifyPrinter().Out("notif5,")
	err = screen.NotifyPrinter().Flush()
	assert.NoError(t, err)

	err = sessionC.ClosePrinter(expectedPrinterC10a, "msg")
	assert.NoError(t, err)
	err = sessionC.Flush()
	assert.NoError(t, err)

	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"A10a1,A10a2,A20a1,A20a2,"+"notif3,notif4,"+"B10a1,B10a2,B30a1,B30a2,"+"C10a1,C30a1,", outW.String())

	err = sessionC.End("msg")
	assert.NoError(t, err)
	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, "notif1,notif2,"+"A10a1,A10a2,A20a1,A20a2,"+"notif3,notif4,"+"B10a1,B10a2,B30a1,B30a2,"+"C10a1,C30a1,C40a1,"+"notif5,", outW.String())
}

func TestTailOnlyBlocking(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo7011c"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	expectedSessionA := "barA7011"
	expectedPrinterA10a := "barA10a"
	expectedPrinterA20a := "barA20a"
	expectedSessionB := "barB7011"
	expectedPrinterB10a := "barB10a"
	expectedPrinterB30a := "barB30a"
	expectedSessionC := "barC7011"
	expectedPrinterC10a := "barC10a"
	expectedPrinterC30a := "barC30a"
	expectedPrinterC40a := "barC40a"

	syncChan := make(chan string)
	go func() {
		syncChan <- "startA"

		sessionA, err := screen.Session(expectedSessionA, 12)
		require.NoError(t, err)
		require.NotNil(t, sessionA)
		err = sessionA.Start(100 * time.Millisecond)
		assert.NoError(t, err)
		prtrA10a, err := sessionA.Printer(expectedPrinterA10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrA10a)
		prtrA20a, err := sessionA.Printer(expectedPrinterA20a, 20)
		require.NoError(t, err)
		require.NotNil(t, prtrA20a)

		prtrA10a.Out("A10a1,")
		prtrA10a.Out("A10a2,")
		err = sessionA.ClosePrinter(expectedPrinterA10a, "msg")
		assert.NoError(t, err)

		screen.NotifyPrinter().Out("notif1,")
		err = screen.NotifyPrinter().Flush()
		assert.NoError(t, err)
		// err = screenTailer.flushNotifications()
		// assert.NoError(t, err)

		prtrA20a.Out("A20a1,")
		prtrA20a.Out("A20a2,")
		err = sessionA.ClosePrinter(expectedPrinterA20a, "msg")
		assert.NoError(t, err)
		err = sessionA.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,", filez.ReadStringOrPanic(sessionA.TmpOutName))

		screen.NotifyPrinter().Out("notif2,")
		err = screen.NotifyPrinter().Flush()
		assert.NoError(t, err)

		// Wait before printing
		syncChan <- "startB"

		sessionB, err := screen.Session(expectedSessionB, 42)
		require.NoError(t, err)
		require.NotNil(t, sessionB)
		err = sessionB.Start(100 * time.Millisecond)
		assert.NoError(t, err)
		prtrB10a, err := sessionB.Printer(expectedPrinterB10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrB10a)
		prtrB30a, err := sessionB.Printer(expectedPrinterB30a, 30)
		require.NoError(t, err)
		require.NotNil(t, prtrB30a)

		prtrB10a.Out("B10a1,")
		prtrB10a.Out("B10a2,")
		err = sessionB.ClosePrinter(expectedPrinterB10a, "msg")
		assert.NoError(t, err)

		screen.NotifyPrinter().Out("notif3,")
		err = screen.NotifyPrinter().Flush()
		assert.NoError(t, err)

		prtrB30a.Out("B30a1,")
		prtrB30a.Out("B30a2,")
		err = sessionB.ClosePrinter(expectedPrinterB30a, "msg")
		assert.NoError(t, err)
		err = sessionB.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "B10a1,B10a2,B30a1,B30a2,", filez.ReadStringOrPanic(sessionB.TmpOutName))

		syncChan <- "endAB"

		// Wait before printing
		syncChan <- "startC"

		sessionC, err := screen.Session(expectedSessionC, 42)
		require.NoError(t, err)
		require.NotNil(t, sessionC)
		err = sessionC.Start(100 * time.Millisecond)
		assert.NoError(t, err)
		prtrC10a, err := sessionC.Printer(expectedPrinterC10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrC10a)
		prtrC30a, err := sessionC.Printer(expectedPrinterC30a, 30)
		require.NoError(t, err)
		require.NotNil(t, prtrC30a)
		prtrC40a, err := sessionC.Printer(expectedPrinterC40a, 40)
		require.NoError(t, err)
		require.NotNil(t, prtrC40a)

		prtrC10a.Out("C10a1,")
		err = sessionC.ClosePrinter(expectedPrinterC10a, "msg")
		assert.NoError(t, err)
		prtrC30a.Out("C30a1,")
		err = sessionC.ClosePrinter(expectedPrinterC30a, "msg")
		assert.NoError(t, err)
		prtrC40a.Out("C40a1,")
		err = sessionC.ClosePrinter(expectedPrinterC40a, "msg")
		assert.NoError(t, err)
		err = sessionC.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "C10a1,C30a1,C40a1,", filez.ReadStringOrPanic(sessionC.TmpOutName))
	}()

	err := screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Empty(t, outW.String())

	<-syncChan // startA
	<-syncChan // startB
	<-syncChan // endAB

	// Should Flush A before B because A is the first session available.
	err = screenTailer.TailOnlyBlocking(expectedSessionB, 3*continuousFlushPeriod+10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "B10a1,B10a2,B30a1,B30a2,", outW.String())
}

func TestAsyncScreen_TailOnlyBlocking_ClearedSession(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo3012"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)
	require.NotNil(t, screen)

	expectedSession := "foo3012"
	expectedPrinter := "bar"
	expectedMessage := "baz"

	session, err := screen.Session(expectedSession, 42)
	require.NoError(t, err)
	require.NotNil(t, session)
	err = session.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtr10, err := session.Printer(expectedPrinter, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10)

	prtr10.Out(expectedMessage)
	err = prtr10.Flush()
	assert.NoError(t, err)

	err = session.Flush()
	assert.NoError(t, err)

	// Clearing a session which was elected but not ended should be cleared

	// Force suite election
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)
	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, outW.String())

	// Then end & clear session
	err = session.End("msg")
	assert.NoError(t, err)
	// err = screen.ClearSession(expectedSession)
	// assert.NoError(t, err)

	// Should not block
	err = screenTailer.TailOnlyBlocking(expectedSession, 10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, outW.String())
}

func TestTailBlocking_InOrder(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo7001c"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	expectedSessionA := "barA7001"
	expectedPrinterA10a := "barA10a"
	expectedPrinterA20a := "barA20a"
	expectedSessionB := "barB7001"
	expectedPrinterB10a := "barB10a"
	expectedPrinterB30a := "barB30a"
	expectedSessionC := "barC7001"
	expectedPrinterC10a := "barC10a"
	expectedPrinterC30a := "barC30a"
	expectedPrinterC40a := "barC40a"

	// FIXME: use sync channel
	syncChan := make(chan string)
	go func() {
		// Wait before printing
		syncChan <- "startA"

		sessionA, err := screen.Session(expectedSessionA, 12)
		require.NoError(t, err)
		require.NotNil(t, sessionA)
		err = sessionA.Start(100 * time.Millisecond)
		assert.NoError(t, err)

		prtrA10a, err := sessionA.Printer(expectedPrinterA10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrA10a)
		prtrA20a, err := sessionA.Printer(expectedPrinterA20a, 20)
		require.NoError(t, err)
		require.NotNil(t, prtrA20a)

		prtrA10a.Out("A10a1,")
		prtrA10a.Out("A10a2,")
		err = sessionA.ClosePrinter(expectedPrinterA10a, "msg")
		assert.NoError(t, err)

		screen.NotifyPrinter().Out("notif1,")
		err = screen.NotifyPrinter().Flush()
		require.NoError(t, err)

		prtrA20a.Out("A20a1,")
		prtrA20a.Out("A20a2,")
		err = sessionA.ClosePrinter(expectedPrinterA20a, "msg")
		assert.NoError(t, err)

		syncChan <- "endB"

		err = sessionA.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,", filez.ReadStringOrPanic(sessionA.TmpOutName))

		// Wait before printing
		syncChan <- "startB"

		screen.NotifyPrinter().Out("notif2,")
		err = screen.NotifyPrinter().Flush()
		require.NoError(t, err)

		sessionB, err := screen.Session(expectedSessionB, 42)
		require.NoError(t, err)
		require.NotNil(t, sessionB)
		err = sessionB.Start(100 * time.Millisecond)
		assert.NoError(t, err)
		syncChan <- "startedB"

		prtrB10a, err := sessionB.Printer(expectedPrinterB10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrB10a)
		prtrB30a, err := sessionB.Printer(expectedPrinterB30a, 30)
		require.NoError(t, err)
		require.NotNil(t, prtrB30a)

		prtrB10a.Out("B10a1,")
		prtrB10a.Out("B10a2,")
		err = sessionB.ClosePrinter(expectedPrinterB10a, "msg")
		assert.NoError(t, err)

		prtrB30a.Out("B30a1,")
		prtrB30a.Out("B30a2,")
		err = sessionB.ClosePrinter(expectedPrinterB30a, "msg")
		assert.NoError(t, err)

		syncChan <- "finishB"
		err = sessionB.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "B10a1,B10a2,B30a1,B30a2,", filez.ReadStringOrPanic(sessionB.TmpOutName))

		// Wait before printing
		syncChan <- "startC"

		screen.NotifyPrinter().Out("notif3,")
		err = screen.NotifyPrinter().Flush()
		require.NoError(t, err)

		sessionC, err := screen.Session(expectedSessionC, 42)
		require.NoError(t, err)
		require.NotNil(t, sessionC)
		err = sessionC.Start(100 * time.Millisecond)
		assert.NoError(t, err)
		syncChan <- "startedC"

		prtrC10a, err := sessionC.Printer(expectedPrinterC10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrC10a)
		prtrC30a, err := sessionC.Printer(expectedPrinterC30a, 30)
		require.NoError(t, err)
		require.NotNil(t, prtrC30a)
		prtrC40a, err := sessionC.Printer(expectedPrinterC40a, 40)
		require.NoError(t, err)
		require.NotNil(t, prtrC40a)

		prtrC10a.Out("C10a1,")
		err = sessionC.ClosePrinter(expectedPrinterC10a, "msg")
		assert.NoError(t, err)
		prtrC30a.Out("C30a1,")
		err = sessionC.ClosePrinter(expectedPrinterC30a, "msg")
		assert.NoError(t, err)
		prtrC40a.Out("C40a1,")
		err = sessionC.ClosePrinter(expectedPrinterC40a, "msg")
		assert.NoError(t, err)
		err = sessionC.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "C10a1,C30a1,C40a1,", filez.ReadStringOrPanic(sessionC.TmpOutName))

		syncChan <- "endedC"
	}()

	err := screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Empty(t, outW.String())

	<-syncChan
	<-syncChan
	// Should Tail A before B because A is the first session available, but session A should not be available after timeout
	err = screenTailer.TailBlocking(expectedSessionB, 3*continuousFlushPeriod+10*time.Millisecond)
	assert.Error(t, err)
	assert.Equal(t, "notif1,"+"A10a1,A10a2,A20a1,A20a2,", outW.String())

	<-syncChan
	<-syncChan
	<-syncChan
	err = screenTailer.TailBlocking(expectedSessionA, 3*continuousFlushPeriod+10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "notif1,"+"A10a1,A10a2,A20a1,A20a2,"+"notif2,", outW.String())

	err = screenTailer.TailBlocking(expectedSessionB, 3*continuousFlushPeriod+10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "notif1,"+"A10a1,A10a2,A20a1,A20a2,"+"notif2,"+"B10a1,B10a2,B30a1,B30a2,", outW.String())

	<-syncChan
	<-syncChan
	<-syncChan
	err = screenTailer.TailBlocking(expectedSessionC, 3*continuousFlushPeriod+10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "notif1,"+"A10a1,A10a2,A20a1,A20a2,"+"notif2,"+"B10a1,B10a2,B30a1,B30a2,"+"notif3,"+"C10a1,C30a1,C40a1,", outW.String())
}

func TestAsyncScreen_TailBlocking_ClearedSession(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo3112"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)
	require.NotNil(t, screen)

	expectedSession := "foo3112"
	expectedPrinter := "bar"
	expectedMessage := "baz"

	session, err := screen.Session(expectedSession, 42)
	require.NoError(t, err)
	require.NotNil(t, session)
	err = session.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtr10, err := session.Printer(expectedPrinter, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10)

	prtr10.Out(expectedMessage)
	err = prtr10.Flush()
	assert.NoError(t, err)

	err = session.Flush()
	assert.NoError(t, err)

	// Clearing a session which was elected but not ended should be cleared

	// Force suite election
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)
	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, outW.String())

	// Then end & clear session
	err = session.End("msg")
	assert.NoError(t, err)
	// err = screen.ClearSession(expectedSession)
	// assert.NoError(t, err)

	// Should not block
	err = screenTailer.TailBlocking(expectedSession, 10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, outW.String())
}

func TestTailBlocking_InParallel(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo7201c"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	expectedSessionA := "barA7201"
	expectedPrinterA10a := "barA10a"
	expectedPrinterA20a := "barA20a"
	expectedSessionB := "barB7201"
	expectedPrinterB10a := "barB10a"
	expectedPrinterB30a := "barB30a"
	expectedSessionC := "barC7201"
	expectedPrinterC10a := "barC10a"
	expectedPrinterC30a := "barC30a"
	expectedPrinterC40a := "barC40a"

	syncChanA := make(chan string)
	syncChanB := make(chan string)
	syncChanC := make(chan string)
	go func() {
		// Wait before printing
		syncChanA <- "startA"

		sessionA, err := screen.Session(expectedSessionA, 12)
		require.NoError(t, err)
		require.NotNil(t, sessionA)
		err = sessionA.Start(100 * time.Millisecond)
		assert.NoError(t, err)

		prtrA10a, err := sessionA.Printer(expectedPrinterA10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrA10a)
		prtrA20a, err := sessionA.Printer(expectedPrinterA20a, 20)
		require.NoError(t, err)
		require.NotNil(t, prtrA20a)

		prtrA10a.Out("A10a1,")
		prtrA10a.Out("A10a2,")
		err = sessionA.ClosePrinter(expectedPrinterA10a, "msg")
		assert.NoError(t, err)
		prtrA20a.Out("A20a1,")
		prtrA20a.Out("A20a2,")
		err = sessionA.ClosePrinter(expectedPrinterA20a, "msg")
		assert.NoError(t, err)
		err = sessionA.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,", filez.ReadStringOrPanic(sessionA.TmpOutName))

		syncChanA <- "endA"
	}()

	go func() {
		// Wait before printing
		syncChanB <- "startB"

		sessionB, err := screen.Session(expectedSessionB, 42)
		require.NoError(t, err)
		require.NotNil(t, sessionB)
		err = sessionB.Start(100 * time.Millisecond)
		assert.NoError(t, err)

		prtrB10a, err := sessionB.Printer(expectedPrinterB10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrB10a)
		prtrB30a, err := sessionB.Printer(expectedPrinterB30a, 30)
		require.NoError(t, err)
		require.NotNil(t, prtrB30a)

		prtrB10a.Out("B10a1,")
		prtrB10a.Out("B10a2,")
		err = sessionB.ClosePrinter(expectedPrinterB10a, "msg")
		assert.NoError(t, err)
		prtrB30a.Out("B30a1,")
		prtrB30a.Out("B30a2,")
		err = sessionB.ClosePrinter(expectedPrinterB30a, "msg")
		assert.NoError(t, err)
		err = sessionB.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "B10a1,B10a2,B30a1,B30a2,", filez.ReadStringOrPanic(sessionB.TmpOutName))

		syncChanB <- "endB"
	}()

	// async prints
	go func() {
		// Wait before printing
		syncChanC <- "startC"

		sessionC, err := screen.Session(expectedSessionC, 42)
		require.NoError(t, err)
		require.NotNil(t, sessionC)
		err = sessionC.Start(100 * time.Millisecond)
		assert.NoError(t, err)

		prtrC10a, err := sessionC.Printer(expectedPrinterC10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrC10a)
		prtrC30a, err := sessionC.Printer(expectedPrinterC30a, 30)
		require.NoError(t, err)
		require.NotNil(t, prtrC30a)
		prtrC40a, err := sessionC.Printer(expectedPrinterC40a, 40)
		require.NoError(t, err)
		require.NotNil(t, prtrC40a)

		prtrC10a.Out("C10a1,")
		err = sessionC.ClosePrinter(expectedPrinterC10a, "msg")
		assert.NoError(t, err)
		prtrC30a.Out("C30a1,")
		err = sessionC.ClosePrinter(expectedPrinterC30a, "msg")
		assert.NoError(t, err)
		prtrC40a.Out("C40a1,")
		err = sessionC.ClosePrinter(expectedPrinterC40a, "msg")
		assert.NoError(t, err)
		err = sessionC.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "C10a1,C30a1,C40a1,", filez.ReadStringOrPanic(sessionC.TmpOutName))

		syncChanC <- "endC"
	}()

	err := screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Empty(t, outW.String())

	<-syncChanA
	<-syncChanB
	<-syncChanB
	<-syncChanA
	time.Sleep(10 * time.Millisecond)
	// Should Flush B before A despite of priority because B sessions will be available first
	err = screenTailer.TailBlocking(expectedSessionB, 3*continuousFlushPeriod+10*time.Millisecond)
	assert.NoError(t, err) // should timeout
	assert.Equal(t, "B10a1,B10a2,B30a1,B30a2,"+"A10a1,A10a2,A20a1,A20a2,", outW.String())

	err = screenTailer.TailBlocking(expectedSessionA, 3*continuousFlushPeriod+10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "B10a1,B10a2,B30a1,B30a2,"+"A10a1,A10a2,A20a1,A20a2,", outW.String())

	<-syncChanC
	<-syncChanC
	err = screenTailer.TailBlocking(expectedSessionC, 3*continuousFlushPeriod+10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "B10a1,B10a2,B30a1,B30a2,"+"A10a1,A10a2,A20a1,A20a2,"+"C10a1,C30a1,C40a1,", outW.String())
}

func TestTailBlocking_ContinuousFlow(t *testing.T) {

	// test TailBlocking() print messages contiously and not only at the end of a session
	tmpDir := "/tmp/utilz.zcreen.foo7301c"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	expectedSessionA := "barA7201"
	expectedPrinterA10a := "barA10a"
	expectedPrinterA20a := "barA20a"
	/*
		expectedSessionB := "barB7201"
		expectedPrinterB10a := "barB10a"
		expectedPrinterB30a := "barB30a"
		expectedSessionC := "barC7201"
		expectedPrinterC10a := "barC10a"
		expectedPrinterC30a := "barC30a"
		expectedPrinterC40a := "barC40a"
	*/

	syncChan := make(chan string)
	expectedMsgChan := make(chan string)

	go func() {
		sessionA, err := screen.Session(expectedSessionA, 12)
		require.NoError(t, err)
		require.NotNil(t, sessionA)
		err = sessionA.Start(300 * time.Millisecond)
		assert.NoError(t, err)
		prtrA10a, err := sessionA.Printer(expectedPrinterA10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrA10a)
		prtrA20a, err := sessionA.Printer(expectedPrinterA20a, 20)
		require.NoError(t, err)
		require.NotNil(t, prtrA20a)

		syncChan <- "A0"

		syncChan <- "A1" // wait syncChan consumed
		msg := "A10a1,"
		prtrA10a.Out(msg)
		err = sessionA.Flush()
		assert.NoError(t, err)
		expectedMsgChan <- msg // push expectedMessage

		syncChan <- "A2"
		msg = "A10a2,"
		prtrA10a.Out(msg)
		err = sessionA.Flush()
		assert.NoError(t, err)
		expectedMsgChan <- msg

		syncChan <- "A3"
		err = sessionA.ClosePrinter(expectedPrinterA10a, "msg")
		assert.NoError(t, err)
		expectedMsgChan <- ""

		syncChan <- "A4"
		msg = "A20a1,"
		prtrA20a.Out(msg)
		err = sessionA.Flush()
		assert.NoError(t, err)
		expectedMsgChan <- msg

		syncChan <- "A5"
		msg = "A20a2,"
		prtrA20a.Out(msg)
		err = sessionA.Flush()
		assert.NoError(t, err)
		expectedMsgChan <- msg

		syncChan <- "A6"
		err = sessionA.ClosePrinter(expectedPrinterA20a, "msg")
		assert.NoError(t, err)
		expectedMsgChan <- ""

		syncChan <- "A7"
		err = sessionA.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,", filez.ReadStringOrPanic(sessionA.TmpOutName))
		expectedMsgChan <- ""

		syncChan <- "END"
	}()

	// idea: launch printers with delays in a goroutine then launch tailBlocking in another goroutine sync
	go func() {
		err := screenTailer.TailBlocking(expectedSessionA, 2*time.Second)
		assert.NoError(t, err)
	}()

	assert.Empty(t, outW.String())

	var step string
	var expectedOut string
	<-syncChan // Start first print
	for step = <-syncChan; step != "END"; step = <-syncChan {
		// Wait for tailing lag period
		expectedOut += <-expectedMsgChan
		time.Sleep(continuousFlushPeriod*3 + 10*time.Millisecond) // Wait some tailer flush periods
		out := outW.String()
		assert.Equal(t, expectedOut, out, "testing step %s", step)
	}

}

func TestTailBlocking_OutOfOrder(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo8001c"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	expectedSessionA := "barA8001"
	expectedPrinterA10a := "barA10a"
	expectedPrinterA20a := "barA20a"
	expectedSessionB := "barB8001"
	expectedPrinterB10a := "barB10a"
	expectedPrinterB30a := "barB30a"
	expectedSessionC := "barC8001"
	expectedPrinterC10a := "barC10a"
	expectedPrinterC30a := "barC30a"
	expectedPrinterC40a := "barC40a"

	syncChanA := make(chan string)
	syncChanB := make(chan string)
	syncChanC := make(chan string)
	go func() {
		// Wait before printing
		syncChanA <- "startA"

		sessionA, err := screen.Session(expectedSessionA, 12)
		require.NoError(t, err)
		require.NotNil(t, sessionA)
		err = sessionA.Start(100 * time.Millisecond)
		assert.NoError(t, err)
		prtrA10a, err := sessionA.Printer(expectedPrinterA10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrA10a)
		prtrA20a, err := sessionA.Printer(expectedPrinterA20a, 20)
		require.NoError(t, err)
		require.NotNil(t, prtrA20a)

		prtrA10a.Out("A10a1,")
		prtrA20a.Out("A20a1,")
		prtrA20a.Out("A20a2,")
		prtrA10a.Out("A10a2,")

		err = sessionA.ClosePrinter(expectedPrinterA10a, "msg")
		assert.NoError(t, err)
		err = sessionA.ClosePrinter(expectedPrinterA20a, "msg")
		assert.NoError(t, err)

		err = sessionA.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,", filez.ReadStringOrPanic(sessionA.TmpOutName))

		syncChanA <- "endA"
	}()

	go func() {
		// Wait before printing
		syncChanB <- "startB"

		sessionB, err := screen.Session(expectedSessionB, 42)
		require.NoError(t, err)
		require.NotNil(t, sessionB)
		err = sessionB.Start(100 * time.Millisecond)
		assert.NoError(t, err)

		prtrB10a, err := sessionB.Printer(expectedPrinterB10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrB10a)
		prtrB30a, err := sessionB.Printer(expectedPrinterB30a, 30)
		require.NoError(t, err)
		require.NotNil(t, prtrB30a)
		prtrB30a.Out("B30a1,")
		prtrB30a.Out("B30a2,")
		prtrB10a.Out("B10a1,")
		prtrB10a.Out("B10a2,")

		err = sessionB.ClosePrinter(expectedPrinterB30a, "msg")
		assert.NoError(t, err)
		err = sessionB.ClosePrinter(expectedPrinterB10a, "msg")
		assert.NoError(t, err)

		err = sessionB.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "B10a1,B10a2,B30a1,B30a2,", filez.ReadStringOrPanic(sessionB.TmpOutName))

		syncChanB <- "endB"
	}()

	// async prints
	go func() {
		// Wait before printing
		syncChanC <- "startC"

		sessionC, err := screen.Session(expectedSessionC, 42)
		require.NoError(t, err)
		require.NotNil(t, sessionC)
		err = sessionC.Start(100 * time.Millisecond)
		assert.NoError(t, err)
		prtrC10a, err := sessionC.Printer(expectedPrinterC10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrC10a)
		prtrC30a, err := sessionC.Printer(expectedPrinterC30a, 30)
		require.NoError(t, err)
		require.NotNil(t, prtrC30a)
		prtrC40a, err := sessionC.Printer(expectedPrinterC40a, 40)
		require.NoError(t, err)
		require.NotNil(t, prtrC40a)

		prtrC10a.Out("C10a1,")
		err = sessionC.ClosePrinter(expectedPrinterC10a, "msg")
		assert.NoError(t, err)
		prtrC30a.Out("C30a1,")
		prtrC40a.Out("C40a1,")
		err = sessionC.ClosePrinter(expectedPrinterC40a, "msg")
		assert.NoError(t, err)
		err = sessionC.ClosePrinter(expectedPrinterC30a, "msg")
		assert.NoError(t, err)

		err = sessionC.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "C10a1,C30a1,C40a1,", filez.ReadStringOrPanic(sessionC.TmpOutName))

		syncChanC <- "endC"
	}()

	err := screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Empty(t, outW.String())

	// Should Flush B before A despite of priority because 2 sessions will be available simultaneously
	<-syncChanA
	<-syncChanB
	<-syncChanB
	<-syncChanA
	time.Sleep(10 * time.Millisecond)
	err = screenTailer.TailBlocking(expectedSessionB, 3*continuousFlushPeriod+10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "B10a1,B10a2,B30a1,B30a2,"+"A10a1,A10a2,A20a1,A20a2,", outW.String())

	err = screenTailer.TailBlocking(expectedSessionA, 3*continuousFlushPeriod+10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "B10a1,B10a2,B30a1,B30a2,"+"A10a1,A10a2,A20a1,A20a2,", outW.String())

	<-syncChanC
	<-syncChanC
	err = screenTailer.TailBlocking(expectedSessionC, 3*continuousFlushPeriod+10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "B10a1,B10a2,B30a1,B30a2,"+"A10a1,A10a2,A20a1,A20a2,"+"C10a1,C30a1,C40a1,", outW.String())
}

func TestTailAllBlocking_InOrder(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo9001c"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)

	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	expectedSessionA := "barA9001"
	expectedPrinterA10a := "barA10a"
	expectedPrinterA20a := "barA20a"
	expectedSessionB := "barB9001"
	expectedPrinterB10a := "barB10a"
	expectedPrinterB30a := "barB30a"
	expectedSessionC := "barC9001"
	expectedPrinterC10a := "barC10a"
	expectedPrinterC30a := "barC30a"
	expectedPrinterC40a := "barC40a"

	syncChan := make(chan string)
	go func() {
		// Wait before printing

		syncChan <- "startA"

		sessionA, err := screen.Session(expectedSessionA, 12)
		require.NoError(t, err)
		require.NotNil(t, sessionA)
		err = sessionA.Start(100 * time.Millisecond)
		assert.NoError(t, err)
		prtrA10a, err := sessionA.Printer(expectedPrinterA10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrA10a)
		prtrA20a, err := sessionA.Printer(expectedPrinterA20a, 20)
		require.NoError(t, err)
		require.NotNil(t, prtrA20a)

		prtrA10a.Out("A10a1,")
		prtrA10a.Out("A10a2,")
		err = sessionA.ClosePrinter(expectedPrinterA10a, "msg")
		assert.NoError(t, err)

		screen.NotifyPrinter().Out("notif1,")
		err = screen.NotifyPrinter().Flush()
		require.NoError(t, err)

		prtrA20a.Out("A20a1,")
		prtrA20a.Out("A20a2,")
		err = sessionA.ClosePrinter(expectedPrinterA20a, "msg")
		assert.NoError(t, err)
		err = sessionA.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "A10a1,A10a2,A20a1,A20a2,", filez.ReadStringOrPanic(sessionA.TmpOutName))

		syncChan <- "endA+notif1"

		// Wait before printing
		syncChan <- "startB"

		screen.NotifyPrinter().Out("notif2,")
		err = screen.NotifyPrinter().Flush()
		require.NoError(t, err)

		sessionB, err := screen.Session(expectedSessionB, 42)
		require.NoError(t, err)
		require.NotNil(t, sessionB)
		err = sessionB.Start(100 * time.Millisecond)
		assert.NoError(t, err)
		prtrB10a, err := sessionB.Printer(expectedPrinterB10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrB10a)
		prtrB30a, err := sessionB.Printer(expectedPrinterB30a, 30)
		require.NoError(t, err)
		require.NotNil(t, prtrB30a)

		prtrB10a.Out("B10a1,")
		prtrB10a.Out("B10a2,")
		err = sessionB.ClosePrinter(expectedPrinterB10a, "msg")
		assert.NoError(t, err)

		prtrB30a.Out("B30a1,")
		prtrB30a.Out("B30a2,")
		err = sessionB.ClosePrinter(expectedPrinterB30a, "msg")
		assert.NoError(t, err)
		err = sessionB.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "B10a1,B10a2,B30a1,B30a2,", filez.ReadStringOrPanic(sessionB.TmpOutName))

		syncChan <- "endB+notif2"

		// Wait before printing
		syncChan <- "startC"

		screen.NotifyPrinter().Out("notif3,")
		err = screen.NotifyPrinter().Flush()
		require.NoError(t, err)

		sessionC, err := screen.Session(expectedSessionC, 42)
		require.NoError(t, err)
		require.NotNil(t, sessionC)
		err = sessionC.Start(100 * time.Millisecond)
		assert.NoError(t, err)
		prtrC10a, err := sessionC.Printer(expectedPrinterC10a, 10)
		require.NoError(t, err)
		require.NotNil(t, prtrC10a)
		prtrC30a, err := sessionC.Printer(expectedPrinterC30a, 30)
		require.NoError(t, err)
		require.NotNil(t, prtrC30a)
		prtrC40a, err := sessionC.Printer(expectedPrinterC40a, 40)
		require.NoError(t, err)
		require.NotNil(t, prtrC40a)

		prtrC10a.Out("C10a1,")
		err = sessionC.ClosePrinter(expectedPrinterC10a, "msg")
		assert.NoError(t, err)
		prtrC30a.Out("C30a1,")
		err = sessionC.ClosePrinter(expectedPrinterC30a, "msg")
		assert.NoError(t, err)
		prtrC40a.Out("C40a1,")
		err = sessionC.ClosePrinter(expectedPrinterC40a, "msg")
		assert.NoError(t, err)
		err = sessionC.End("msg")
		assert.NoError(t, err)
		assert.Equal(t, "C10a1,C30a1,C40a1,", filez.ReadStringOrPanic(sessionC.TmpOutName))

		syncChan <- "endC"
	}()

	err := screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Empty(t, outW.String())

	err = screenTailer.TailAllBlocking(3*continuousFlushPeriod + 10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "", outW.String())

	<-syncChan
	<-syncChan
	// Should Flush A before B because A is the first session available.
	err = screenTailer.TailAllBlocking(3*continuousFlushPeriod + 10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "notif1,"+"A10a1,A10a2,A20a1,A20a2,", outW.String())

	<-syncChan
	<-syncChan
	err = screenTailer.TailAllBlocking(3*continuousFlushPeriod + 10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "notif1,"+"A10a1,A10a2,A20a1,A20a2,"+"notif2,"+"B10a1,B10a2,B30a1,B30a2,", outW.String())

	<-syncChan
	<-syncChan
	err = screenTailer.TailAllBlocking(3*continuousFlushPeriod + 10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "notif1,"+"A10a1,A10a2,A20a1,A20a2,"+"notif2,"+"B10a1,B10a2,B30a1,B30a2,"+"notif3,"+"C10a1,C30a1,C40a1,", outW.String())
}

func TestAsyncScreen_TailAllBlocking_ClearedSession(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo3212"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir)
	require.NotNil(t, screen)

	expectedSession := "foo3212"
	expectedPrinter := "bar"
	expectedMessage := "baz"

	session, err := screen.Session(expectedSession, 42)
	require.NoError(t, err)
	require.NotNil(t, session)
	err = session.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtr10, err := session.Printer(expectedPrinter, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr10)

	prtr10.Out(expectedMessage)
	err = prtr10.Flush()
	assert.NoError(t, err)

	err = session.Flush()
	assert.NoError(t, err)

	// Clearing a session which was elected but not ended should be cleared

	// Force suite election
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)
	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, outW.String())

	// Then end & clear session
	err = session.End("msg")
	assert.NoError(t, err)
	// err = screen.ClearSession(expectedSession)
	// assert.NoError(t, err)

	// Should not block
	err = screenTailer.TailAllBlocking(10 * time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, outW.String())
}
