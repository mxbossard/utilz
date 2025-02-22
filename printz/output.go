package printz

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"mby.fr/utils/inout"
)

// Outputs responsible for keeping reference of outputs writers (example: stdout, file, ...)

type Outputs interface {
	Flusher
	Out() io.Writer
	Err() io.Writer
	Flushed() bool
	LastPrint() time.Time
}

type BasicOutputs struct {
	*sync.Mutex
	out, err  *inout.CallbackWriter
	flushed   bool
	lastPrint time.Time
}

func (o BasicOutputs) Flush() error {
	o.Lock()
	defer o.Unlock()

	outs := []*inout.CallbackWriter{o.out, o.err}
	for _, out := range outs {
		f, ok := out.Nested.(Flusher)
		if ok {
			err := f.Flush()
			if err != nil {
				return err
			}
		}
	}
	o.flushed = true
	return nil
}

func (o BasicOutputs) Out() io.Writer {
	return o.out
}

func (o BasicOutputs) Err() io.Writer {
	return o.err
}

func (o *BasicOutputs) Flushed() bool {
	o.Lock()
	defer o.Unlock()
	return o.flushed
}

func (o *BasicOutputs) LastPrint() time.Time {
	o.Lock()
	defer o.Unlock()
	return o.lastPrint
}

func callbackWriter(o *BasicOutputs, w io.Writer) (cbw *inout.CallbackWriter) {
	cbw = &inout.CallbackWriter{}
	cbw.Nested = w
	cbw.Callback = func(p []byte) {
		o.Lock()
		defer o.Unlock()
		o.flushed = false
		o.lastPrint = time.Now()
	}
	cbw.CallbackAfter = func() {
		o.Lock()
		defer o.Unlock()
		o.flushed = false
		o.lastPrint = time.Now()
	}
	return
}

func NewOutputs(out, err io.Writer) Outputs {
	bo := &BasicOutputs{Mutex: &sync.Mutex{}, flushed: true}
	bo.out = callbackWriter(bo, out)
	bo.err = callbackWriter(bo, err)
	return bo
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

func NewDiscardingOutputs() Outputs {
	return NewOutputs(io.Discard, io.Discard)
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
	//buffered := BasicOutputs{buffOut, buffErr}
	buffered := NewOutputs(buffOut, buffErr)
	return buffered
}
