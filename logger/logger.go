package logger

import (
	"io"
	"os"
	"time"
	"fmt"
	"log"
)

const (
	ansiClear = "\033[0m"
	traceAnsiColor = "\033[0;97;46m"
	debugAnsiColor = "\033[0;97;44m"
	infoAnsiColor = "\033[0;97;42m"
	warnAnsiColor = "\033[0;31;43m"
	errorAnsiColor = "\033[0;97;41m"
	fatalAnsiColor = "\033[0;97;45m"

)

type Logger interface {
	Trace(string, ...interface{})
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
	Fatal(string, ...interface{})
}

type BasicLogger struct {
	out io.Writer
	name string
	colored, timed bool
}

func (l BasicLogger) log(kind, color, format string, a ...interface{}) {
	var prefix string
	if l.timed {
		prefix += time.Now().Format(time.RFC3339Nano)
	}
	if l.colored {
		prefix += fmt.Sprintf("%s%6s%s ", color, kind, ansiClear)
	} else {
		prefix += fmt.Sprintf("%6s ", kind)
	}
	if l.name != "" {
		prefix += fmt.Sprintf("[%s] ", l.name)
	}
	_, err := fmt.Fprintf(l.out, prefix + format, a...)
	if err != nil {
		log.Fatal(err)
	}
}

func (l BasicLogger) Trace(format string, a ...interface{}) {
	l.log("TRACE", traceAnsiColor, format, a...)
}

func (l BasicLogger) Debug(format string, a ...interface{}) {
	l.log("DEBUG", debugAnsiColor, format, a...)
}

func (l BasicLogger) Info(format string, a ...interface{}) {
	l.log("INFO", infoAnsiColor, format, a...)
}

func (l BasicLogger) Warn(format string, a ...interface{}) {
	l.log("WARN", warnAnsiColor, format, a...)
}

func (l BasicLogger) Error(format string, a ...interface{}) {
	l.log("ERROR", errorAnsiColor, format, a...)
}

func (l BasicLogger) Fatal(format string, a ...interface{}) {
	l.log("Fatal", fatalAnsiColor, format, a...)
	os.Exit(1)
}

func New(out io.Writer, name string, colored, timed bool) Logger {
	return BasicLogger{
		out: out, 
		name: name,
		colored: colored,
		timed: timed,
	}
}

func Default(name string) Logger {
	return New(os.Stdout, name, true, true)
}
