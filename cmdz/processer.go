package cmdz

import "io"

type IOProcesser = func([]byte, error) ([]byte, error)

type ProcessedReader struct {
	io.Reader
	processers []IOProcesser
}

func (r ProcessedReader) Read(p []byte) (int, error) {
	buffer := make([]byte, 256)
	n, err := r.Reader.Read(buffer)
	for _, prc := range r.processers {
		buffer, err = prc(buffer[0:n], err)
	}
	p = buffer
	return len(p), err
}

type ProccessedWriter struct {
	io.Writer
	processers []IOProcesser
}

func (w ProccessedWriter) Write(p []byte) (int, error) {
	buffer := p
	var err error
	for _, prc := range w.processers {
		buffer, err = prc(buffer, err)
	}
	return w.Writer.Write(buffer)
}

type basicOutWriter struct {
	stdout     io.Writer
	stderr     io.Writer
	processers []OutProcesser0
}

func (w basicOutWriter) OutWriter() io.Writer {
	return nil
}

func (w basicOutWriter) ErrWriter() io.Writer {
	return nil
}

func (w basicOutWriter) Flush() error {

}

type basicProcesser struct {
	Executer
	inProcessers  []InProcesser
	outProcessers []OutProcesser
}

func (e *basicProcesser) ProcessIn(pcrs ...InProcesser) Executer {
	e.inProcessers = append(e.inProcessers, pcrs...)
	return e
}

func (e *basicProcesser) ProcessOut(pcrs ...OutProcesser) Executer {
	e.outProcessers = append(e.outProcessers, pcrs...)
	return e
}

func (e *basicProcesser) BlockRun() (int, error) {

	rc, err := e.Executer.BlockRun()

	return rc, err
}

func Proccessed(e Executer) *basicProcesser {
	return &basicProcesser{Executer: e}
}
