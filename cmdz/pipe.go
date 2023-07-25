package cmdz

type (
	basicPipe struct {
		Executer
		feeder   Executer
		pipeFail bool
	}
)

func (e *basicPipe) Pipe(c Executer) Piper {
	e.Executer = c
	return e
}

func (e *basicPipe) PipeFail(c Executer) Piper {
	e.pipeFail = true
	return e.Pipe(c)
}

/*
func (e *basicPipe) BlockRun() (int, error) {
	if e.Executer == nil {
		return -1, fmt.Errorf("basicPipe don't have a sink !")
	}
	if e.feeder == nil {
		return -1, fmt.Errorf("basicPipe don't have a feed !")
	}

	f := e.feeder
	originalStdout := f.config.stdout
	originalStdin := f.config.stdin
	b := bytes.Buffer{}

	// Replace configured stdin / stdout temporarilly
	f.pipedOutput = true
	f.config.stdout = &b
	//f.stdoutRecord.Nested = &b

	e.pipedInput = true
	e.config.stdin = &b
	//e.stdinRecord.Nested = &b

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

	f.config.stdout = originalStdout
	e.config.stdin = originalStdin

	return rc, err
}
*/

func Pipable(i Executer) Piper {
	return &basicPipe{feeder: i}
}

func Pipe(i, o Executer) Piper {
	return &basicPipe{Executer: o, feeder: i}
}

/*
func InOutPipe(i Inputer, o Outputer) Piper {
	return &basicPipe{Executer: o, feeder: i}
}
*/
