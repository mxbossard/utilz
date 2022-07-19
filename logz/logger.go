package logz

import (
	"io"
	"os"
	"time"
	"fmt"
	"strings"
	"log"

	"mby.fr/utils/ansi"
	"mby.fr/utils/format"
)

const (
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
	namePadding int
}

func (l Basic) log(kind, color, f string, a ...interface{}) {
	f = strings.TrimSpace(f) + "\n"
	var prefix string
	if l.timed {
		time := time.Now().Format(time.RFC3339Nano)
		prefix += fmt.Sprintf("%-30s ", time)
	}

	if l.name != "" {
		paddedName := format.PadRight(l.name, l.namePadding)
		prefix += fmt.Sprintf("[%-25s] ", paddedName)
	}

	if l.colored {
		level := fmt.Sprintf("%s %s %s", color, kind, ansi.Reset)
		prefix += fmt.Sprintf("%-21s ", level)
	} else {
		prefix += fmt.Sprintf("%-6s ", kind)
	}

	_, err := fmt.Fprintf(l.out, prefix + f, a...)
	if err != nil {
		log.Fatal(err)
	}
}

func (l Basic) Trace(format string, a ...interface{}) {
	l.log("TRACE", ansi.HilightCyan, format, a...)
}

func (l Basic) Debug(format string, a ...interface{}) {
	l.log("DEBUG", ansi.HilightGreen, format, a...)
}

func (l Basic) Info(format string, a ...interface{}) {
	l.log("INFO ", ansi.HilightBlue, format, a...)
}

func (l Basic) Warn(format string, a ...interface{}) {
	l.log("WARN ", ansi.HilightYellow, format, a...)
}

func (l Basic) Error(format string, a ...interface{}) {
	l.log("ERROR", ansi.HilightRed, format, a...)
}

func (l Basic) Fatal(format string, a ...interface{}) {
	l.log("FATAL", ansi.HilightPurple, format, a...)
	os.Exit(1)
}

func New(out io.Writer, name string, namePadding int, colored, timed bool) Logger {
	return Basic{
		out: out, 
		name: name,
		namePadding: namePadding,
		colored: colored,
		timed: timed,
	}
}

func Default(name string, padding int) Logger {
	return New(os.Stdout, name, padding, true, true)
}
