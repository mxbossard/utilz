package zcreen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/printz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
	Disallow multiple Sink at same time => fileLock
	If a sink is inactive for a long period => relase the lock
	Add a ticker which update sink activity / lock
*/

func TestBuildAndCloseUniqScreen(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestBuildAndCloseUniqScreen"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen := NewAsyncScreen(tmpDir, false)
	assert.NotNil(t, screen)
	assert.DirExists(t, tmpDir)

	expectedSession := "foo21001"

	assert.Panics(t, func() {
		NewAsyncScreen(tmpDir, false)
	}, "Until closed it should be impossible to build another instance of screen")

	session, err := screen.Session(expectedSession, 0)
	assert.NoError(t, err)
	require.NotNil(t, session)
	err = session.Start(time.Second)
	assert.NoError(t, err)
	prtr, err := session.Printer("bar", 0)
	assert.NoError(t, err)
	require.NotNil(t, prtr)

	assert.NotPanics(t, func() {
		prtr.Out("msg")
	})

	err = screen.Close()
	assert.NoError(t, err)

	assert.Panics(t, func() {
		prtr.Out("msg")
	}, "Printing in a closed screen should panic")

	assert.NotPanics(t, func() {
		NewAsyncScreen(tmpDir, false)
	}, "Once closed it should be possible to build another instance of screen")

	_, err = screen.Session("foo", 0)
	assert.Error(t, err, "Closed screen should not allow session creation")
}

func TestAsyncScreen_BasicOut_CleanRestart(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestAsyncScreen_BasicOut_CleanRestart"
	require.NoError(t, os.RemoveAll(tmpDir))
	screen1 := NewAsyncScreen(tmpDir, false)
	require.NotNil(t, screen1)

	expectedSession := "foo22001"
	expectedPrinter := "bar"
	expectedMessage1 := "pif"
	expectedMessage2 := "paf"

	// Usage of first screen
	session, err := screen1.Session(expectedSession, 42)
	assert.NoError(t, err)
	require.NotNil(t, session)
	err = session.Start(time.Second)
	assert.NoError(t, err)
	prtr1, err := session.Printer(expectedPrinter, 10)
	assert.NoError(t, err)
	require.NotNil(t, prtr1)

	prtr1.Out(expectedMessage1)

	// Close screen should flush printers automatically on Close()
	err = screen1.Close()
	assert.NoError(t, err)

	sessionDirTmpFilepath := filepath.Join(tmpDir, sessionDirPrefix+expectedSession)
	sessionTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedSession + outFileNameSuffix + "*")
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	assert.FileExists(t, sessionTmpOutFilepath)
	assert.Equal(t, expectedMessage1, func() string { s, _ := filez.ReadString(sessionTmpOutFilepath); return s }())

	var screen2 *screen
	// Restart screen
	assert.NotPanics(t, func() {
		screen2 = NewAsyncScreen(tmpDir, false)
	}, "Once closed it should be possible to build another instance of screen")
	require.NotNil(t, screen2)

	// Usage of second screen
	session2, err := screen2.Session(expectedSession, 42)
	assert.NoError(t, err)
	require.NotNil(t, session2)
	// Session MUST be reopened
	err = session2.Start(time.Second)
	assert.NoError(t, err)
	// prtr2 MUST append next to prtr1
	prtr2, err := session2.Printer(expectedPrinter, 10)
	assert.NoError(t, err)
	require.NotNil(t, prtr2)

	prtr2.Out(expectedMessage2)

	err = screen2.Close()
	assert.NoError(t, err)

	// Verify prtr1 & prtr2 where persisted in printer backing file
	session2TmpOutFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedSession + outFileNameSuffix + "*")
		sort.Strings(matches)
		require.NotEmpty(t, matches)
		return matches[1]
	}()
	assert.Equal(t, expectedMessage2, func() string { s, _ := filez.ReadString(session2TmpOutFilepath); return s }())

	// Test tailing zcreen outputs both  messages concatenated
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.NotEmpty(t, outW.String())
	assert.Equal(t, expectedMessage1+expectedMessage2, outW.String())
}

func TestAsyncScreen_BasicOut_DirtyRestart(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestAsyncScreen_BasicOut_DirtyRestart"
	require.NoError(t, os.RemoveAll(tmpDir))

	expectedSession := "foo23001"
	expectedPrinter := "bar"
	expectedMessage1 := "pif"
	expectedMessage2 := "paf"

	syncChan := make(chan bool)
	go func() {
		screen1 := NewAsyncScreen(tmpDir, false)
		require.NotNil(t, screen1)

		// Usage of first screen
		session, err := screen1.Session(expectedSession, 42)
		assert.NoError(t, err)
		require.NotNil(t, session)
		err = session.Start(time.Second)
		assert.NoError(t, err)
		prtr1, err := session.Printer(expectedPrinter, 10)
		assert.NoError(t, err)
		require.NotNil(t, prtr1)

		prtr1.Out(expectedMessage1)

		// Do not close screen
		err = session.Flush()
		assert.NoError(t, err)
		//err = screen1.Close()
		//assert.NoError(t, err)
		syncChan <- true
	}()

	<-syncChan

	sessionDirTmpFilepath := filepath.Join(tmpDir, sessionDirPrefix+expectedSession)
	sessionTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedSession + outFileNameSuffix + "*")
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	assert.FileExists(t, sessionTmpOutFilepath)
	assert.Equal(t, expectedMessage1, func() string { s, _ := filez.ReadString(sessionTmpOutFilepath); return s }())

	// Wait lock release
	// time.Sleep(fileLockingTimeout*2 + 100*time.Millisecond)
	//time.Sleep(100 * time.Millisecond)

	var screen2 *screen
	// Restart screen
	assert.NotPanics(t, func() {
		screen2 = NewAsyncScreen(tmpDir, true)
	}, "Once closed it should be possible to build another instance of screen")
	require.NotNil(t, screen2)

	// Usage of second screen
	session2, err := screen2.Session(expectedSession, 42)
	assert.NoError(t, err)
	require.NotNil(t, session2)
	// Session MUST be reopened
	err = session2.Start(time.Second)
	assert.NoError(t, err)
	// prtr2 MUST append next to prtr1
	prtr2, err := session2.Printer(expectedPrinter, 10)
	assert.NoError(t, err)
	require.NotNil(t, prtr2)

	prtr2.Out(expectedMessage2)

	err = screen2.Close()
	assert.NoError(t, err)

	// Verify prtr1 & prtr2 where persisted in printer backing file
	session2TmpOutFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedSession + outFileNameSuffix + "*")
		sort.Strings(matches)
		require.NotEmpty(t, matches)
		return matches[1]
	}()
	assert.Equal(t, expectedMessage2, func() string { s, _ := filez.ReadString(session2TmpOutFilepath); return s }())

	// Test tailing zcreen outputs both  messages concatenated
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	err = screenTailer.tailAll()
	assert.NoError(t, err)
	assert.NotEmpty(t, outW.String())
	assert.Equal(t, expectedMessage1+expectedMessage2, outW.String())
}

func TestAsyncScreen_BasicOut_TailingBeforeCleanRestart(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestAsyncScreen_BasicOut_TailingBeforeDirtyRestart"
	require.NoError(t, os.RemoveAll(tmpDir))

	expectedSession := "foo24001"
	expectedPrinter := "bar"
	expectedMessage1 := "pif"
	expectedMessage2 := "paf"

	// Test tailing zcreen outputs both  messages concatenated
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)

	// Start tailer async
	syncChan := make(chan bool)
	go func() {
		screenTailer := NewAsyncScreenTailerWaiting(outs, tmpDir, 2*time.Second)
		err := screenTailer.TailSuppliedBlocking([]string{expectedSession}, 5000*time.Millisecond)
		assert.NoError(t, err)
		syncChan <- true
	}()

	fmt.Printf("building async screen ...\n")
	screen1 := NewAsyncScreen(tmpDir, true)
	require.NotNil(t, screen1)
	fmt.Printf("async screen built.\n")

	// Usage of first screen
	session, err := screen1.Session(expectedSession, 42)
	assert.NoError(t, err)
	require.NotNil(t, session)
	err = session.Start(5 * time.Second)
	assert.NoError(t, err)
	prtr1, err := session.Printer(expectedPrinter, 10)
	assert.NoError(t, err)
	require.NotNil(t, prtr1)

	prtr1.Out(expectedMessage1)

	// Do not close screen
	err = session.Flush()
	assert.NoError(t, err)
	//err = screen1.Close()
	//assert.NoError(t, err)

	sessionDirTmpFilepath := filepath.Join(tmpDir, sessionDirPrefix+expectedSession)
	sessionTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedSession + outFileNameSuffix + "*")
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	assert.FileExists(t, sessionTmpOutFilepath)
	assert.Equal(t, expectedMessage1, func() string { s, _ := filez.ReadString(sessionTmpOutFilepath); return s }())

	err = screen1.Close()
	assert.NoError(t, err)

	fmt.Print("screen 1 closed\n")

	// Restart screen (build a new screen)
	time.Sleep(1 * time.Second)

	var screen2 *screen
	assert.NotPanics(t, func() {
		screen2 = NewAsyncScreen(tmpDir, true)
	}, "Once closed it should be possible to build another instance of screen")
	require.NotNil(t, screen2)

	// Usage of second screen
	session2, err := screen2.Session(expectedSession, 42)
	assert.NoError(t, err)
	require.NotNil(t, session2)
	// Session MUST be reopened
	err = session2.Start(time.Second)
	assert.NoError(t, err)
	// prtr2 MUST append next to prtr1
	prtr2, err := session2.Printer(expectedPrinter, 10)
	assert.NoError(t, err)
	require.NotNil(t, prtr2)

	prtr2.Out(expectedMessage2)

	fmt.Print("will end session\n")
	err = session2.End("session2 end message")
	assert.NoError(t, err)
	fmt.Print("session ended\n")

	// err = screen2.Close()
	// assert.NoError(t, err)

	// Wait TailSuppliedBlocking() to finish it's works
	<-syncChan

	// Verify prtr1 & prtr2 where persisted in printer backing file
	session2TmpOutFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedSession + outFileNameSuffix + "*")
		sort.Strings(matches)
		require.NotEmpty(t, matches)
		return matches[1]
	}()
	assert.Equal(t, expectedMessage2, func() string { s, _ := filez.ReadString(session2TmpOutFilepath); return s }())

	// FIXME: wait tailer flushing period
	//time.Sleep(100 * time.Millisecond)

	// err = screenTailer.tailAll()
	// assert.NoError(t, err)
	assert.NotEmpty(t, outW.String())
	assert.Equal(t, expectedMessage1+expectedMessage2, outW.String())
}

func TestAsyncScreen_BasicOut_TailingBeforeDirtyRestart(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestAsyncScreen_BasicOut_TailingBeforeDirtyRestart"
	require.NoError(t, os.RemoveAll(tmpDir))

	expectedSession := "foo24001"
	expectedPrinter := "bar"
	expectedMessage1 := "pif"
	expectedMessage2 := "paf"

	// Test tailing zcreen outputs both  messages concatenated
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)

	// Start tailer async
	syncChan := make(chan bool)
	go func() {
		// time.Sleep(100 * time.Millisecond)
		fmt.Printf("building async tailer ...\n")
		screenTailer := NewAsyncScreenTailerWaiting(outs, tmpDir, 1*time.Second)
		fmt.Printf("async tailer built.\n")
		<-syncChan
		err := screenTailer.TailSuppliedBlocking([]string{expectedSession}, 5000*time.Millisecond)
		assert.NoError(t, err)
		syncChan <- true
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Printf("building async screen ...\n")
	screen1 := NewAsyncScreen(tmpDir, true)
	require.NotNil(t, screen1)
	require.DirExists(t, tmpDir)
	fmt.Printf("async screen built.\n")
	syncChan <- true

	// Usage of first screen
	session, err := screen1.Session(expectedSession, 42)
	assert.NoError(t, err)
	require.NotNil(t, session)
	err = session.Start(time.Second)
	assert.NoError(t, err)
	prtr1, err := session.Printer(expectedPrinter, 10)
	assert.NoError(t, err)
	require.NotNil(t, prtr1)

	prtr1.Out(expectedMessage1)

	// Do not close screen
	err = session.Flush()
	assert.NoError(t, err)
	//err = screen1.Close()
	//assert.NoError(t, err)

	sessionDirTmpFilepath := filepath.Join(tmpDir, sessionDirPrefix+expectedSession)
	sessionTmpOutFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedSession + outFileNameSuffix + "*")
		require.NotEmpty(t, matches)
		return matches[0]
	}()
	assert.FileExists(t, sessionTmpOutFilepath)
	assert.Equal(t, expectedMessage1, func() string { s, _ := filez.ReadString(sessionTmpOutFilepath); return s }())

	// Restart screen (build a new screen)

	var screen2 *screen
	assert.NotPanics(t, func() {
		screen2 = NewAsyncScreen(tmpDir, true)
	}, "Once closed it should be possible to build another instance of screen")
	require.NotNil(t, screen2)

	// Usage of second screen
	session2, err := screen2.Session(expectedSession, 42)
	assert.NoError(t, err)
	require.NotNil(t, session2)
	// Session MUST be reopened
	err = session2.Start(time.Second)
	assert.NoError(t, err)
	// prtr2 MUST append next to prtr1
	prtr2, err := session2.Printer(expectedPrinter, 10)
	assert.NoError(t, err)
	require.NotNil(t, prtr2)

	prtr2.Out(expectedMessage2)

	err = session2.End("session2 end message")
	assert.NoError(t, err)

	// err = screen2.Close()
	// assert.NoError(t, err)

	// Wait TailSuppliedBlocking() to finish it's works
	<-syncChan

	// Verify prtr1 & prtr2 where persisted in printer backing file
	session2TmpOutFilepath := func() string {
		matches, _ := filepath.Glob(sessionDirTmpFilepath + "/" + expectedSession + outFileNameSuffix + "*")
		sort.Strings(matches)
		require.NotEmpty(t, matches)
		return matches[1]
	}()
	assert.Equal(t, expectedMessage2, func() string { s, _ := filez.ReadString(session2TmpOutFilepath); return s }())

	// FIXME: wait tailer flushing period
	//time.Sleep(100 * time.Millisecond)

	// err = screenTailer.tailAll()
	// assert.NoError(t, err)
	assert.NotEmpty(t, outW.String())
	assert.Equal(t, expectedMessage1+expectedMessage2, outW.String())
}
