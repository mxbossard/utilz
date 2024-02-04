package cmdz

import (
	"bytes"
	"fmt"
)

type (
	basicPipe struct {
		Executer
		feeder   Executer
		pipeFail bool
	}
)

func (e *basicPipe) Pipe(c Executer) Executer {
	e.Executer = c
	return e
}

func (e *basicPipe) PipeFail(c Executer) Executer {
	e.pipeFail = true
	return e.Pipe(c)
}

func (e *basicPipe) BlockRun() (int, error) {
	if e.Executer == nil {
		return -1, fmt.Errorf("basicPipe don't have a sink !")
	}
	if e.feeder == nil {
		return -1, fmt.Errorf("basicPipe don't have a feed !")
	}

	f := e.feeder
	//f.init()
	originalStdout := f.getConfig().stdout
	originalStdin := f.getConfig().stdin
	b := bytes.Buffer{}

	// Replace configured stdin / stdout temporarilly
	f.SetStdout(&b)
	e.SetInput(&b)
	defer f.SetStdout(originalStdout)
	defer e.SetInput(originalStdin)

	frc, ferr := e.feeder.BlockRun()
	if _, ok := ferr.(failure); ferr != nil && !ok {
		// If feeder error is not a failure{} return it immediately
		return frc, ferr
	}

	if e.pipeFail && frc > 0 {
		// if pipefail enabled and rc > 0 fail shortly
		return frc, ferr
	}

	rc, err := e.Executer.BlockRun()

	return rc, err
}
