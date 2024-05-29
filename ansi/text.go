package ansi

import (
	"fmt"
	"strings"
)

type FormattedText struct {
	F          Color
	Nested     []*Formatted
	Formatting bool
}

func (f FormattedText) format() string {
	b := strings.Builder{}
	for _, s := range f.Nested {
		b.WriteString(s.String())
	}
	return b.String()
}

func (f FormattedText) Raw() string {
	b := strings.Builder{}
	for _, s := range f.Nested {
		b.WriteString(s.Raw())
	}
	return b.String()
}

func (f FormattedText) String() string {
	if f.Formatting {
		return f.format()
	}
	return f.Raw()
}

func (f *FormattedText) Disable() *FormattedText {
	f.Formatting = false
	return f
}

func (f *FormattedText) Enable() *FormattedText {
	f.Formatting = true
	return f
}

func (f *FormattedText) Cat(o any) *FormattedText {
	switch n := o.(type) {
	case Formatted:
		f.Nested = append(f.Nested, &n)
	case *Formatted:
		f.Nested = append(f.Nested, n)
	case Stringer:
		s := n.String()
		f.Nested = append(f.Nested, Format(f.F, s))
	case string:
		f.Nested = append(f.Nested, Format(f.F, n))
	default:
		msg := fmt.Sprintf("canot Cat() type: %T", n)
		panic(msg)
	}
	return f
}

func (f *FormattedText) Catln(o any) *FormattedText {
	return f.Cat(o).Fcat(None, "\n")
}

func (f *FormattedText) Join(o any, sep string) *FormattedText {
	return f.Cat(o).Fcat(None, sep)
}

func (f *FormattedText) Fcat(format Color, o any) *FormattedText {
	var formatted *Formatted
	switch n := o.(type) {
	case Formatted:
		formatted = Format(format, n.Squash(true).Raw())
	case *Formatted:
		formatted = Format(format, n.Squash(true).Raw())
	case Stringer, string:
		formatted = Format(format, n)
	default:
		msg := fmt.Sprintf("canot Cat() type: %T", n)
		panic(msg)
	}
	f.Cat(formatted)
	return f
}

func (f *FormattedText) Fcatln(format Color, o any) *FormattedText {
	return f.Fcat(format, o).Fcat(None, "\n")
}

func (f *FormattedText) Fjoin(format Color, o any, sep string) *FormattedText {
	return f.Fcat(format, o).Fcat(None, sep)
}

func Text(color Color, inputs ...any) *FormattedText {
	f := FormattedText{
		F:          color,
		Formatting: true,
	}

	for _, item := range inputs {
		f.Cat(item)
	}

	return &f
}
