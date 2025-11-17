package zcreen

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mxbossard/utilz/printz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
	Allow multiple sink at same time. Each sink must write in separate files.
	Tailers are responsible to concat correctly sink files.
*/

func TestRunningTwoScreensSequentially(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestRunningTwoScreensSequentially"
	require.NoError(t, os.RemoveAll(tmpDir))

	uniqSession1 := "paf47001-uniq1"
	uniqSession2 := "paf47001-uniq2"
	sharedSession := "paf47001-shared"

	// ----- First screen1
	screen1 := NewAsyncScreen(tmpDir, false)
	assert.NotNil(t, screen1)
	assert.DirExists(t, tmpDir)

	u1s, err := screen1.Session(uniqSession1, 0)
	assert.NoError(t, err)
	require.NotNil(t, u1s)
	err = u1s.Start(time.Second)
	assert.NoError(t, err)

	s1s, err := screen1.Session(sharedSession, 0)
	assert.NoError(t, err)
	require.NotNil(t, s1s)
	err = s1s.Start(time.Second)
	assert.NoError(t, err)

	// ----- Tailer
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	// ----- Screen 1 write tests
	// Write on first screen only printer
	u1_prtr_10, err := u1s.Printer("u1", 10)
	assert.NoError(t, err)
	require.NotNil(t, u1_prtr_10)
	assert.NotPanics(t, func() {
		u1_prtr_10.Out("u1_prtr_10\n")
	})

	// Write on shared screen only printer
	s0_prtr_10, err := s1s.Printer("s0", 10)
	assert.NoError(t, err)
	require.NotNil(t, s0_prtr_10)
	assert.NotPanics(t, func() {
		s0_prtr_10.Out("s0_prtr_10\n")
	})

	s0_prtr_30a, err := s1s.Printer("s0", 30)
	assert.NoError(t, err)
	require.NotNil(t, s0_prtr_30a)
	assert.NotPanics(t, func() {
		s0_prtr_30a.Out("s0_prtr_30a\n")
	})

	s0_prtr_50a, err := s1s.Printer("s0", 50)
	assert.NoError(t, err)
	require.NotNil(t, s0_prtr_50a)
	assert.NotPanics(t, func() {
		s0_prtr_50a.Out("s0_prtr_50\n")
	})

	err = screen1.Close()
	require.NoError(t, err)

	// ----- Second screen
	screen2 := NewAsyncScreen(tmpDir, false)
	assert.NotNil(t, screen2)
	assert.DirExists(t, tmpDir)

	u2s, err := screen2.Session(uniqSession2, 0)
	assert.NoError(t, err)
	require.NotNil(t, u2s)
	err = u2s.Start(time.Second)
	assert.NoError(t, err)

	s2s, err := screen2.Session(sharedSession, 0)
	assert.NoError(t, err)
	require.NotNil(t, s2s)
	err = s2s.Start(time.Second)
	assert.NoError(t, err)

	// ----- Screen 2 write tests
	// Write on second screen only printer
	u2_prtr_10, err := u2s.Printer("u2", 10)
	assert.NoError(t, err)
	require.NotNil(t, u2_prtr_10)

	assert.NotPanics(t, func() {
		u2_prtr_10.Out("u2_prtr_10\n")
	})

	// Write on shared screen only printer
	s0_prtr_30b, err := s2s.Printer("s0", 30)
	assert.NoError(t, err)
	require.NotNil(t, s0_prtr_30b)
	assert.NotPanics(t, func() {
		s0_prtr_30b.Out("s0_prtr_30b\n")
	})

	s0_prtr_20, err := s2s.Printer("s0", 20)
	assert.NoError(t, err)
	require.NotNil(t, s0_prtr_20)
	assert.NotPanics(t, func() {
		s0_prtr_20.Out("s0_prtr_20\n")
	})

	s0_prtr_40, err := s2s.Printer("s0", 40)
	assert.NoError(t, err)
	require.NotNil(t, s0_prtr_40)
	assert.NotPanics(t, func() {
		s0_prtr_40.Out("s0_prtr_40\n")
	})

	err = screen2.Close()
	require.NoError(t, err)

	// ----- Tailing suite by suite

	// Check First screen content
	outW.Reset()
	errW.Reset()
	err = screenTailer.TailOnlyBlocking(uniqSession1, 2*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "u1_prtr_10\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Check Second screen content
	outW.Reset()
	errW.Reset()
	err = screenTailer.TailOnlyBlocking(uniqSession2, 2*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "u2_prtr_10\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Check Shared screen content
	outW.Reset()
	errW.Reset()
	err = screenTailer.TailOnlyBlocking(sharedSession, 2*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "s0_prtr_10\ns0_prtr_20\ns0_prtr_30a\ns0_prtr_30b\ns0_prtr_40\ns0_prtr_50\n", outW.String())
	assert.Equal(t, "", errW.String())

}

func TestRunningTwoScreensConcurrently(t *testing.T) {
	tmpDir := "/tmp/utilz.zcreen.TestRunningTwoScreensConcurrently"
	require.NoError(t, os.RemoveAll(tmpDir))

	uniqSession1 := "paf47002-uniq1"
	uniqSession2 := "paf47002-uniq2"
	sharedSession := "paf47002-shared"

	// ----- First screen1
	screen1 := NewAsyncScreen(tmpDir, false)
	assert.NotNil(t, screen1)
	assert.DirExists(t, tmpDir)

	u1s, err := screen1.Session(uniqSession1, 0)
	assert.NoError(t, err)
	require.NotNil(t, u1s)
	err = u1s.Start(time.Second)
	assert.NoError(t, err)

	s1s, err := screen1.Session(sharedSession, 0)
	assert.NoError(t, err)
	require.NotNil(t, s1s)
	err = s1s.Start(time.Second)
	assert.NoError(t, err)

	// ----- Tailer
	outW := &strings.Builder{}
	errW := &strings.Builder{}
	outs := printz.NewOutputs(outW, errW)
	screenTailer := NewAsyncScreenTailer(outs, tmpDir)

	// ----- Second screen
	screen2 := NewAsyncScreen(tmpDir, false)
	assert.NotNil(t, screen2)
	assert.DirExists(t, tmpDir)

	u2s, err := screen2.Session(uniqSession2, 0)
	assert.NoError(t, err)
	require.NotNil(t, u2s)
	err = u2s.Start(time.Second)
	assert.NoError(t, err)

	s2s, err := screen2.Session(sharedSession, 0)
	assert.NoError(t, err)
	require.NotNil(t, s2s)
	err = s2s.Start(time.Second)
	assert.NoError(t, err)

	// ----- Screen 1 write tests
	// Write on first screen only printer
	u1_prtr_10, err := u1s.Printer("u1", 10)
	assert.NoError(t, err)
	require.NotNil(t, u1_prtr_10)
	assert.NotPanics(t, func() {
		u1_prtr_10.Out("u1_prtr_10\n")
	})

	// Write on shared screen only printer
	s0_prtr_10, err := s1s.Printer("s0", 10)
	assert.NoError(t, err)
	require.NotNil(t, s0_prtr_10)
	assert.NotPanics(t, func() {
		s0_prtr_10.Out("s0_prtr_10\n")
	})

	s0_prtr_30a, err := s1s.Printer("s0", 30)
	assert.NoError(t, err)
	require.NotNil(t, s0_prtr_30a)
	assert.NotPanics(t, func() {
		s0_prtr_30a.Out("s0_prtr_30a\n")
	})

	s0_prtr_50a, err := s1s.Printer("s0", 50)
	assert.NoError(t, err)
	require.NotNil(t, s0_prtr_50a)
	assert.NotPanics(t, func() {
		s0_prtr_50a.Out("s0_prtr_50\n")
	})

	// err = screen1.Close()
	// require.NoError(t, err)

	// ----- Screen 2 write tests
	// Write on second screen only printer
	u2_prtr_10, err := u2s.Printer("u2", 10)
	assert.NoError(t, err)
	require.NotNil(t, u2_prtr_10)

	assert.NotPanics(t, func() {
		u2_prtr_10.Out("u2_prtr_10\n")
	})

	// Write on shared screen only printer
	s0_prtr_30b, err := s2s.Printer("s0", 30)
	assert.NoError(t, err)
	require.NotNil(t, s0_prtr_30b)
	assert.NotPanics(t, func() {
		s0_prtr_30b.Out("s0_prtr_30b\n")
	})

	s0_prtr_20, err := s2s.Printer("s0", 20)
	assert.NoError(t, err)
	require.NotNil(t, s0_prtr_20)
	assert.NotPanics(t, func() {
		s0_prtr_20.Out("s0_prtr_20\n")
	})

	s0_prtr_40, err := s2s.Printer("s0", 40)
	assert.NoError(t, err)
	require.NotNil(t, s0_prtr_40)
	assert.NotPanics(t, func() {
		s0_prtr_40.Out("s0_prtr_40\n")
	})

	// err = screen2.Close()
	// require.NoError(t, err)

	u1s.End("pif")
	u2s.End("paf")
	s1s.End("foo")
	s2s.End("bar")

	// ----- Tailing suite by suite

	// Check First screen content
	outW.Reset()
	errW.Reset()
	err = screenTailer.TailOnlyBlocking(uniqSession1, 2*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "u1_prtr_10\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Check Second screen content
	outW.Reset()
	errW.Reset()
	err = screenTailer.TailOnlyBlocking(uniqSession2, 2*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "u2_prtr_10\n", outW.String())
	assert.Equal(t, "", errW.String())

	// Check Shared screen content
	outW.Reset()
	errW.Reset()
	err = screenTailer.TailOnlyBlocking(sharedSession, 2*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "s0_prtr_10\ns0_prtr_20\ns0_prtr_30a\ns0_prtr_30b\ns0_prtr_40\ns0_prtr_50\n", outW.String())
	assert.Equal(t, "", errW.String())
}
