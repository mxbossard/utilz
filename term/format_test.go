package term

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const format = "\033[1;30m"
const clear = "\033[0m"

func expectedAnsiFormat(in string) string {
	return format + in + clear
}

func TestAnsiFormatter(t *testing.T) {
	formatter := AnsiFormatter{format}
	var buffer strings.Builder
	writer := NewFormattingWriter(&buffer, formatter)

	cases := []struct {
		in, want string
		err error
	}{
		{"", "", nil},
		{"\n", "\n", nil},
		{"foo", expectedAnsiFormat("foo"), nil},
		{"foo\n", expectedAnsiFormat("foo") + "\n", nil},
		{"foo\nbar", expectedAnsiFormat("foo") + "\n" + expectedAnsiFormat("bar"), nil},
		{"foo\nbar\n", expectedAnsiFormat("foo") + "\n" + expectedAnsiFormat("bar") + "\n", nil},
	}
	for i, c := range cases {
		got := formatter.Format(c.in)
		assert.Equal(t, c.want, got, "Bad Formatter value for case #%d", i)

		buffer.Reset()
		_, err := writer.Write([]byte(c.in))
		assert.ErrorIs(t, err, c.err, "Bad error returned for case #%d", i)
		assert.Equal(t, c.want, buffer.String(), "Bad data written for  case #%d", i)
	}
}

func TestLeftPadFormatter(t *testing.T) {
	formatter := LeftPadFormatter{10}
	var buffer strings.Builder
	writer := NewFormattingWriter(&buffer, formatter)

	cases := []struct {
		in, want string
		err error
	}{
		{"", "          ", nil},
		{"\n", "          \n", nil},
		{"foo", "       foo", nil},
		{"foo\n", "       foo\n", nil},
		{"foo\nbar", "       foo\n       bar", nil},
		{"foo\nbar\n", "       foo\n       bar\n", nil},
	}
	for i, c := range cases {
		got := formatter.Format(c.in)
		assert.Equal(t, c.want, got, "Bad Formatter value for case #%d", i)

		buffer.Reset()
		_, err := writer.Write([]byte(c.in))
		assert.ErrorIs(t, err, c.err, "Bad error returned for case #%d", i)
		assert.Equal(t, c.want, buffer.String(), "Bad data written for  case #%d", i)
	}
}

func TestLineFormatter(t *testing.T) {
	prefix := "prefix"
	suffix := "suffix"
	formatter := LineFormatter{func(line string) string {

	}}
	var buffer strings.Builder
	writer := NewFormattingWriter(&buffer, formatter)

	cases := []struct {
		in, want string
		err error
	}{
		{"", "          ", nil},
		{"\n", "          \n", nil},
		{"foo", "       foo", nil},
		{"foo\n", "       foo\n", nil},
		{"foo\nbar", "       foo\n       bar", nil},
		{"foo\nbar\n", "       foo\n       bar\n", nil},
	}
	for i, c := range cases {
		got := formatter.Format(c.in)
		assert.Equal(t, c.want, got, "Bad Formatter value for case #%d", i)

		buffer.Reset()
		_, err := writer.Write([]byte(c.in))
		assert.ErrorIs(t, err, c.err, "Bad error returned for case #%d", i)
		assert.Equal(t, c.want, buffer.String(), "Bad data written for  case #%d", i)
	}
}

