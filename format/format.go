package format

import (
	"fmt"
	"strings"

	"mby.fr/utils/ansi"
)

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
	Nested     any
	Formatting bool
	Squashing  bool

	F ansi.Color

	LeftPad, RightPad, TruncateLength int
	TruncateLeft, TruncateRight       bool
	TruncatedReplacement              string
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
	if f.F == ansi.None {
		return nested
	}

	unformatted := Unformat(nested)
	if f.Squashing {
		nested = unformatted
	}

	txtLength := len(unformatted)
	if f.TruncateLength > 0 && txtLength > int(f.TruncateLength) {
		if f.TruncateLeft {
			nested = ansi.TruncateLeftPrefix(nested, int(f.TruncateLength), f.TruncatedReplacement)
		} else if f.TruncateRight {
			nested = ansi.TruncateRightSuffix(nested, int(f.TruncateLength), f.TruncatedReplacement)
		} else {
			nested = ansi.TruncateMid(nested, int(f.TruncateLength), f.TruncatedReplacement)
		}
	}

	if f.LeftPad > 0 {
		spaceCount := f.LeftPad - txtLength
		if spaceCount > 0 {
			nested += strings.Repeat(" ", spaceCount)
		}
	}

	if f.RightPad > 0 {
		spaceCount := f.RightPad - txtLength
		if spaceCount > 0 {
			nested = strings.Repeat(" ", spaceCount) + nested
		}
	}

	return fmt.Sprintf("%v%s%v", f.F, nested, ansi.Reset)
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

func New(color ansi.Color, s any) *Formatted {
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
	return ansi.Unformat(s)
}

func Sprintf(color ansi.Color, s string, objects ...any) string {
	var formattedObjects []any
	for _, obj := range objects {
		switch o := obj.(type) {
		case Formatted:
			formatted := fmt.Sprintf("%s%s%s", ansi.Reset, o.String(), color)
			formattedObjects = append(formattedObjects, formatted)
		case *Formatted:
			formatted := fmt.Sprintf("%s%s%s", ansi.Reset, o.String(), color)
			formattedObjects = append(formattedObjects, formatted)
		default:
			formattedObjects = append(formattedObjects, obj)
		}
	}
	return New(color, fmt.Sprintf(s, formattedObjects...)).Squash(false).String()
}
