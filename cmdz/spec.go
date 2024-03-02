package cmdz

import (
	"io"
	"time"

	"mby.fr/utils/inout"
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

	IOProcesser               = inout.IOProcesser
	IOProcesserCallback       = inout.IOProcesserCallback
	StringIOProcesserCallback = inout.StringIOProcesserCallback

	Inner[T any] interface {
		Stdin() io.Reader
		SetInput(stdin io.Reader) T
	}

	Outer[T any] interface {
		Stdout() io.Writer
		Stderr() io.Writer
		SetStdout(io.Writer) T
		SetStderr(io.Writer) T
		SetOutputs(stdout, stderr io.Writer) T
	}

	InOuter[T any] interface {
		Inner[T]
		Outer[T]
	}

	Configurer[T any] interface {
		ErrorOnFailure(bool) T
		Retries(count, delayInMs int) T
		Timeout(duration time.Duration) T
		CombinedOutputs() T
		AddEnv(key, value string) T
		AddEnviron(environ ...string) T
		AddArgs(args ...string) T
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
		init()
		reset()
		fallback(*config)

		BlockRun() (int, error)
		AsyncRun() *execPromise
		ResultCodes() []int
		//ResultCode() int
		ExitCode() int

		StartTimes() []time.Time
		StartTime() time.Time
		Durations() []time.Duration
		Duration() time.Duration
		//Executions() []Executer
		//Execution() Executer
	}

	/*
		Commander interface {
			InOuter[Commander]
			Configurer[Commander]
			Runner
		}
	*/

	Executer interface {
		InOuter[Executer]
		Configurer[Executer]
		Recorder
		Reporter
		Runner

		getConfig() config

		//Pipe(Executer) Executer
		//PipeFail(Executer) Executer

		//And(Executer) Executer
		//Or(Exeuter) Executer
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
	}

	Formatter[O any] interface {
		Configurer[Formatter[O]]
		Format() (O, error)
	}

	Piper interface {
		//Executer
		Pipe(Executer) Executer
		PipeFail(Executer) Executer
	}

	/*
		InOutPiper[T any] interface {
			Executer
			Pipe(Inner[T]) Outer[T]
			PipeFail(Inner[T]) Outer[T]
		}
	*/
)
