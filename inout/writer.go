package inout

import (
	"sync"
	"bytes"
	"io"
	"time"
	//"fmt"
)

type Flusher interface {
	Flush() error
}

type CallbackLineWriter struct {
        sync.Mutex
        Flusher
        Callback func(string)
        buffer bytes.Buffer
}

func (w CallbackLineWriter) Flush() error {
        w.Lock()
        defer w.Unlock()
        w.Callback(w.buffer.String())
        w.buffer.Reset()
        return nil
}

func (w CallbackLineWriter) Write(p []byte) (n int, err error) {
        w.Lock()
        defer w.Unlock()
        n, err = w.buffer.Write(p)
        if err != nil {
                return
        }
        //fmt.Println("avant:", w.buffer.String(), w.buffer.Len())
        for line, err := w.buffer.ReadBytes(byte('\n')); err == nil ; {
                //fmt.Println("pendant:", w.buffer.String(), w.buffer.Len())
                n := w.buffer.Len()
                //fmt.Println("n", n, w.Len(), len(line))
                w.buffer.Truncate(n)
                //fmt.Println("apres:", w.String())
                w.Callback(string(line))
                if w.buffer.Len() == 0 {
                        break
                }
        }
        return
}

type ActivityWriter struct {
	Nested io.Writer
	Activity time.Time
}

func (w *ActivityWriter) Write(b []byte) (int, error) {
	t := time.Now()
	w.Activity = t
	return w.Nested.Write(b)
}

