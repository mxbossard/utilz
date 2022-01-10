package logger

import (
	"io"
	"os"
	"time"
	"fmt"
	"strings"
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

type Basic struct {
	out io.Writer
	name string
	colored, timed bool
}

func (l Basic) log(kind, color, format string, a ...interface{}) {
	format = strings.TrimSpace(format) + "\n"
	var prefix string
	if l.timed {
		time := time.Now().Format(time.RFC3339Nano)
		prefix += fmt.Sprintf("%-30s ", time)
	}

	if l.name != "" {
		prefix += fmt.Sprintf("[ %-25s ] ", l.name)
	}

	if l.colored {
		level := fmt.Sprintf("%s %s %s", color, kind, ansiClear)
		prefix += fmt.Sprintf("%-21s ", level)
	} else {
		prefix += fmt.Sprintf("%-6s ", kind)
	}

	_, err := fmt.Fprintf(l.out, prefix + format, a...)
	if err != nil {
		log.Fatal(err)
	}
}

func (l Basic) Trace(format string, a ...interface{}) {
	l.log("TRACE", traceAnsiColor, format, a...)
}

func (l Basic) Debug(format string, a ...interface{}) {
	l.log("DEBUG", debugAnsiColor, format, a...)
}

func (l Basic) Info(format string, a ...interface{}) {
	l.log("INFO", infoAnsiColor, format, a...)
}

func (l Basic) Warn(format string, a ...interface{}) {
	l.log("WARN", warnAnsiColor, format, a...)
}

func (l Basic) Error(format string, a ...interface{}) {
	l.log("ERROR", errorAnsiColor, format, a...)
}

func (l Basic) Fatal(format string, a ...interface{}) {
	l.log("Fatal", fatalAnsiColor, format, a...)
	os.Exit(1)
}

func New(out io.Writer, name string, colored, timed bool) Logger {
	return Basic{
		out: out, 
		name: name,
		colored: colored,
		timed: timed,
	}
}

func Default(name string) Logger {
	return New(os.Stdout, name, true, true)
}
