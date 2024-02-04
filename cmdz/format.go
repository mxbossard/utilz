package cmdz

import "time"

type (
	basicFormat[O any] struct {
		Executer
		outFormatter func(int, []byte, []byte) (O, error)
	}
)

func (f *basicFormat[O]) Format() (out O, err error) {
	rc, err := f.Executer.BlockRun()
	if err != nil {
		return out, err
	}
	stdout := []byte(f.Executer.StdoutRecord())
	stderr := []byte(f.Executer.StderrRecord())
	o, err := f.outFormatter(rc, stdout, stderr)
	return o, err
}

// ----- Override Configurer methods -----
func (e *basicFormat[O]) ErrorOnFailure(ok bool) Formatter[O] {
	e.Executer = e.Executer.ErrorOnFailure(ok)
	return e
}

func (e *basicFormat[O]) CombinedOutputs() Formatter[O] {
	e.Executer = e.Executer.CombinedOutputs()
	return e
}

func (e *basicFormat[O]) Retries(count, delayInMs int) Formatter[O] {
	e.Executer = e.Executer.Retries(count, delayInMs)
	return e
}

func (e *basicFormat[O]) Timeout(duration time.Duration) Formatter[O] {
	e.Executer = e.Executer.Timeout(duration)
	return e
}
