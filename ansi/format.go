package ansi

import (
	"fmt"
	"regexp"
)

var formatPattern = regexp.MustCompile(`\033\[0(;\d{2,3}){0,2}m`)

type Stringer interface {
	String() string
}

type Formatter interface {
	Stringer

	Format() string
	Raw() string
	Disable() Formatter
	Enable() Formatter
	Squash(bool) Formatter
}

type Formatted struct {
	F          Color
	Nested     any
	Formatting bool
	Squashing  bool
}

func (f Formatted) Format() string {
	var nested string
	switch n := f.Nested.(type) {
	case Formatted:
		if f.Squashing {
			nested = n.Raw()
		} else {
			nested = n.Format()
		}
	case Stringer:
		nested = n.String()
	case string:
		nested = n
	default:
		msg := fmt.Sprintf("canot Format() nested type: %T", n)
		panic(msg)
	}
	if f.F == None {
		return nested
	}
	if f.Squashing {
		nested = Unformat(nested)
	}
	return fmt.Sprintf("%v%s%v", f.F, nested, Reset)
}

func (f Formatted) Raw() string {
	var nested string
	switch n := f.Nested.(type) {
	case Formatted:
		nested = n.Raw()
	case Stringer:
		nested = n.String()
	case string:
		nested = n
	default:
		msg := fmt.Sprintf("canot Raw() nested type: %T", n)
		panic(msg)
	}
	if f.Squashing {
		nested = Unformat(nested)
	}
	return nested
}

func (f Formatted) String() string {
	if f.Formatting {
		return f.Format()
	}
	return f.Raw()
}

func (f *Formatted) Disable() *Formatted {
	f.Formatting = false
	return f
}

func (f *Formatted) Enable() *Formatted {
	f.Formatting = true
	return f
}

func (f *Formatted) Squash(ok bool) *Formatted {
	f.Squashing = ok
	return f
}

func Format(color Color, s any) *Formatted {
	f := Formatted{
		F:          color,
		Nested:     s,
		Formatting: true,
		Squashing:  true,
	}
	return &f
}

func Unformat(o any) string {
	var s string
	switch n := o.(type) {
	case Formatted:
		s = n.Raw()
	case *Formatted:
		s = n.Raw()
	case Stringer:
		s = n.String()
	case string:
		s = n
	default:
		msg := fmt.Sprintf("canot Unformat() type: %T", n)
		panic(msg)
	}
	return formatPattern.ReplaceAllString(s, "")
}

func Sprintf(color Color, s string, objects ...any) string {
	return Format(color, fmt.Sprintf(s, objects...)).String()
}

func String0(color Color, s any) *Formatted {
	return Format(color, s)
}
