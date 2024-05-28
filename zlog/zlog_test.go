package zlog

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func open() *strings.Builder {
	b := &strings.Builder{}
	SetDefaultOutput(b)
	return b
}

func TestZlog(t *testing.T) {
	logger := New("foo")
	assert.NotNil(t, logger)
}

func TestNew_WithQualifier(t *testing.T) {
	b := open()
	SetDefaultLogLevel(LevelError)

	logger := New("foo")
	logger.Error("bar")

	logged := b.String()
	assert.Contains(t, logged, "ERROR [foo] bar")
}

func TestNew_WithHandler(t *testing.T) {
	b := open()
	SetDefaultLogLevel(LevelError)

	handler := slog.NewTextHandler(b, nil)

	logger := New(handler)
	logger.Error("bar")

	logged := b.String()
	assert.Contains(t, logged, "level=ERROR")
	assert.Contains(t, logged, "msg=bar")
	assert.Contains(t, logged, "qualifier=mby.fr/utils/zlog")
}

func TestTrace(t *testing.T) {
	b := open()
	SetDefaultLogLevel(LevelError)

	logger := New("foo")
	logger.Trace("bar")

	assert.Empty(t, b.String())

	b.Reset()
	SetDefaultLogLevel(LevelTrace)

	logger.Trace("bar")

	assert.Contains(t, b.String(), "TRACE [foo] bar")
}

func TestPerf(t *testing.T) {
	b := open()
	SetDefaultLogLevel(LevelError)

	logger := New("foo")
	logger.Perf("bar")

	assert.Empty(t, b.String())

	b.Reset()
	SetDefaultLogLevel(LevelPerf)

	logger.Perf("bar")

	assert.Contains(t, b.String(), "PERF [foo] bar")
}

func TestFatal(t *testing.T) {
	t.Skip("cannot test fatal which exit")
	b := open()
	SetDefaultLogLevel(LevelError)

	logger := New("foo")
	logger.Fatal("bar")

	assert.Contains(t, b.String(), "FATAL [foo] bar")

	b.Reset()
	SetDefaultLogLevel(LevelFatal)

	logger.Fatal("bar")

	assert.Contains(t, b.String(), "FATAL [foo] bar")
}

func TestStartPerf(t *testing.T) {
	b := open()
	SetDefaultLogLevel(LevelPerf)

	logger := New("foo")
	p := logger.StartPerf()
	p.End()

	assert.Contains(t, b.String(), "PERF [foo] TestStartPerf() started ...")
	assert.Contains(t, b.String(), "PERF [foo] TestStartPerf() ended in")
}
