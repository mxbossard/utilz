package inout

import (
	"fmt"
	"io"
	"strings"
)

type RecordingReader struct {
	Nested io.Reader
	Record strings.Builder
}

func (w *RecordingReader) Read(b []byte) (n int, err error) {
	if w.Nested != nil {
		n, err = w.Nested.Read(b)
		if err != nil {
			return
		}

		if n > 0 {
			w, err2 := w.Record.Write(b[0:n])
			if err2 != nil {
				return n, err2
			}
			if w != n {
				return n, fmt.Errorf("Bad byte count recorded !")
			}
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
