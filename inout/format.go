package inout

import (
	"io"
	//"fmt"
	"strings"

	"mby.fr/utils/ansi"
)

// FIXME: For now this package only works for UTF-8

type Formatter interface {
	Format(string) string
}

// OneLineFormatter should not be called with "\n" inside input.
type OneLineFormatter func (string) string

func formatLines(in string, formatter OneLineFormatter, formatEmptyLines bool) (out string) {
        lines := strings.Split(in, "\n")
	//fmt.Println("in:", in, "lines:", lines, "len:", len(lines))

        // Write formatted lines with ending by \n
        for _, line := range lines[:len(lines) - 1] {
                if line != "" || formatEmptyLines {
                        formatted := formatter(line)
			//fmt.Println("line", line, "=>", formatted)
                        out += formatted
                }
		out += "\n"
        }

        // Write last formatted line
        lastLine := lines[len(lines) - 1]
        if lastLine != "" || formatEmptyLines && len(lines) == 1 {
                formatted := formatter(lastLine)
		//fmt.Println("lastLine", lastLine, "=>", formatted)
                out += formatted
        }

	//fmt.Println("formatLines:", in, "=>", out)
	return
}

type LineFormatter struct {
	Olf OneLineFormatter
}

func (f LineFormatter) Format(in string) string {
	return formatLines(in, f.Olf, true)
}

type AnsiFormatter struct {
        AnsiFormat string
}

func (f AnsiFormatter) Format(in string) string {
	var olf OneLineFormatter = func (line string) string {
		return f.AnsiFormat + line + ansi.Reset
	}
	return formatLines(in, olf, false)
}

type LeftPadFormatter struct {
        Pad int
}

func (f LeftPadFormatter) Format(in string) string {
        var olf OneLineFormatter = func (line string) (out string) {
		spaceCount := f.Pad - len(line)
		if spaceCount > 0 {
			out += strings.Repeat(" ", spaceCount)
		}
		out += line
		return
        }
        return formatLines(in, olf, true)
}

type PrefixFormatter struct {
	Prefix string
        LeftPad int
        RightPad int
}

func (f PrefixFormatter) Format(in string) string {
        var olf OneLineFormatter = func (line string) (out string) {
		if f.LeftPad > 0 {
			spaceCount := f.LeftPad - len(f.Prefix)
			if spaceCount > 0 {
				out += strings.Repeat(" ", spaceCount)
			}
		}

		out += f.Prefix

		if f.RightPad > 0 {
			spaceCount := f.RightPad - len(out)
			if spaceCount > 0 {
				out += strings.Repeat(" ", spaceCount)
			}
		}

		out += line

		return
        }
        return formatLines(in, olf, true)
}

type FormattingWriter struct {
	out io.Writer
	formatter Formatter
}

func (w FormattingWriter) Write(in []byte) (n int, err error) {
	s := string(in)
	formatted := w.formatter.Format(s)
	return w.out.Write([]byte(formatted))
}

func NewFormattingWriter(out io.Writer, f Formatter) FormattingWriter {
	return FormattingWriter{out, f}
}
