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
	e1 = Execution("echo", "foo")
	e2 = Execution("echo", "bar")
	e3 = Execution("echo", "baz")

	sleep10ms  = Execution("sleep", "0.01")
	sleep11ms  = Execution("sleep", "0.011")
	sleep100ms = Execution("sleep", "0.1")
	sleep200ms = Execution("sleep", "0.2")
)

func TestSequence_Serial(t *testing.T) {
	e1.reset()
	e2.reset()
	e3.reset()

	s := Sequence().Serial(e1)
	assert.Equal(t, "echo foo", s.String())

	rc1, err1 := s.BlockRun()
	require.NoError(t, err1)
	assert.Equal(t, 0, rc1)
	assert.Equal(t, "foo\n", e1.StdoutRecord.String())
	assert.Equal(t, "", e2.StdoutRecord.String())
	assert.Equal(t, "", e3.StdoutRecord.String())

	s2 := Sequence().Serial(e1, e2)
	assert.Equal(t, "echo foo\necho bar", s2.String())

	rc2, err2 := s2.BlockRun()
	require.NoError(t, err2)
	assert.Equal(t, 0, rc2)
	assert.Equal(t, "foo\n", e1.StdoutRecord.String())
	assert.Equal(t, "bar\n", e2.StdoutRecord.String())
	assert.Equal(t, "", e3.StdoutRecord.String())

	s2.Serial(e3)
	assert.Equal(t, "echo foo\necho bar\necho baz", s2.String())

	rc3, err3 := s2.BlockRun()
	require.NoError(t, err3)
	assert.Equal(t, 0, rc3)
	assert.Equal(t, "foo\n", e1.StdoutRecord.String())
	assert.Equal(t, "bar\n", e2.StdoutRecord.String())
	assert.Equal(t, "baz\n", e3.StdoutRecord.String())

	// Test serial timings
	s2.Serial(sleep10ms)
	start := time.Now()
	_, err := s2.BlockRun()
	duration := time.Since(start).Microseconds()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, duration, int64(10000), "Serial too quick !")
	assert.Less(t, duration, int64(20000), "Serial too slow !")

	s2.Serial(sleep10ms)
	start = time.Now()
	_, err = s2.BlockRun()
	duration = time.Since(start).Microseconds()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, duration, int64(20000), "Serial too quick !")
	assert.Less(t, duration, int64(30000), "Serial too slow !")

	s2.Serial(sleep100ms)
	start = time.Now()
	_, err = s2.BlockRun()
	duration = time.Since(start).Microseconds()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, duration, int64(120000), "Serial too quick !")
	assert.Less(t, duration, int64(150000), "Serial too slow !")
}

func TestSequence_Parallel(t *testing.T) {
	e1.reset()
	e2.reset()
	e3.reset()

	p := Sequence().Parallel(e1)
	assert.Equal(t, "echo foo", p.String())

	rc, err := p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo\n", e1.StdoutRecord.String())
	assert.Equal(t, "", e2.StdoutRecord.String())
	assert.Equal(t, "", e3.StdoutRecord.String())

	p2 := Sequence().Parallel(e1, e2)
	assert.Equal(t, "echo foo\necho bar", p2.String())

	rc2, err2 := p2.BlockRun()
	require.NoError(t, err2)
	assert.Equal(t, 0, rc2)
	assert.Equal(t, "foo\n", e1.StdoutRecord.String())
	assert.Equal(t, "bar\n", e2.StdoutRecord.String())
	assert.Equal(t, "", e3.StdoutRecord.String())

	p2.Parallel(e3)
	assert.Equal(t, "echo foo\necho bar\necho baz", p2.String())

	rc3, err3 := p2.BlockRun()
	require.NoError(t, err3)
	assert.Equal(t, 0, rc3)
	assert.Equal(t, "foo\n", e1.StdoutRecord.String())
	assert.Equal(t, "bar\n", e2.StdoutRecord.String())
	assert.Equal(t, "baz\n", e3.StdoutRecord.String())

	// Test serial timings
	p2.Parallel(sleep10ms)
	start := time.Now()
	_, err = p2.BlockRun()
	duration := time.Since(start).Microseconds()
	require.NoError(t, err)
	//assert.Len(t, 4, rc)
	assert.GreaterOrEqual(t, duration, int64(10000), "Parallel too quick !")
	assert.Less(t, duration, int64(20000), "Parallel too slow !")

	p2.Parallel(sleep11ms)
	start = time.Now()
	_, err = p2.BlockRun()
	duration = time.Since(start).Microseconds()
	require.NoError(t, err)
	//assert.Len(t, 5, rc)
	assert.GreaterOrEqual(t, duration, int64(11000), "Parallel too quick !")
	assert.Less(t, duration, int64(21000), "Parallel too slow !")

	p2.Parallel(sleep100ms)
	start = time.Now()
	_, err = p2.BlockRun()
	duration = time.Since(start).Microseconds()
	require.NoError(t, err)
	//assert.Len(t, 5, rc)
	assert.GreaterOrEqual(t, duration, int64(100000), "Parallel too quick !")
	assert.Less(t, duration, int64(120000), "Parallel too slow !")
}
