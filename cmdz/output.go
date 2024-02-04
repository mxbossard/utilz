package cmdz

import (
	"fmt"
	"time"
)

type (
	basicOutput struct {
		Executer
	}
)

func (e *basicOutput) Output() ([]byte, error) {
	return nil, fmt.Errorf("basicOutput.Output() not implemented yet !")
}

func (e *basicOutput) OutputString() (string, error) {
	rc, err := e.Executer.ErrorOnFailure(true).BlockRun()
	if err != nil {
		return "", err
	}
	if rc > 0 {
		err = failure{rc, e}
		return "", err
	}
	stdout := []byte(e.StdoutRecord())
	return string(stdout), nil
}

func (e *basicOutput) CombinedOutput() ([]byte, error) {
	return nil, fmt.Errorf("basicOutput.CombinedOutput() not implemented yet !")
}

func (e *basicOutput) CombinedOutputString() (string, error) {
	return e.CombinedOutputs().OutputString()
}

// ----- Override Configurer methods -----
func (e *basicOutput) ErrorOnFailure(ok bool) Outputer {
	e.Executer = e.Executer.ErrorOnFailure(ok)
	return e
}

func (e *basicOutput) CombinedOutputs() Outputer {
	e.Executer = e.Executer.CombinedOutputs()
	return e
}

func (e *basicOutput) Retries(count, delayInMs int) Outputer {
	e.Executer = e.Executer.Retries(count, delayInMs)
	return e
}

func (e *basicOutput) Timeout(duration time.Duration) Outputer {
	e.Executer = e.Executer.Timeout(duration)
	return e
}
