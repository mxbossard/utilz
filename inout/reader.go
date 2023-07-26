package inout

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
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

// Processor which copy input into output and can process errors
// Problems : copy byte arrays is bad ; not fail shortly on error is bad
type IOProcesser = func(input []byte, inErr error) (output []byte, outErr error)

// Processor which enfore the use of one byte buffer ptr only which can be replace by a taller byte buffer
type IOProcesser2 = func(buffer []byte, sizeIn int) (sizeOut int, err error)
type IOProcesser3 = func(buffer *[]byte, sizeIn int) (sizeOut int, err error)

type ProcessingStreamReader struct {
	Nested     io.Reader
	Processers []IOProcesser
	outBuffer  *bytes.Buffer
}

func (r *ProcessingStreamReader) AddProcesser(p IOProcesser) *ProcessingStreamReader {
	r.Processers = append(r.Processers, p)
	return r
}

// buffer length controll the amount of bytes read from Nester reader.
func (r *ProcessingStreamReader) Read(buffer []byte) (int, error) {
	if r.Nested == nil {
		log.Fatalf("No nested reader configured in ProcessingStreamReader !")
	}
	if r.outBuffer == nil {
		r.outBuffer = &bytes.Buffer{}
	}

	var n int
	var err error
	if n, err = r.Nested.Read(buffer); n > 0 && n <= len(buffer) && err != io.EOF {
		tmpBuffer := buffer
		for _, prc := range r.Processers {
			var res []byte
			res, err = prc(tmpBuffer[0:n], err)
			n = len(res)
			tmpBuffer = res
			// tmpBuffer = make([]byte, n)
			// for k := 0; k < n; k++ {
			// 	tmpBuffer[k] = res[k]
			// }
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

type ProcessingBufferReader struct {
	Nested     io.Reader
	Processers []IOProcesser
	outBuffer  *bytes.Buffer
}

func (r *ProcessingBufferReader) AddProcesser(p IOProcesser) *ProcessingBufferReader {
	r.Processers = append(r.Processers, p)
	return r
}

// Buffer nested reader until nothing left to read before processing
func (r *ProcessingBufferReader) Read(buffer []byte) (int, error) {
	if r.Nested == nil {
		log.Fatalf("No nested reader configured in ProcessingBufferReader !")
	}

	var err error
	var n int
	if r.outBuffer == nil {
		r.outBuffer = &bytes.Buffer{}
	}

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

	if tmpBuffer.Len() > 0 {
		for _, prc := range r.Processers {
			var res []byte
			res, err = prc(tmpBuffer.Bytes(), err)
			tmpBuffer.Reset()
			tmpBuffer.Write(res)
		}
	}
	if err != nil {
		return 0, err
	}
	r.outBuffer.Write(tmpBuffer.Bytes())

	return r.outBuffer.Read(buffer)
}

func LineProcesser(proc IOProcesser3) IOProcesser3 {
	return func(buffer *[]byte, sizeIn int) (sizeOut int, err error) {
		var p, size int
		clone := *buffer
		for i, b := range *buffer {
			if b == '\n' {
				if i == 0 {
					size = i
				}
				clone = clone[p : p+size]
				size, err = proc(&clone, sizeOut)
				if err != nil {
					return 0, err
				}
				p = i
				for k := 0; k < size; k++ {
					(*buffer)[sizeOut+k] = clone[k]
				}
				sizeOut += size
			}
		}
		return
	}
}

func LineStringProcesser(proc func(in string) (out string, err error)) IOProcesser2 {
	return func(buffer []byte, sizeIn int) (sizeOut int, err error) {
		var p int
		sb := strings.Builder{}
		for i, b := range buffer {
			if b == '\n' {
				inString := string(buffer[p:i])
				outString, err := proc(inString)
				if err != nil {
					return 0, err
				}
				_, err = sb.WriteString(outString + "\n")
				if err != nil {
					return 0, err
				}
				p = i + 1
			}
		}
		sizeOut = sb.Len()
		// Grow slice if possible
		buffer = buffer[:sizeOut]
		for k, b := range []byte(sb.String()) {
			buffer[k] = b
		}
		return
	}
}
