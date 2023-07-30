package cmdz

import (
	//"fmt"
	//"io"
	//"context"
	//"log"
	//"os/exec"
	"strings"
	"testing"
	"time"

	//"mby.fr/utils/promise"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerial(t *testing.T) {
	e1 := Cmd("echo", "foo")
	e2 := Cmd("echo", "bar")
	e3 := Cmd("echo", "baz")
	sleep10ms := Cmd("sleep", "0.01")
	sleep100ms := Cmd("sleep", "0.1")

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
	assert.Less(t, duration, int64(99000), "Serial too slow !")
	assert.Equal(t, "foo\nbar\nbaz\n", s2.StdoutRecord())
	assert.Equal(t, "", s2.StderrRecord())

	s2.Add(sleep100ms)
	start = time.Now()
	_, err = s2.BlockRun()
	duration = time.Since(start).Microseconds()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, duration, int64(120000), "Serial too quick !")
	assert.Less(t, duration, int64(199000), "Serial too slow !")
	assert.Equal(t, "foo\nbar\nbaz\n", s2.StdoutRecord())
	assert.Equal(t, "", s2.StderrRecord())
}

func TestSerial_Retries(t *testing.T) {
	e1 := Cmd("echo", "foo")
	e2 := Cmd("echo", "bar")
	f1 := Cmd("false")

	s := Serial(e1, f1, e2)
	rc, err := s.BlockRun()

	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e1.ResultCodes())
	assert.Equal(t, []int{1}, f1.ResultCodes())
	assert.Equal(t, []int{0}, e2.ResultCodes())

	s.Retries(2, 10)
	rc, err = s.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e1.ResultCodes())
	assert.Equal(t, []int{1, 1, 1}, f1.ResultCodes())
	assert.Equal(t, []int{0}, e2.ResultCodes())

	s.Add(f1)
	rc, err = s.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 1, rc)
}

func TestSerial_Outputs(t *testing.T) {
	e1 := Cmd("echo", "foo")
	e2 := Cmd("echo", "bar")
	f1 := Cmd("false")

	sb := strings.Builder{}
	s := Serial(e1, f1, e2).SetOutputs(&sb, nil)
	rc, err := s.BlockRun()

	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo\nbar\n", sb.String())
}

func TestSerial_FailFast(t *testing.T) {
	e1 := Cmd("echo", "foo")
	e2 := Cmd("echo", "bar")
	f1 := Cmd("false")

	s := Serial(e1, f1, e2).FailFast(true).ErrorOnFailure(false)
	rc, err := s.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 1, rc)
	assert.Equal(t, []int{0}, e1.ResultCodes())
	assert.Equal(t, []int{1}, f1.ResultCodes())
	assert.Equal(t, []int(nil), e2.ResultCodes())
	assert.Equal(t, "foo\n", s.StdoutRecord())

	s.Retries(2, 10)
	rc, err = s.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 1, rc)
	assert.Equal(t, []int{0}, e1.ResultCodes())
	assert.Equal(t, []int{1, 1, 1}, f1.ResultCodes())
	assert.Equal(t, []int(nil), e2.ResultCodes())
	assert.Equal(t, "foo\n", s.StdoutRecord())
}

func TestSerial_ErrorOnFailure(t *testing.T) {
	e1 := Cmd("echo", "foo")
	f := Cmd("/bin/false").ErrorOnFailure(true)
	s := Serial(e1, f)
	rc, err := s.BlockRun()
	require.Error(t, err, "should error")
	assert.Equal(t, -1, rc)
}

func TestOr(t *testing.T) {
	e1 := Cmd("echo", "foo")
	e2 := Cmd("echo", "bar")
	e3 := Cmd("echo", "baz")

	s := Or(e1)
	assert.Equal(t, "echo foo", s.String())

	rc1, err1 := s.BlockRun()
	require.NoError(t, err1)
	assert.Equal(t, 0, rc1)
	assert.Equal(t, "foo\n", e1.StdoutRecord())
	assert.Equal(t, "", e2.StdoutRecord())
	assert.Equal(t, "", e3.StdoutRecord())
	assert.Equal(t, "foo\n", s.StdoutRecord())
	assert.Equal(t, "", s.StderrRecord())

	s2 := Or(e1, e2)
	assert.Equal(t, "echo foo || echo bar", s2.String())

	rc2, err2 := s2.BlockRun()
	require.NoError(t, err2)
	assert.Equal(t, 0, rc2)
	assert.Equal(t, "foo\n", e1.StdoutRecord())
	assert.Equal(t, "", e2.StdoutRecord())
	assert.Equal(t, "", e3.StdoutRecord())
	assert.Equal(t, "foo\n", s2.StdoutRecord())
	assert.Equal(t, "", s2.StderrRecord())

	s2.Add(e3)
	assert.Equal(t, "echo foo || echo bar || echo baz", s2.String())

	rc3, err3 := s2.BlockRun()
	require.NoError(t, err3)
	assert.Equal(t, 0, rc3)
	assert.Equal(t, "foo\n", e1.StdoutRecord())
	assert.Equal(t, "", e2.StdoutRecord())
	assert.Equal(t, "", e3.StdoutRecord())
	assert.Equal(t, "foo\n", s2.StdoutRecord())
	assert.Equal(t, "", s2.StderrRecord())
}

func TestOr_Retries(t *testing.T) {
	e1 := Cmd("echo", "foo")
	e2 := Cmd("echo", "bar")
	f1 := Cmd("false")

	s := Or(e1, f1, e2)
	rc, err := s.BlockRun()

	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e1.ResultCodes())
	assert.Equal(t, []int(nil), f1.ResultCodes())
	assert.Equal(t, []int(nil), e2.ResultCodes())

	s = Or(f1, e1, e2)
	rc, err = s.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{1}, f1.ResultCodes())
	assert.Equal(t, []int{0}, e1.ResultCodes())
	assert.Equal(t, []int(nil), e2.ResultCodes())

	s.Retries(2, 10)
	rc, err = s.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{1, 1, 1}, f1.ResultCodes())
	assert.Equal(t, []int{0}, e1.ResultCodes())
	assert.Equal(t, []int(nil), e2.ResultCodes())
}

func TestOr_Outputs(t *testing.T) {
	e1 := Cmd("echo", "foo")
	e2 := Cmd("echo", "bar")
	f1 := Cmd("false")

	sb := strings.Builder{}
	s := Or(e1, f1, e2).SetOutputs(&sb, nil)
	rc, err := s.BlockRun()

	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo\n", sb.String())
}

func TestOr_ErrorOnFailure(t *testing.T) {
	e1 := Cmd("echo", "foo")
	f := Cmd("/bin/false").ErrorOnFailure(true)
	s := Or(f, e1)
	rc, err := s.BlockRun()
	require.Error(t, err, "should error")
	assert.Equal(t, -1, rc)
}

func TestParallel(t *testing.T) {
	e1 := Cmd("echo", "foo")
	e2 := Cmd("echo", "bar")
	e3 := Cmd("echo", "baz")
	sleep10ms := Cmd("sleep", "0.01")
	sleep11ms := Cmd("sleep", "0.011")
	sleep100ms := Cmd("sleep", "0.1")

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

func TestParallel_Retries(t *testing.T) {
	e1 := Cmd("echo", "foo")
	e2 := Cmd("echo", "bar")
	f1 := Cmd("false")
	f2 := Cmd("false")

	p := Parallel(e1, f1, e2)
	rc, err := p.BlockRun()

	require.NoError(t, err)
	assert.Equal(t, 1, rc)
	assert.Equal(t, []int{0}, e1.ResultCodes())
	assert.Equal(t, []int{1}, f1.ResultCodes())
	assert.Equal(t, []int{0}, e2.ResultCodes())

	p.Retries(2, 10)
	rc, err = p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 1, rc)
	assert.Equal(t, []int{0}, e1.ResultCodes())
	assert.Equal(t, []int{1, 1, 1}, f1.ResultCodes())
	assert.Equal(t, []int{0}, e2.ResultCodes())

	p.Add(f2)
	rc, err = p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 1, rc)
}

func TestParallel_Outputs(t *testing.T) {
	c1 := Cmd("/bin/sh", "-c", "sleep 0.1 ; echo foo")
	c2 := Cmd("/bin/sh", "-c", "sleep 0.2 ; echo bar")
	f1 := Cmd("false")

	sb := strings.Builder{}
	p := Parallel(c1, f1, c2).SetOutputs(&sb, nil)
	rc, err := p.BlockRun()

	require.NoError(t, err)
	assert.Equal(t, 1, rc)
	assert.Equal(t, "foo\nbar\n", sb.String())
}

func TestParallel_FailFast(t *testing.T) {
	c1 := Cmd("/bin/sh", "-c", "sleep 0.1 && echo foo")
	c2 := Cmd("/bin/sh", "-c", "sleep 0.2 && echo bar")
	f1 := Cmd("false")

	p := Parallel(c1, f1, c2).FailFast(true)
	rc, err := p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 1, rc)
	assert.Equal(t, "", p.StdoutRecord())
	assert.Equal(t, []int(nil), c1.ResultCodes())
	assert.Equal(t, []int{1}, f1.ResultCodes())
	assert.Equal(t, []int(nil), c2.ResultCodes())

	c1.reset()
	c2.reset()
	f1.reset()
	p.Retries(2, 10)
	rc, err = p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 1, rc)
	assert.Equal(t, []int(nil), c1.ResultCodes())
	assert.Equal(t, []int{1, 1, 1}, f1.ResultCodes())
	assert.Equal(t, []int(nil), c2.ResultCodes())
	assert.Equal(t, "", p.StdoutRecord())
}

func TestParallel_ErrorOnFailure(t *testing.T) {
	e1 := Cmd("echo", "foo")
	f := Cmd("/bin/false").ErrorOnFailure(true)
	p := Parallel(e1, f)
	rc, err := p.BlockRun()
	require.Error(t, err, "should error")
	assert.Equal(t, -1, rc)
}

func Test_Chaining(t *testing.T) {
	e1 := Cmd("echo", "foo")
	es2 := Cmd("/bin/sh", "-c", "sleep 0.2 && echo bar")
	e3 := Cmd("echo", "baz")

	p := Parallel()
	p.Add(e1, Serial(es2, e3))

	s := Serial()
	s.Add(p)
	s.Add(e1)

	rc, err := s.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo\nbar\nbaz\nfoo\n", s.StdoutRecord())
}

func Test_FallbackPersitence(t *testing.T) {
	// FIXME
	t.Skip("Skipped because not implemented.")

	f := Cmd("/bin/false")

	rc, err := f.BlockRun()
	assert.NoError(t, err)
	assert.Equal(t, 1, rc)
	// Retries should be 0 by default
	assert.Equal(t, []int{1}, f.ResultCodes(), "Bad retries defaulr")

	s := Serial(f).Retries(2, 0)
	rc, err = s.BlockRun()
	assert.NoError(t, err)
	assert.Equal(t, 1, rc)
	// Retries should be 2 fallback by Serial config
	assert.Equal(t, []int{1, 1, 1}, f.ResultCodes(), "Bad retries fallback")

	rc, err = f.BlockRun()
	assert.NoError(t, err)
	assert.Equal(t, 1, rc)
	// Retries should be 0 by default fallback should not persist
	assert.Equal(t, []int{1}, f.ResultCodes(), "Undesired fallback persistance")
}
