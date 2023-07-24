package cmdz

import (
	"io"
)

type (
	Executer interface {
		reset()
		fallback(*config)

		String() string
		ReportError() string
		BlockRun() (int, error)
		//Output() ([]byte, error)
		//OutputString() (string, error)
		//CombinedOutput() ([]byte, error)
		//CombinedOutputString() (string, error)
		AsyncRun() *execPromise
		
		//StdinRecord() string
		StdoutRecord() string
		StderrRecord() string
		
		FailOnError() Executer

		//Pipe(Executer) Executer
		//PipeFail(Executer) Executer
	}

	Inputer interface {
		Input([]byte) error
	}

	Outputer interface {
		Output() ([]byte, error)
		combinedOutput() ([]byte, error)
	}

	OutputProcesser = func(int, []byte, []byte) ([]byte, error)
	OutputStringProcesser = func(int, string, string) (string, error)

	Formatter[O, E any] interface {
		Format(Outputer) (O, E, error)
	}

	Iner interface {
		Executer
		Inputer
	}

	Outer interface {
		Executer
		Outputer
	}

	Piper interface {
		Pipe(*Iner) *Outer
		PipeFail(*Iner) *Outer
	}

	InOuter interface {
		Executer
		Inputer
		Outputer
	}

	ProcessWriter struct {
		stdOut 	   io.Writer
		stdErr 	   io.Writer
		processert []OutputProcesser
	}
)

func (w ProcessWriter) OutWriter() io.Writer {
	return nil
}

func (w ProcessWriter) ErrWriter() io.Writer {
	return nil
}
