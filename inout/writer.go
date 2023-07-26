package inout

import (
	"bytes"
	"io"
	"log"
	"sync"
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
	buffer   bytes.Buffer
}

func (w *CallbackLineWriter) Flush() error {
	w.Lock()
	defer w.Unlock()
	if w.buffer.Len() > 0 {
		w.Callback(w.buffer.String())
		w.buffer.Reset()
	}
	return nil
}

func (w *CallbackLineWriter) Write(p []byte) (n int, err error) {
	w.Lock()
	defer w.Unlock()
	n, err = w.buffer.Write(p)
	if err != nil {
		return
	}
	for {
		if w.buffer.Len() == 0 {
			break
		}
		var line []byte
		line, err = w.buffer.ReadBytes(byte('\n'))
		if err == io.EOF && len(line) > 0 {
			// write back line in buffer
			_, err = w.buffer.Write(line)
			break
		}
		if err != nil {
			return 0, err
		}
		w.Callback(string(line))
	}
	return
}

type ActivityWriter struct {
	Nested   io.Writer
	Activity time.Time
}

func (w *ActivityWriter) Write(b []byte) (int, error) {
	t := time.Now()
	w.Activity = t
	return w.Nested.Write(b)
}

type RecordingWriter struct {
	Nested io.Writer
	Record bytes.Buffer
}

func (w *RecordingWriter) Write(b []byte) (int, error) {
	_, err := w.Record.Write(b)
	if err != nil {
		return 0, err
	}
	if w.Nested != nil {
		return w.Nested.Write(b)
	}
	return 0, nil
}

func (w *RecordingWriter) Reset() {
	w.Record.Reset()
}

func (w RecordingWriter) String() string {
	return w.Record.String()
}

type ProcessingStreamWriter struct {
	Nested     io.Writer
	Processers []IOProcesser
	outBuffer  *bytes.Buffer
}

func (w *ProcessingStreamWriter) AddProcesser(p IOProcesser) *ProcessingStreamWriter {
	w.Processers = append(w.Processers, p)
	return w
}

// Naively apply each IOProcessor on buffer successively
func (w *ProcessingStreamWriter) Write(buffer []byte) (int, error) {
	if w.Nested == nil {
		log.Fatalf("No nested writer configured in ProcessingStreamWriter !")
	}

	if w.outBuffer == nil {
		w.outBuffer = &bytes.Buffer{}
	}
	var err error
	tmpBuffer := buffer
	n := len(buffer)
	if n > 0 {
		for _, prc := range w.Processers {
			var res []byte
			res, err = prc(tmpBuffer, err)
			n = len(res)
			tmpBuffer = res
		}
		if err != nil {
			return 0, err
		}
		w.Nested.Write(tmpBuffer[0:n])
	}
	return n, err
}

type ProcessingBufferWriter struct {
	Nested     io.Writer
	Processers []IOProcesser
	inBuffer   *bytes.Buffer
	delimiter  *byte
}

func (w *ProcessingBufferWriter) AddProcesser(p IOProcesser) *ProcessingBufferWriter {
	w.Processers = append(w.Processers, p)
	return w
}

// Buffer all writes until delimiter if present before processing and then writing to nested
func (w *ProcessingBufferWriter) Write(buffer []byte) (int, error) {
	if w.Nested == nil {
		log.Fatalf("No nested writer configured in ProcessingBufferWriter !")
	}
	if w.inBuffer == nil {
		w.inBuffer = &bytes.Buffer{}
	}

	_, err := w.inBuffer.Write(buffer)
	if err != nil {
		return 0, err
	}

	var tmpBuffer []byte
	outBuffer := bytes.Buffer{}
	if w.inBuffer.Len() > 0 {
		doLoop := true
		for doLoop {
			if w.delimiter != nil {
				tmpBuffer, err = w.inBuffer.ReadBytes(*w.delimiter)
			} else {
				tmpBuffer = buffer
				doLoop = false
			}
			if err == io.EOF {
				err = nil
				break
			} else if err != nil {
				return 0, err
			}
			n := len(tmpBuffer)
			if n > 0 {
				for _, prc := range w.Processers {
					var res []byte
					res, err = prc(tmpBuffer, err)
					if err != nil {
						return 0, err
					}
					n = len(res)
					tmpBuffer = res
				}
			}
			outBuffer.Write(tmpBuffer[0:n])
		}
		if err == io.EOF {
			// Delimiter not found put back read bytes into buffer
			_, err = w.inBuffer.Write(tmpBuffer)
			if err != nil {
				return 0, err
			}
		} else if err != nil {
			return 0, err
		}
	}

	if outBuffer.Len() > 0 {
		return w.Nested.Write(outBuffer.Bytes())
	}

	return 0, nil
}
