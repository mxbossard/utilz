package cmdz

type (
	Executer interface {
		reset()
		fallback(*config)

		String() string
		ReportError() string
		BlockRun() (int, error)
		//Output() (string, error)
		//CombinedOutput() (string, error)
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

	Processer interface {
		Process([]byte) ([]byte, error)
	}

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
)
