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
	tmpPath  string
	sessions map[string]*session
	outputs  printz.Outputs
}

func (s *screen) Session(name string, priorityOrder int32) *session {
	if session, ok := s.sessions[name]; ok {
		return session
	}
	session := &session{
		name:          name,
		priorityOrder: priorityOrder,
		outputs:       s.outputs,
		printers:      make(map[string]printz.Printer),
	}
	s.sessions[name] = session

	return session
}

func (*screen) ConfigPrinter(name string) (s *session) {
	// TODO later
	return
}

func (*screen) Flush() (err error) {
	// Flush the display
	// TODO
	return
}

func (*screen) AsyncFlush(timeout time.Duration) (err error) {
	// Launch goroutine wich will continuously flush async display
	// TODO
	return
}

func (*screen) BlockTail(timeout time.Duration) (err error) {
	// Tail async display blocking until end
	// TODO
	return
}

type session struct {
	name            string
	priorityOrder   int32
	started, closed bool

	outputs  printz.Outputs
	printers map[string]printz.Printer
}

func (s *session) Printer(name string) printz.Printer {
	if prtr, ok := s.printers[name]; ok {
		return prtr
	}
	prtr := printz.New(s.outputs)
	s.printers[name] = prtr

	return prtr
}

func (s *session) Start() (err error) {
	s.started = true
	return
}

func (s *session) Close() (err error) {
	s.closed = true
	return
}

func (s *session) Flush() (err error) {
	// TODO
	return
}

func NewScreen(outputs printz.Outputs) *screen {
	return &screen{
		sessions: make(map[string]*session),
		outputs:  printz.NewStandardOutputs(),
	}
}

func NewAsyncScreen(outputs printz.Outputs, tmpPath string) *screen {
	return &screen{
		sessions: make(map[string]*session),
		outputs:  outputs,
		tmpPath:  tmpPath,
	}
}
