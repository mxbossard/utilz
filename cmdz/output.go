package cmdz

type (
	Processed struct {
		Executer
		processors []OutputProcesser
	}
)

func (p Processed) Output() ([]byte, error) {
	return nil, nil
}

func (p Processed) CombinedOutput() ([]byte, error) {
	return nil, nil
}

func (p Processed) Pipe(*Iner) *Outer {
	return nil
}

func (p Processed) PipeFail(*Iner) *Outer {
	return nil
}
