package display

import (
	"time"

	"mby.fr/utils/printz"
)

/**
- A screen is a structure which abstract the whole screen.
  - a screen may be splitted into multiple parts which are displayed simultaneously.
  - for now only one part is supported to be displayed at once
- A session is a structure which permit to print data on the screen.
  - Multiple sessions can coexist simultaneously
  - Only one sessions is elected at once to be displayed until closed
- A displayer is the interface used to print data on the screen.
  - stdout & stderr

## Features
- A displayer can be configured with Formatters
- The configuration should be stored à Screen level

## Ideas
- reuse printz.Printer : how to format ?
- screen.InitPrinter(name)
- screen.ConfigPrinter(name, formaters)
- change format.Formatter for signature: Format(string|Stringer...) Formatted
- use templates ?
*/

type displayer interface {
	// Replaced by printz.Printer ?
	Stdout()
	Stderr()
	Flush() error
	//Quiet(bool)
}

type screen struct {
}

func (*screen) Session(name string, priority int32) (s *session) {

	return
}

func (*screen) ConfigPrinter(name string) (s *session) {

	return
}

func (*screen) AsyncFlush(timeout time.Duration) (err error) {
	// Launch goroutine wich will continuously flush async display
	return
}

func (*screen) BlockTail(timeout time.Duration) (err error) {
	// Tail async display blocking until end
	return
}

type session struct {
}

func (*session) Printer(name string) (p printz.Printer) {

	return
}

func (*session) Close() (err error) {

	return
}

func (*session) Flush() (err error) {
	return
}

func NewScreen() *screen {
	return &screen{}
}

func NewAsyncScreen(tmpPath string) *screen {
	return &screen{}
}
