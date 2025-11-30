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
	session, err := s.Session("bar1001", 42)
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.DirExists(t, tmpDir)
	assert.DirExists(t, tmpDir+"/"+sessionDirPrefix+"bar1001")

	err = session.Start(100 * time.Millisecond)
	require.NoError(t, err)
	expectedSessionDir := filepath.Join(tmpDir, sessionDirPrefix+"bar1001")
	assert.DirExists(t, expectedSessionDir)
	matches, err := filepath.Glob(expectedSessionDir + "/bar1001" + outFileNameSuffix + "*")
	require.NoError(t, err)
	require.Len(t, matches, 1)
	matches, err = filepath.Glob(expectedSessionDir + "/bar1001" + errFileNameSuffix + "*")
	require.NoError(t, err)
	require.Len(t, matches, 1)
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

	sessionSerFilepath := filepath.Join(tmpDir, expectedSession+serializedExtension)
	sessionDirTmpFilepath := filepath.Join(tmpDir, sessionDirPrefix+expectedSession)
	printersDirTmpFilepath := printersDirPath(sessionDirTmpFilepath)

	sessionTmpOutFilepath := func() string {
		matches := tmpFilenameMatches(sessionDirTmpFilepath, expectedSession, outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	sessionTmpErrFilepath := func() string {
		matches := tmpFilenameMatches(sessionDirTmpFilepath, expectedSession, errFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpOutFilepath := func() string {
		matches := priorizedTmpFilenameMatches(printersDirTmpFilepath, expectedPrinter, outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpErrFilepath := func() string {
		matches := priorizedTmpFilenameMatches(printersDirTmpFilepath, expectedPrinter, errFileNameSuffix)
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

	assert.NotEmpty(t, func() string { s, _ := filez.ReadString(sessionSerFilepath); return s }())
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
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(printerTmpErrFilepath))
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

	sessionSerFilepath := filepath.Join(tmpDir, expectedSession+serializedExtension)
	sessionDirTmpFilepath := filepath.Join(tmpDir, sessionDirPrefix+expectedSession)
	printersDirTmpFilepath := printersDirPath(sessionDirTmpFilepath)
	sessionTmpOutFilepath := func() string {
		matches := tmpFilenameMatches(sessionDirTmpFilepath, expectedSession, outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	sessionTmpErrFilepath := func() string {
		matches := tmpFilenameMatches(sessionDirTmpFilepath, expectedSession, errFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpOutFilepath := func() string {
		matches := priorizedTmpFilenameMatches(printersDirTmpFilepath, expectedPrinter, outFileNameSuffix)
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	printerTmpErrFilepath := func() string {
		matches := priorizedTmpFilenameMatches(printersDirTmpFilepath, expectedPrinter, errFileNameSuffix)
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

	assert.NotEmpty(t, func() string { s, _ := filez.ReadString(sessionSerFilepath); return s }())
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
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpOutFilepath))
	assert.Empty(t, filez.ReadStringOrPanic(sessionTmpErrFilepath))
	assert.Equal(t, expectedOutMessage, filez.ReadStringOrPanic(printerTmpOutFilepath))
	assert.Equal(t, expectedErrMessage, filez.ReadStringOrPanic(printerTmpErrFilepath))
}
