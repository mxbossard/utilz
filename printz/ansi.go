package printz

import (
	"fmt"
	"log"

	"github.com/mxbossard/utilz/anzi"
)

// ANSI formatting for content
type ansiFormatted struct {
	Format            anzi.Color
	Content           interface{}
	LeftPad, RightPad int
}

func NewAnsi(format anzi.Color, content any) (f ansiFormatted) {
	f.Format = format
	f.Content = content
	return
}

func NewAnsiLeftPadded(format anzi.Color, content any, padding int) (f ansiFormatted) {
	f.Format = format
	f.Content = content
	f.LeftPad = padding
	return
}

func NewAnsiRightPadded(format anzi.Color, content any, padding int) (f ansiFormatted) {
	f.Format = format
	f.Content = content
	f.RightPad = padding
	return
}

func ansiFormatParams(color anzi.Color, params ...any) (formattedParams []any) {
	for _, p := range params {
		if f, ok := p.(ansiFormatted); ok {
			s, err := stringify(f)
			if err != nil {
				log.Fatal(err)
			}
			s += string(color)
			formattedParams = append(formattedParams, s)
		} else {
			formattedParams = append(formattedParams, p)
		}
	}
	return
}

func Sprintf(format string, params ...any) string {
	return fmt.Sprintf(format, ansiFormatParams("", params...)...)
}

func SColoredPrintf(color anzi.Color, format string, params ...any) string {
	return fmt.Sprintf(format, ansiFormatParams(color, params...)...)
}
