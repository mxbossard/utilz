package display

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
-
*/

type displayer interface {
	Stdout()
	Stderr()
	Flush() error
	Quiet(bool)
	//SetVerbose(model.VerboseLevel)
}

type screen struct {
}

func (*screen) Session(name string) (s *session) {

	return
}

func (*screen) ConfigDisplay(name string) (s *session) {

	return
}

func (*screen) Flush() (err error) {
	return
}

func (*screen) Async() (err error) {
	return
}

type session struct {
}

func (*session) Display(name string) (d *displayer) {

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
