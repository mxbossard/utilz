package inoutz

import (
	"io"
	"iter"
)

type splitIterator struct {
	io.Reader
	splitChar  byte
	bufferSize int
}

// Return byte count pushed to a callback
// Buffer MUST be greater than greatest word between two splitChar
func (r *splitIterator) All(errChan chan error) iter.Seq2[int, []byte] {
	return func(yield func(int, []byte) bool) {
		buf := make([]byte, r.bufferSize)
		var k int

		var err error
		var p, lastDelim, pos int
		eof := false
		n := -1
		for !eof {
			n, err = r.Reader.Read(buf[p:])
			eof = n+p < len(buf)
			// fmt.Printf("just read: %s; eof: %v\n", string(buf[p:p+n]), eof)
			lastDelim = -1
			for i := range len(buf) {
				if buf[i] == r.splitChar {
					// fmt.Printf("buf: %s\n", string(buf))
					// fmt.Printf("yield: #%d (%s) [%d:%d]\n", pos, string(buf[lastDelim+1:i]), lastDelim+1, i)
					if !yield(pos, buf[lastDelim+1:i]) {
						return
					}
					k += (i - lastDelim)
					lastDelim = i
					pos++
				}
			}
			if err == io.EOF || eof {
				break
			} else if err != nil {
				errChan <- err
				break
			} else if lastDelim == -1 {
				panic("buffer not big enough")
			}

			// Copy buffer tail on head then continue reading
			p = 0
			// fmt.Printf("copying: %s at head (eof: %v)\n", string(buf[lastDelim+1:]), eof)
			for i := lastDelim + 1; i < len(buf); i++ {
				buf[p] = buf[i]
				p++
			}
		}

		if lastDelim < len(buf)-1 {
			// Stream does not end with delim
			// fmt.Printf("buf remains: %s\n", string(buf[lastDelim+1:n+p]))
			// fmt.Printf("yield: #%d (%s) [%d:%d]\n", pos, string(buf[lastDelim+1:n+p]), lastDelim+1, n+p)
			if !yield(pos, buf[lastDelim+1:n+p]) {
				return
			}
			k += (n - lastDelim - 1)
		}
	}
}

func NewSplitIterator(r io.Reader, split byte, bufferSize int) *splitIterator {
	return &splitIterator{
		Reader:     r,
		splitChar:  split,
		bufferSize: bufferSize,
	}
}
