package inout

import (
	"io"
	//"fmt"
)

func CopyChannelingErrors(r io.Reader, w io.Writer, errors chan error) {
	_, err := io.Copy(w, r)
	if err != nil {
		errors <- err
	}
}
