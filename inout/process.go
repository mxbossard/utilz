package inout

import (
	"bytes"
	"io"
	"log"
	"strings"
	"sync"
)

type (
	// Processor which copy input into output and can process errors
	// Problems : copy byte arrays is bad ; not fail shortly on error is bad
	//IOProcesser0 = func(input []byte, inErr error) (output []byte, outErr error)

	// Processor which enfore the use of one byte buffer ptr only which can be replace by a taller byte buffer
	//IOProcesser2 = func(buffer []byte, sizeIn int) (sizeOut int, err error)

	// Give controll over the buffer : can grow it or replace it if needed
	IOProcesserCallback = func(buffer *[]byte, sizeIn int) (sizeOut int, err error)

	// Embed sizeIn and sizeout inside buffer by resyzing the slice. Is it perf worthy ???
	SliceProcesserCallback = func(buffer *[]byte) (err error)

	StringIOProcesserCallback = func(in string) (out string, err error)

	IOProcesser interface {
		Process(buffer *[]byte, sizeIn int) (sizeOut int, err error)
		AvailableBuffer() ([]byte, error)
		Reset() error
	}

	IOProcesserReader interface {
		io.Reader
		Nest(io.Reader)
		Add(...IOProcesser)
		Reset() error
	}

	IOProcesserWriter interface {
		io.Writer
		Nest(io.Writer)
		Add(...IOProcesser)
		Flush() error
		Reset() error
	}
)

type BasicIOProcesser struct {
	Callback IOProcesserCallback
}

func (p *BasicIOProcesser) AvailableBuffer() (buffer []byte, err error) {
	// Nothing buffered
	return nil, nil
}

func (p *BasicIOProcesser) Reset() error {
	// Nothing to Reset()
	return nil
}

func (p *BasicIOProcesser) Process(buffer *[]byte, sizeIn int) (sizeOut int, err error) {
	return p.Callback(buffer, sizeIn)
}

type LineIOProcesser struct {
	sync.Mutex
	buffer   bytes.Buffer
	Callback IOProcesserCallback
}

func (p *LineIOProcesser) AvailableBuffer() ([]byte, error) {
	p.Lock()
	remainsLength := p.buffer.Len()
	tmpBuffer := make([]byte, remainsLength, remainsLength+32)
	copy(tmpBuffer, p.buffer.Bytes())
	p.Unlock()
	sizeOut, err := p.Callback(&tmpBuffer, remainsLength)
	if err != nil {
		return nil, err
	}
	GrowOrCopy(&tmpBuffer, sizeOut+1)
	tmpBuffer[sizeOut] = '\n'
	sizeOut++
	return tmpBuffer[0:sizeOut], err
}

func (p *LineIOProcesser) Reset() error {
	p.Lock()
	defer p.Unlock()
	p.buffer.Reset()
	return nil
}

func (p *LineIOProcesser) Process(buffer *[]byte, sizeIn int) (sizeOut int, err error) {
	p.Lock()
	defer p.Unlock()
	var ptr int
	_, err = p.buffer.Write(*buffer)
	if err != nil {
		return 0, err
	}
	tempBuffer := bytes.Buffer{}
	slice := make([]byte, len(*buffer), len(*buffer)+32)
	for i, b := range p.buffer.Bytes() {
		if b == '\n' {
			p.buffer.Read(slice[0 : i-ptr+1])
			size, err := p.Callback(&slice, i-ptr)
			if err != nil {
				return 0, err
			}
			ptr = i + 1
			_, err = tempBuffer.Write(slice[0:size])
			if err != nil {
				return 0, err
			}
			_, err = tempBuffer.Write([]byte("\n"))
			if err != nil {
				return 0, err
			}
		}
	}
	sizeOut = tempBuffer.Len()

	GrowOrCopy(buffer, sizeOut)
	copy((*buffer)[0:], tempBuffer.Bytes())

	return
}

func LineStringProcesser0(proc func(in string) (out string, err error)) IOProcesserCallback {
	return func(buffer *[]byte, sizeIn int) (sizeOut int, err error) {
		var p int
		sb := strings.Builder{}
		for i, b := range *buffer {
			if b == '\n' {
				inString := string((*buffer)[p:i])
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
		GrowOrCopy(buffer, sizeOut)
		copy(*buffer, []byte(sb.String()))
		return
	}
}

func BasicProcesser(callback IOProcesserCallback) IOProcesser {
	return &BasicIOProcesser{callback}
}

func LineProcesser(callback IOProcesserCallback) IOProcesser {
	return &LineIOProcesser{Callback: callback}
}

func StringLineProcesser(callback StringIOProcesserCallback) IOProcesser {
	wrapper := func(buffer *[]byte, sizeIn int) (int, error) {
		in := string((*buffer)[0:sizeIn])
		out, err := callback(in)
		if err != nil {
			return 0, err
		}
		n := len(out)
		log.Printf("str callback ouy: [%s]", out)
		GrowOrCopy(buffer, n)
		copy((*buffer)[0:n], []byte(out))
		log.Printf("bytes callback out: [%v]", *buffer)
		return n, nil
	}
	return &LineIOProcesser{Callback: wrapper}
}

// Resize slice under capacity or copy into bigger slice of supplied capacity
func GrowOrCopy(b *[]byte, sizes ...int) {
	if len(sizes) == 0 {
		log.Fatalf("Must supply slice length !")
	} else if len(sizes) > 2 {
		log.Fatalf("Must supply slice length and capacity nothing else !")
	}

	length := sizes[0]
	if cap(*b) >= length {
		// grow b size
		*b = (*b)[0:length]
	} else {
		var new []byte
		if len(sizes) == 1 {
			new = make([]byte, length, length+32)
		} else {
			capacity := sizes[1]
			new = make([]byte, length, capacity)
		}
		copy(new[0:length], *b)
		*b = new
	}
}

/*
func WindowedGrowOrCopy(b *[]byte, start, end int, sizes ...int) {
	GrowOrCopy(b, sizes...)
	*b = (*b)[start:end]
}
*/
