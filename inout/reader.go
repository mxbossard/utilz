package inout

import (
	"bytes"
	"fmt"
	"io"
	"log"
)

type RecordingReader struct {
	Nested io.Reader
	Record bytes.Buffer
}

func (r *RecordingReader) Read(b []byte) (n int, err error) {
	if r.Nested == nil {
		log.Fatalf("No nested reader configured in RecordingReader !")
	}

	n, err = r.Nested.Read(b)
	if err != nil {
		return
	}

	if n > 0 {
		w, err2 := r.Record.Write(b[0:n])
		if err2 != nil {
			return n, err2
		}
		if w != n {
			return n, fmt.Errorf("Bad byte count recorded !")
		}
	}

	return
}

func (w *RecordingReader) Reset() {
	w.Record.Reset()
}

func (w RecordingReader) String() string {
	return w.Record.String()
}

type processingReader struct {
	Nested     io.Reader
	Processers []IOProcesser
}

func (r *processingReader) Nest(n io.Reader) {
	r.Nested = n
}

func (r *processingReader) Add(p ...IOProcesser) {
	r.Processers = append(r.Processers, p...)
}

// Return all buffered bytes
func (r *processingReader) AvailableBuffer() ([]byte, error) {
	// TODO
	return nil, nil
}

// Return all buffered bytes processed
/*
func (r *processingReader) ProcessedAvailableBuffer() ([]byte, error) {
	// FIXME: bad implementation probably a bad id to process data which was not processed
	if len(r.Processers) == 0 {
		return nil, nil
	}
	buffer, err := r.Processers[0].AvailableBuffer()
	if err != nil {
		return nil, err
	}
	var n int
	for _, p := range r.Processers[1:] {
		n, err = p.Process(&buffer, len(buffer))
		if err != nil {
			return nil, err
		}
		remains, err := p.AvailableBuffer()
		k := len(remains)
		if err != nil {
			return nil, err
		}
		GrowOrCopy(&buffer, n+k)
		copy(buffer[n:n+k], remains)
	}
	return buffer, err
}
*/

// Reinit Writer
func (r *processingReader) Reset() error {
	for _, p := range r.Processers {
		p.Reset()
	}
	return nil
}

type ProcessingStreamReader struct {
	processingReader
	outBuffer bytes.Buffer
}

// buffer length controll the amount of bytes read from Nester reader.
func (r *ProcessingStreamReader) Read(buffer []byte) (int, error) {
	if r.Nested == nil {
		log.Fatalf("No nested reader configured in ProcessingStreamReader !")
	}

	var n int
	var err error
	if n, err = r.Nested.Read(buffer); n > 0 && n <= len(buffer) && err != io.EOF {
		tmpBuffer := buffer
		for _, prc := range r.Processers {
			n, err = prc.Process(&tmpBuffer, n)
			if err != nil {
				return 0, err
			}
		}
		r.outBuffer.Write(tmpBuffer[0:n])
	}
	if err == io.EOF {
		err = nil
	} else if err != nil {
		return 0, err
	}

	n, err = r.outBuffer.Read(buffer)
	return n, err
}

func NewProcessingStreamReader(nested io.Reader) *ProcessingStreamReader {
	return &ProcessingStreamReader{
		processingReader: processingReader{Nested: nested},
	}
}

type ProcessingBufferReader struct {
	processingReader
	outBuffer bytes.Buffer
}

// Buffer nested reader until nothing left to read before processing
func (r *ProcessingBufferReader) Read(buffer []byte) (int, error) {
	if r.Nested == nil {
		log.Fatalf("No nested reader configured in ProcessingBufferReader !")
	}

	var err error
	var n int
	tmpBuffer := bytes.Buffer{}
	for {
		if tmpBuffer.Len() > 10000000 {
			log.Fatalf("ProcessingBufferReader: reading too long input from nested reader !")
		}

		n, err = r.Nested.Read(buffer)
		if err == io.EOF || n == 0 {
			err = nil
			break
		} else if err != nil {
			return 0, err
		}
		_, err = tmpBuffer.Write(buffer[0:n])
		if err != nil {
			return 0, err
		}
	}

	n = tmpBuffer.Len()
	if n > 0 {
		allBytes := make([]byte, n, n+32)
		copy(allBytes, tmpBuffer.Bytes())
		for _, prc := range r.Processers {
			n, err = prc.Process(&allBytes, n)
			if err != nil {
				return 0, err
			}
		}
		_, err = r.outBuffer.Write(allBytes)
		if err != nil {
			return 0, err
		}
	}
	return r.outBuffer.Read(buffer)
}

func NewProcessingBufferReader(nested io.Reader) *ProcessingBufferReader {
	return &ProcessingBufferReader{
		processingReader: processingReader{Nested: nested},
	}
}

type ReaderProxy struct {
	io.Reader
}

func (w *ReaderProxy) Set(new io.Reader) {
	w.Reader = new
}
