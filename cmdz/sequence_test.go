package cmdz

import (
	//"fmt"
	//"io"
	//"context"
	//"log"
	//"os/exec"
	//"strings"
	"testing"
	"time"

	//"mby.fr/utils/promise"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	e1 = Cmd("echo", "foo")
	e2 = Cmd("echo", "bar")
	e3 = Cmd("echo", "baz")

	sleep10ms  = Cmd("sleep", "0.01")
	sleep11ms  = Cmd("sleep", "0.011")
	sleep100ms = Cmd("sleep", "0.1")
	sleep200ms = Cmd("sleep", "0.2")
)

func TestSerial(t *testing.T) {
	e1.reset()
	e2.reset()
	e3.reset()

	s := Serial(e1)
	assert.Equal(t, "echo foo", s.String())

	rc1, err1 := s.BlockRun()
	require.NoError(t, err1)
	assert.Equal(t, 0, rc1)
	assert.Equal(t, "foo\n", e1.StdoutRecord())
	assert.Equal(t, "", e2.StdoutRecord())
	assert.Equal(t, "", e3.StdoutRecord())
	assert.Equal(t, "foo\n", s.StdoutRecord())
	assert.Equal(t, "", s.StderrRecord())

	s2 := Serial(e1, e2)
	assert.Equal(t, "echo foo\necho bar", s2.String())

	rc2, err2 := s2.BlockRun()
	require.NoError(t, err2)
	assert.Equal(t, 0, rc2)
	assert.Equal(t, "foo\n", e1.StdoutRecord())
	assert.Equal(t, "bar\n", e2.StdoutRecord())
	assert.Equal(t, "", e3.StdoutRecord())
	assert.Equal(t, "foo\nbar\n", s2.StdoutRecord())
	assert.Equal(t, "", s2.StderrRecord())

	s2.Add(e3)
	assert.Equal(t, "echo foo\necho bar\necho baz", s2.String())

	rc3, err3 := s2.BlockRun()
	require.NoError(t, err3)
	assert.Equal(t, 0, rc3)
	assert.Equal(t, "foo\n", e1.StdoutRecord())
	assert.Equal(t, "bar\n", e2.StdoutRecord())
	assert.Equal(t, "baz\n", e3.StdoutRecord())
	assert.Equal(t, "foo\nbar\nbaz\n", s2.StdoutRecord())
	assert.Equal(t, "", s2.StderrRecord())

	// Test serial timings
	s2.Add(sleep10ms)
	start := time.Now()
	_, err := s2.BlockRun()
	duration := time.Since(start).Microseconds()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, duration, int64(10000), "Serial too quick !")
	assert.Less(t, duration, int64(20000), "Serial too slow !")
	assert.Equal(t, "foo\nbar\nbaz\n", s2.StdoutRecord())
	assert.Equal(t, "", s2.StderrRecord())

	s2.Add(sleep10ms)
	start = time.Now()
	_, err = s2.BlockRun()
	duration = time.Since(start).Microseconds()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, duration, int64(20000), "Serial too quick !")
	assert.Less(t, duration, int64(30000), "Serial too slow !")
	assert.Equal(t, "foo\nbar\nbaz\n", s2.StdoutRecord())
	assert.Equal(t, "", s2.StderrRecord())

	s2.Add(sleep100ms)
	start = time.Now()
	_, err = s2.BlockRun()
	duration = time.Since(start).Microseconds()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, duration, int64(120000), "Serial too quick !")
	assert.Less(t, duration, int64(150000), "Serial too slow !")
	assert.Equal(t, "foo\nbar\nbaz\n", s2.StdoutRecord())
	assert.Equal(t, "", s2.StderrRecord())
}

func TestParallel(t *testing.T) {
	e1.reset()
	e2.reset()
	e3.reset()

	p := Parallel(e1)
	assert.Equal(t, "echo foo", p.String())

	rc, err := p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo\n", e1.StdoutRecord())
	assert.Equal(t, "", e2.StdoutRecord())
	assert.Equal(t, "", e3.StdoutRecord())
	assert.Equal(t, "foo\n", p.StdoutRecord())
	assert.Equal(t, "", p.StderrRecord())

	p2 := Parallel(e1, e2)
	assert.Equal(t, "echo foo\necho bar", p2.String())

	rc2, err2 := p2.BlockRun()
	require.NoError(t, err2)
	assert.Equal(t, 0, rc2)
	assert.Equal(t, "foo\n", e1.StdoutRecord())
	assert.Equal(t, "bar\n", e2.StdoutRecord())
	assert.Equal(t, "", e3.StdoutRecord())
	assert.Equal(t, "foo\nbar\n", p2.StdoutRecord())
	assert.Equal(t, "", p2.StderrRecord())

	p2.Add(e3)
	assert.Equal(t, "echo foo\necho bar\necho baz", p2.String())

	rc3, err3 := p2.BlockRun()
	require.NoError(t, err3)
	assert.Equal(t, 0, rc3)
	assert.Equal(t, "foo\n", e1.StdoutRecord())
	assert.Equal(t, "bar\n", e2.StdoutRecord())
	assert.Equal(t, "baz\n", e3.StdoutRecord())
	assert.Equal(t, "foo\nbar\nbaz\n", p2.StdoutRecord())
	assert.Equal(t, "", p2.StderrRecord())

	// Test serial timings
	p2.Add(sleep10ms)
	start := time.Now()
	_, err = p2.BlockRun()
	duration := time.Since(start).Microseconds()
	require.NoError(t, err)
	//assert.Len(t, 4, rc)
	assert.GreaterOrEqual(t, duration, int64(10000), "Parallel too quick !")
	assert.Less(t, duration, int64(20000), "Parallel too slow !")
	assert.Equal(t, "foo\nbar\nbaz\n", p2.StdoutRecord())
	assert.Equal(t, "", p2.StderrRecord())

	p2.Add(sleep11ms)
	start = time.Now()
	_, err = p2.BlockRun()
	duration = time.Since(start).Microseconds()
	require.NoError(t, err)
	//assert.Len(t, 5, rc)
	assert.GreaterOrEqual(t, duration, int64(11000), "Parallel too quick !")
	assert.Less(t, duration, int64(21000), "Parallel too slow !")
	assert.Equal(t, "foo\nbar\nbaz\n", p2.StdoutRecord())
	assert.Equal(t, "", p2.StderrRecord())

	p2.Add(sleep100ms)
	start = time.Now()
	_, err = p2.BlockRun()
	duration = time.Since(start).Microseconds()
	require.NoError(t, err)
	//assert.Len(t, 5, rc)
	assert.GreaterOrEqual(t, duration, int64(100000), "Parallel too quick !")
	assert.Less(t, duration, int64(120000), "Parallel too slow !")
	assert.Equal(t, "foo\nbar\nbaz\n", p2.StdoutRecord())
	assert.Equal(t, "", p2.StderrRecord())
}
