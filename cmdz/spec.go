package cmdz

import (
	"io"

	"mby.fr/utils/promise"
)

// []byte                       => |Executer|     => (int, []byte, []byte, error)
// ([]byte, error)              => |InProcesser|  => ([]byte, error)
// (int, []byte, []byte, error) => |OutProcesser| => (int, []byte, []byte, error)
// []byte                       => |Inputer|
//                                 |Outputer|     => ([]byte, error)
// (int, []byte, []byte, error) => |OutFormatter| => (O, E)
//                                 |Formatter|    => (O, E)

type (
	execPromise   = promise.Promise[int]
	execsPromise  = promise.Promise[[]int]
	bytesPromise  = promise.Promise[[]byte]
	stringPromise = promise.Promise[string]

	Inner[T any] interface {
		Stdin() io.Reader
		Input(stdin io.Reader) T
	}

	Outer[T any] interface {
		Stdout() io.Writer
		Stderr() io.Writer
		Outputs(stdout, stderr io.Writer) T
	}

	InOuter[T any] interface {
		Inner[T]
		Outer[T]
	}

	Configurer[T any] interface {
		ErrorOnFailure(bool) T
		Retries(count, delayInMs int) T
		Timeout(delayInMs int) T
		CombinedOutputs() T
	}

	Recorder interface {
		StdinRecord() string
		StdoutRecord() string
		StderrRecord() string
	}

	Reporter interface {
		String() string
		ReportError() string
	}

	Runner interface {
		reset()
		fallback(*config)

		BlockRun() (int, error)
		AsyncRun() *execPromise
		ResultCodes() []int
	}

	Commander interface {
		InOuter[Commander]
		Configurer[Commander]
		Runner
	}

	Executer interface {
		InOuter[Executer]
		Configurer[Executer]
		Recorder
		Reporter
		Runner

		//Pipe(Executer) Executer
		//PipeFail(Executer) Executer

		//And(Executer) Executer
		//Or(Exeuter) Executer
	}

	Inputer interface {
		//Executer
		Input([]byte) error
	}

	InProcesser0       = func([]byte, error) ([]byte, error)
	InProcesser        = func([]byte) ([]byte, error)
	InStringProcesser  = func(string) (string, error)
	OutProcesser       = func(int, []byte, []byte) ([]byte, error)
	OutProcesser0      = func(int, []byte, []byte, error) (int, []byte, []byte, error)
	OutStringProcesser = func(int, string, string) (string, error)

	Outputer interface {
		Configurer[Outputer]
		Recorder

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
		Configurer[Formatter[O, E]]
		Format() (O, E)
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
)
