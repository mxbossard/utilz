package cmdz

import (
	"io"

	"mby.fr/utils/promise"
)

type (
	execPromise   = promise.Promise[int]
	execsPromise  = promise.Promise[[]int]
	bytesPromise  = promise.Promise[[]byte]
	stringPromise = promise.Promise[string]

	Runner interface {
	}

	Executer interface {
		reset()
		fallback(*config)

		String() string
		ReportError() string
		BlockRun() (int, error)
		AsyncRun() *execPromise
		ResultCodes() []int

		//StdinRecord() string
		StdoutRecord() string
		StderrRecord() string

		ErrorOnFailure(bool) Executer
		CombineOutputs() Executer
		Retries(count, delayInMs int) Executer

		//Pipe(Executer) Executer
		//PipeFail(Executer) Executer

		//And(Executer) Executer
		//Or(Exeuter) Executer
	}

	Inputer interface {
		Executer
		Input([]byte) error
	}

	InProcesser        = func([]byte) ([]byte, error)
	InStringProcesser  = func(string) (string, error)
	OutProcesser       = func(int, []byte, []byte) ([]byte, error)
	OutStringProcesser = func(int, string, string) (string, error)

	Outputer interface {
		Executer
		Output() ([]byte, error)
		OutputString() (string, error)
		//AsyncOutput() *bytesPromise
		//AsyncOutputString() *stringPromise

		CombinedOutput() ([]byte, error)
		CombinedOutputString() (string, error)
		//AsyncCombinedOutput() *bytesPromise
		//AsyncCombinedOutputString() *stringPromise

		//InProcess([]OutProcesser) Outputer
		//InStringProcess([]OutStringProcesser) Outputer

		OutProcess(...OutProcesser) Outputer
		OutStringProcess(...OutStringProcesser) Outputer
	}

	Formatter[O, E any] interface {
		Format(Outputer) (O, E, error)
	}

	Piper interface {
		Executer
		Pipe(Executer) Piper
		PipeFail(Executer) Piper
	}

	InOutPiper interface {
		Executer
		Pipe(Inputer) Outputer
		PipeFail(Inputer) Outputer
	}

	InOuter interface {
		Executer
		Inputer
		Outputer
	}

	ProcessWriter struct {
		stdOut     io.Writer
		stdErr     io.Writer
		processert []OutProcesser
	}
)

func (w ProcessWriter) OutWriter() io.Writer {
	return nil
}

func (w ProcessWriter) ErrWriter() io.Writer {
	return nil
}
