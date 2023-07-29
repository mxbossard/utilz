package cmdz

type (
	basicFormat[O, E any] struct {
		Executer
		outFormatter func(int, []byte, []byte, error) (O, E)
	}
)

func (f *basicFormat[O, E]) Format() (O, E) {
	rc, err := f.Executer.BlockRun()
	stdout := []byte(f.Executer.StdoutRecord())
	stderr := []byte(f.Executer.StderrRecord())
	o, e := f.outFormatter(rc, stdout, stderr, err)
	return o, e
}

// ----- Override Configurer methods -----
func (e *basicFormat[O, E]) ErrorOnFailure(ok bool) Formatter[O, E] {
	e.Executer = e.Executer.ErrorOnFailure(ok)
	return e
}

func (e *basicFormat[O, E]) CombinedOutputs() Formatter[O, E] {
	e.Executer = e.Executer.CombinedOutputs()
	return e
}

func (e *basicFormat[O, E]) Retries(count, delayInMs int) Formatter[O, E] {
	e.Executer = e.Executer.Retries(count, delayInMs)
	return e
}

func (e *basicFormat[O, E]) Timeout(delayInMs int) Formatter[O, E] {
	e.Executer = e.Executer.Timeout(delayInMs)
	return e
}

func Formatted[O, E any](e Executer, f func(int, []byte, []byte, error) (O, E)) *basicFormat[O, E] {
	return &basicFormat[O, E]{Executer: e, outFormatter: f}
}

func FormattedCmd[O, E any](f func(int, []byte, []byte, error) (O, E), binary string, args ...string) *basicFormat[O, E] {
	c := Cmd(binary, args...)
	return Formatted(c, f)
}
