package printz

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// Outputs responsible for keeping reference of outputs writers (example: stdout, file, ...)

type Outputs interface {
	Flusher
	Out() io.Writer
	Err() io.Writer
}

type BasicOutputs struct {
	out, err io.Writer
}

func (o BasicOutputs) Flush() error {
	outs := []io.Writer{o.out, o.err}
	for _, out := range outs {
		f, ok := out.(Flusher)
		if ok {
			err := f.Flush()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (o BasicOutputs) Out() io.Writer {
	return o.out
}

func (o BasicOutputs) Err() io.Writer {
	return o.err
}

func NewOutputs(out, err io.Writer) Outputs {
	return &BasicOutputs{out, err}
}

func NewStandardOutputs() Outputs {
	return NewOutputs(os.Stdout, os.Stderr)
}

func NewStringOutputs() (outW, errW *strings.Builder, outs Outputs) {
	outW = &strings.Builder{}
	errW = &strings.Builder{}
	outs = NewOutputs(outW, errW)
	return
}

type dummyWriter struct {
	io.Writer
}

func (w dummyWriter) Write(p []byte) (int, error) {
	return w.Writer.Write(p)
}

func (w dummyWriter) Flush() error {
	f, ok := w.Writer.(Flusher)
	if ok {
		err := f.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}

func NewBufferedOutputs(outputs Outputs) Outputs {
	// Use a dummyWriter to be able to nest multiple bufio.Writer
	buffOut := bufio.NewWriter(dummyWriter{outputs.Out()})
	buffErr := bufio.NewWriter(dummyWriter{outputs.Err()})
	buffered := BasicOutputs{buffOut, buffErr}
	return buffered
}
