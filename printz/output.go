package printz

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/inoutz"
)

// Outputs responsible for keeping reference of outputs writers (example: stdout, file, ...)

type Outputs interface {
	Flusher
	Out() io.Writer
	Err() io.Writer
	Flushed() bool
	LastPrint() time.Time
	Counts() (int64, int64)
}

type BasicOutputs struct {
	*sync.Mutex
	out, err   *inoutz.CallbackWriter
	flushed    bool
	lastPrint  time.Time
	nOut, nErr *int64
}

func (o *BasicOutputs) Flush() error {
	o.Lock()
	defer o.Unlock()

	outs := []*inoutz.CallbackWriter{o.out, o.err}
	for _, out := range outs {
		if out.Nested != nil {
			f, ok := out.Nested.(Flusher)
			if ok {
				err := f.Flush()
				if err != nil {
					return err
				}
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

func (o *BasicOutputs) Counts() (int64, int64) {
	return *o.nOut, *o.nErr
}

func callbackWriter(o *BasicOutputs, w io.Writer, written *int64) (cbw *inoutz.CallbackWriter) {
	cbw = &inoutz.CallbackWriter{}
	cbw.Nested = w
	cbw.Callback = func(p []byte) {
		o.Lock()
		defer o.Unlock()
		o.flushed = false
		o.lastPrint = time.Now()
	}
	cbw.CallbackAfter = func(n int) {
		o.Lock()
		defer o.Unlock()
		o.flushed = false
		o.lastPrint = time.Now()
		*written += int64(n)
	}
	return
}

func NewOutputs(out, err io.Writer) Outputs {
	bo := &BasicOutputs{Mutex: &sync.Mutex{}, flushed: true}
	bo.nOut = new(int64)
	bo.nErr = new(int64)
	bo.out = callbackWriter(bo, out, bo.nOut)
	bo.err = callbackWriter(bo, err, bo.nErr)
	return bo
}

func NewFileOutputs(outFilepath, errFilepath string, flag int, perm fs.FileMode) (Outputs, *os.File, *os.File) {
	out := filez.OpenOrPanic(outFilepath, flag, perm)
	err := filez.OpenOrPanic(errFilepath, flag, perm)
	return NewOutputs(out, err), out, err
}

func lazyCallbackFileWriter(o *BasicOutputs, path string, flag int, perm fs.FileMode, written *int64) (cbw *inoutz.CallbackWriter) {
	cbw = &inoutz.CallbackWriter{}
	cbw.Nested = nil
	cbw.Callback = func(p []byte) {
		o.Lock()
		defer o.Unlock()
		if cbw.Nested == nil {
			dir := filepath.Dir(path)
			err := filez.MkdirAll(dir, filez.DefaultDirPerms)
			if err != nil {
				panic(err)
			}
			cbw.Nested = filez.OpenOrPanic(path, flag, perm)
		}
		o.flushed = false
		o.lastPrint = time.Now()
	}
	cbw.CallbackAfter = func(n int) {
		o.Lock()
		defer o.Unlock()
		o.flushed = false
		o.lastPrint = time.Now()
		*written += int64(n)
	}
	return
}

// Outputs sinking in files which are opened on first write.
func NewLazyFileOutputs(outFilepath, errFilepath string, flag int, perm fs.FileMode) Outputs {
	bo := &BasicOutputs{Mutex: &sync.Mutex{}, flushed: true}
	bo.nOut = new(int64)
	bo.nErr = new(int64)
	bo.out = lazyCallbackFileWriter(bo, outFilepath, flag, perm, bo.nOut)
	bo.err = lazyCallbackFileWriter(bo, errFilepath, flag, perm, bo.nErr)
	return bo
}

func NewStandardOutputs() Outputs {
	return NewOutputs(os.Stdout, os.Stderr)
}

func NewStringBuilderOutputs() (outW, errW *strings.Builder, outs Outputs) {
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
