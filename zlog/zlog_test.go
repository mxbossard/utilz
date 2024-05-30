package zlog

import (
	"log"
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

func TestDefault_Before(t *testing.T) {
	b := open()
	SetLogLevelThreshold(LevelError)

	log.Printf("foo")

	logged := b.String()
	assert.Contains(t, logged, "foo")

	b.Reset()

	SetLogLevelThreshold(LevelDebug)
	log.Printf("bar")

	logged = b.String()
	assert.Contains(t, logged, "bar")
}

func TestDefault_After(t *testing.T) {
	b := open()

	SetDefault()
	SetLogLevelThreshold(LevelError)

	log.Printf("foo")

	logged := b.String()
	assert.Empty(t, logged)

	b.Reset()

	SetLogLevelThreshold(LevelDebug)
	log.Printf("bar")

	logged = b.String()
	assert.Contains(t, logged, "DEBUG [default] bar")
	//assert.Contains(t, logged, "source=zlog/zlog_test.go:")
	//assert.NotContains(t, logged, "source=zlog/zlog.go:")
}

func TestNew_WithQualifier(t *testing.T) {
	b := open()
	SetLogLevelThreshold(LevelError)

	logger := New("foo")
	logger.Error("bar")

	logged := b.String()
	assert.Contains(t, logged, "ERROR [foo] bar")
	assert.Contains(t, logged, "source=zlog/zlog_test.go:")
	assert.NotContains(t, logged, "source=zlog/zlog.go:")
}

func TestNew_WithHandler(t *testing.T) {
	b := open()
	SetLogLevelThreshold(LevelError)

	handler := slog.NewTextHandler(b, defaultHandlerOptions)

	logger := New(handler)
	logger.Error("bar")

	logged := b.String()
	assert.Contains(t, logged, "level=ERROR")
	assert.Contains(t, logged, "msg=bar")
	assert.Contains(t, logged, "qualifier=mby.fr/utils/zlog")
	assert.Contains(t, logged, "/mby.fr/utils/zlog/zlog_test.go:")
	assert.NotContains(t, logged, "/mby.fr/utils/zlog/zlog.go:")
}

func TestTrace(t *testing.T) {
	b := open()
	SetLogLevelThreshold(LevelError)

	logger := New("foo")
	logger.Trace("baz trace")

	assert.Empty(t, b.String())

	b.Reset()
	SetLogLevelThreshold(LevelTrace)

	logger.Trace("baz trace2")

	logged := b.String()
	assert.Contains(t, logged, "TRACE [foo] baz trace")
	assert.Contains(t, logged, "source=zlog/zlog_test.go:")
	assert.NotContains(t, logged, "source=zlog/zlog.go:")
}

func TestPerf(t *testing.T) {
	b := open()
	SetLogLevelThreshold(LevelError)

	logger := New("foo")
	logger.Perf("baz perf")

	assert.Empty(t, b.String())

	b.Reset()
	SetLogLevelThreshold(LevelPerf)

	logger.Perf("baz perf2")

	logged := b.String()
	assert.Contains(t, logged, "PERF [foo] baz perf")
	assert.Contains(t, logged, "source=zlog/zlog_test.go:")
	assert.NotContains(t, logged, "source=zlog/zlog.go:")
}

func TestDebug(t *testing.T) {
	b := open()
	SetLogLevelThreshold(LevelError)

	logger := New("foo")
	logger.Debug("baz debug")

	assert.Empty(t, b.String())

	b.Reset()
	SetLogLevelThreshold(LevelDebug)

	logger.Debug("baz debug2")

	logged := b.String()
	assert.Contains(t, logged, "DEBUG [foo] baz debug2")
	assert.Contains(t, logged, "source=zlog/zlog_test.go:")
	assert.NotContains(t, logged, "source=zlog/zlog.go:")
}

func TestInfo(t *testing.T) {
	b := open()
	SetLogLevelThreshold(LevelError)

	logger := New("foo")
	logger.Info("baz info")

	assert.Empty(t, b.String())

	b.Reset()
	SetLogLevelThreshold(LevelInfo)

	logger.Info("baz info2")

	logged := b.String()
	assert.Contains(t, logged, "INFO [foo] baz info2")
	assert.Contains(t, logged, "source=zlog/zlog_test.go:")
	assert.NotContains(t, logged, "source=zlog/zlog.go:")
}

func TestWarn(t *testing.T) {
	b := open()
	SetLogLevelThreshold(LevelError)

	logger := New("foo")
	logger.Warn("baz warn")

	assert.Empty(t, b.String())

	b.Reset()
	SetLogLevelThreshold(LevelWarn)

	logger.Warn("baz warn2")

	logged := b.String()
	assert.Contains(t, logged, "WARN [foo] baz warn2")
	assert.Contains(t, logged, "source=zlog/zlog_test.go:")
	assert.NotContains(t, logged, "source=zlog/zlog.go:")
}

func TestError(t *testing.T) {
	b := open()
	SetLogLevelThreshold(LevelFatal)

	logger := New("foo")
	logger.Error("baz error")

	assert.Empty(t, b.String())

	b.Reset()
	SetLogLevelThreshold(LevelDebug)

	logger.Error("baz error2")

	logged := b.String()
	assert.Contains(t, logged, "ERROR [foo] baz error2")
	assert.Contains(t, logged, "source=zlog/zlog_test.go:")
	assert.NotContains(t, logged, "source=zlog/zlog.go:")
}

func TestFatal(t *testing.T) {
	t.Skip("cannot test fatal which exit")
	b := open()
	SetLogLevelThreshold(LevelError)

	logger := New("foo")
	logger.Fatal("baz fatal")

	logged := b.String()
	assert.Contains(t, logged, "FATAL [foo] baz fatal")
	assert.Contains(t, logged, "source=zlog/zlog_test.go:")
	assert.NotContains(t, logged, "source=zlog/zlog.go:")

	b.Reset()
	SetLogLevelThreshold(LevelFatal)

	logger.Fatal("baz fatal2")

	logged = b.String()
	assert.Contains(t, logged, "FATAL [foo] baz fatal2")
	assert.Contains(t, logged, "source=zlog/zlog_test.go:")
	assert.NotContains(t, logged, "source=zlog/zlog.go:")
}

func TestStartPerf(t *testing.T) {
	b := open()
	SetLogLevelThreshold(LevelPerf)

	logger := New("foo")
	p := logger.StartPerf()
	p.End()

	logged := b.String()
	assert.Contains(t, logged, "PERF [foo] TestStartPerf() started ...")
	assert.Contains(t, logged, "PERF [foo] TestStartPerf() ended in")
	assert.Contains(t, logged, "source=zlog/zlog_test.go:")
	assert.NotContains(t, logged, "source=zlog/zlog.go:")
}

func TestColors(t *testing.T) {
	//t.Skip()
	b := open()
	UseColoredDefaultHandler()
	SetLogLevelThreshold(LevelTrace)

	logger := New()
	logger.Trace("trace message", "key", "value")
	logger.Perf("perf message", "key", "value")
	logger.Debug("debug message", "key", "value")
	logger.Info("info message", "key", "value")
	logger.Warn("warn message", "key", "value")
	logger.Error("error message", "key", "value")
	//logger.Fatal("trace message")

	logged := b.String()
	assert.Contains(t, logged, "source=zlog/zlog_test.go:")
	assert.NotContains(t, logged, "source=zlog/zlog.go:")
	//assert.Equal(t, "", logged)
}
