package zcreen

import (
	"os"
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

func TestScreen_Get(t *testing.T) {
	t.Skip()
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	s := NewScreen(outs)
	assert.NotNil(t, s)
	assert.Implements(t, (*Sink)(nil), s)
}

func TestScreen_GetAsyncScreen(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo40b"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir, true)
	assert.NotNil(t, s)
	assert.DirExists(t, tmpDir)

	require.NotPanics(t, func() {
		duplicate := NewAsyncScreen(tmpDir, false)
		assert.NotNil(t, duplicate)
	})
}

func TestScreen_GetSession(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo1001"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir, false)
	require.NotNil(t, s)
	expectedSessionDir := forgeSessionDirPath(tmpDir, "bar1001", 42)
	session, err := s.Session("bar1001", 42)
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.DirExists(t, tmpDir)
	assert.DirExists(t, expectedSessionDir)

	err = session.Start(100 * time.Millisecond)
	require.NoError(t, err)
	assert.DirExists(t, expectedSessionDir)
}

func TestScreen_GetPrinter(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo2001"
	require.NoError(t, os.RemoveAll(tmpDir))
	s := NewAsyncScreen(tmpDir, false)
	require.NotNil(t, s)
	session, err := s.Session("foo2001", 42)
	require.NoError(t, err)
	require.NotNil(t, session)
	session.Start(100 * time.Millisecond)
	prtr, err := session.Printer("bar", 10)
	require.NoError(t, err)
	assert.NotNil(t, prtr)
}

func TestScreen_BasicOut(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo3001"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir, false)
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

	sessionDirFilepath := forgeSessionDirPath(tmpDir, expectedSession, 42)
	sessionSerFilepath := sessionSerializedPath(sessionDirFilepath)
	printersDirFilepath := printersDirPath(sessionDirFilepath)

	matches := priorizedPrinterFilenameMatches(printersDirFilepath, expectedPrinter, outFileQualifier)
	assert.Empty(t, matches)
	matches = priorizedPrinterFilenameMatches(printersDirFilepath, expectedPrinter, errFileQualifier)
	assert.Empty(t, matches)

	require.DirExists(t, tmpDir)
	require.DirExists(t, sessionDirFilepath)
	assert.FileExists(t, sessionSerFilepath)

	assert.NotEmpty(t, func() string { s, _ := filez.ReadString(sessionSerFilepath); return s }())
	ser, err := deserializeSession(sessionSerFilepath)
	require.NoError(t, err)
	assert.NotNil(t, ser)

	prtr10.Out(expectedMessage)

	err = prtr10.Flush()
	assert.NoError(t, err)

	printerTmpOutFilepath := func() string {
		matches := priorizedPrinterFilenameMatches(printersDirFilepath, expectedPrinter, outFileQualifier)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	matches = priorizedPrinterFilenameMatches(printersDirFilepath, expectedPrinter, errFileQualifier)
	require.Empty(t, matches)

	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
}

func TestScreen_ClearSession(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo3002"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir, false)
	require.NotNil(t, screen)

	expectedSession := "foo3002"
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

	err = session.End("msg")
	assert.NoError(t, err)
}

func TestScreen_BasicOutAndErr(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo3003"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir, false)
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

	sessionDirFilepath := forgeSessionDirPath(tmpDir, expectedSession, 42)
	sessionSerFilepath := sessionSerializedPath(sessionDirFilepath)
	printersDirFilepath := printersDirPath(sessionDirFilepath)
	matches := priorizedPrinterFilenameMatches(printersDirFilepath, expectedPrinter, outFileQualifier)
	assert.Empty(t, matches)
	matches = priorizedPrinterFilenameMatches(printersDirFilepath, expectedPrinter, errFileQualifier)
	assert.Empty(t, matches)

	require.DirExists(t, tmpDir)
	require.DirExists(t, sessionDirFilepath)
	assert.FileExists(t, sessionSerFilepath)

	assert.NotEmpty(t, func() string { s, _ := filez.ReadString(sessionSerFilepath); return s }())
	ser, err := deserializeSession(sessionSerFilepath)
	require.NoError(t, err)
	assert.NotNil(t, ser)

	prtr10.Out(expectedOutMessage)
	prtr10.Err(expectedErrMessage)

	err = prtr10.Flush()
	assert.NoError(t, err)

	printerTmpOutFilepath := func() string {
		matches := priorizedPrinterFilenameMatches(printersDirFilepath, expectedPrinter, outFileQualifier)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpErrFilepath := func() string {
		matches := priorizedPrinterFilenameMatches(printersDirFilepath, expectedPrinter, errFileQualifier)
		require.NotEmpty(t, matches)
		return matches[0]
	}()

	assert.Equal(t, expectedOutMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Equal(t, expectedErrMessage, filez.ReadStringOrPanic(printerTmpErrFilepath))

	err = session.Flush()
	assert.NoError(t, err)
	assert.Equal(t, expectedOutMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Equal(t, expectedErrMessage, filez.ReadStringOrPanic(printerTmpErrFilepath))
}

func TestScreen_Resync(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.foo3004"
	require.NoError(t, os.RemoveAll(tmpDir))

	expectedSession1 := "foo30041"
	expectedPrinter1 := "bar1"
	expectedMessage1 := "baz1"
	expectedMessage1b := "baz1b"

	expectedSession2 := "foo30042"
	expectedPrinter2 := "bar2"
	expectedMessage2 := "baz2"

	expectedSession3 := "foo30043"
	expectedPrinter3 := "bar3"
	expectedMessage3 := "baz3"

	// build a first screen1 with session & printer
	screen1 := NewAsyncScreen(tmpDir, false)
	require.NotNil(t, screen1)
	session1, err := screen1.Session(expectedSession1, 42)
	require.NoError(t, err)
	require.NotNil(t, session1)
	err = session1.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtr110, err := session1.Printer(expectedPrinter1, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr110)
	prtr110.Out(expectedMessage1)

	session3, err := screen1.Session(expectedSession3, 42)
	require.NoError(t, err)
	require.NotNil(t, session3)
	err = session3.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtr310, err := session3.Printer(expectedPrinter3, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr310)
	prtr310.Out(expectedMessage3)
	session3.End("end3")

	screen1.Close()

	// build a second screen with a new session
	screen2 := NewAsyncScreen(tmpDir, false)
	require.NotNil(t, screen2)
	err = screen2.Resync()
	assert.NoError(t, err)
	session2, err := screen2.Session(expectedSession2, 42)
	require.NoError(t, err)
	require.NotNil(t, session2)
	err = session2.Start(100 * time.Millisecond)
	assert.NoError(t, err)
	prtr210, err := session2.Printer(expectedPrinter2, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr210)
	prtr210.Out(expectedMessage2)

	// Use an already existing session & printer
	session1b, err := screen2.Session(expectedSession1, 42)
	require.NoError(t, err)
	require.NotNil(t, session1b)
	prtr110b, err := session2.Printer(expectedPrinter1, 10)
	require.NoError(t, err)
	require.NotNil(t, prtr210)
	prtr110b.Out(expectedMessage1b)

	// Use an already existing and ended session
	session3b, err := screen2.Session(expectedSession3, 42)
	require.NoError(t, err)
	require.NotNil(t, session3b)
	_, err = session3b.Printer(expectedPrinter3, 10)
	assert.Error(t, err)
}
