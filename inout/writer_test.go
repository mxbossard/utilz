package inout

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallbackLineWriter(t *testing.T) {
	var flushed []string
	callback := func(msg string) {
		flushed = append(flushed, msg)
	}
	w := CallbackLineWriter{Callback: callback}

	lines := "foo\nbar\nbaz"
	n, err := w.Write([]byte(lines))
	require.NoError(t, err)
	assert.Equal(t, len(lines), n)
	assert.Len(t, flushed, 2)
	assert.Contains(t, flushed, "foo\n")
	assert.Contains(t, flushed, "bar\n")

	err = w.Flush()
	require.NoError(t, err)
	assert.Len(t, flushed, 3)
	assert.Contains(t, flushed, "foo\n")
	assert.Contains(t, flushed, "bar\n")
	assert.Contains(t, flushed, "baz")

	n, err = w.Write([]byte("pif\n"))
	require.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Len(t, flushed, 4)
	assert.Contains(t, flushed, "foo\n")
	assert.Contains(t, flushed, "bar\n")
	assert.Contains(t, flushed, "baz")
	assert.Contains(t, flushed, "pif\n")

	err = w.Flush()
	require.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Len(t, flushed, 4)
	assert.Contains(t, flushed, "foo\n")
	assert.Contains(t, flushed, "bar\n")
	assert.Contains(t, flushed, "baz")
	assert.Contains(t, flushed, "pif\n")

}

func TestRecordingWriter(t *testing.T) {
	nested := strings.Builder{}
	rw := RecordingWriter{}
	rw.Nested = &nested

	expectedMsg1 := "foo"
	expectedMsg2 := "bar"
	expectedMsg3 := "baz"

	n, err := rw.Write([]byte(expectedMsg1))
	require.NoError(t, err)
	assert.Equal(t, len(expectedMsg1), n)
	assert.Equal(t, expectedMsg1, nested.String())
	assert.Equal(t, expectedMsg1, rw.String())

	n, err = rw.Write([]byte(expectedMsg2))
	require.NoError(t, err)
	assert.Equal(t, len(expectedMsg2), n)
	assert.Equal(t, expectedMsg1+expectedMsg2, nested.String())
	assert.Equal(t, expectedMsg1+expectedMsg2, rw.String())

	rw.Reset()
	n, err = rw.Write([]byte(expectedMsg3))
	require.NoError(t, err)
	assert.Equal(t, len(expectedMsg3), n)
	assert.Equal(t, expectedMsg1+expectedMsg2+expectedMsg3, nested.String())
	assert.Equal(t, expectedMsg3, rw.String())
}

func TestProcessingStreamWriter_NoProcesser(t *testing.T) {
	expectedMsg1 := "foo"
	expecterdBytes1 := []byte(expectedMsg1)
	nested := &bytes.Buffer{}

	pr := ProcessingStreamWriter{Nested: nested}

	n, err := pr.Write(expecterdBytes1)
	require.NoError(t, err)
	assert.Equal(t, len(expecterdBytes1), n)
	assert.Equal(t, expecterdBytes1, nested.Bytes())
}

func TestProcessingStreamWriter(t *testing.T) {
	expectedMsg1 := "foo"
	expectedBytes1 := []byte(expectedMsg1)
	expectedMsg2 := "bar"
	expectedBytes2 := []byte(expectedMsg2)
	expectedMsg3 := "END"
	expectedBytes3 := []byte(expectedMsg3)
	errProcesser := func(in []byte, err error) ([]byte, error) {
		return nil, fmt.Errorf("errProcesserError")
	}
	barProcesser := func(in []byte, err error) ([]byte, error) {
		return expectedBytes2, nil
	}
	endProcesser := func(in []byte, err error) ([]byte, error) {
		out := append(in, expectedBytes3...)
		return out, nil
	}

	nested := &bytes.Buffer{}
	pr := ProcessingStreamWriter{Nested: nested}
	pr.AddProcesser(errProcesser)
	n, err := pr.Write(expectedBytes1)
	require.Error(t, err)
	assert.Equal(t, 0, n)

	nested = &bytes.Buffer{}
	pr = ProcessingStreamWriter{Nested: nested}
	pr.AddProcesser(barProcesser)
	n, err = pr.Write(expectedBytes1)
	require.NoError(t, err)
	assert.Equal(t, len(expectedBytes2), n)
	assert.Equal(t, expectedBytes2, nested.Bytes())

	nested = &bytes.Buffer{}
	pr = ProcessingStreamWriter{Nested: nested}
	pr.AddProcesser(endProcesser)
	n, err = pr.Write(expectedBytes1)
	require.NoError(t, err)
	assert.Equal(t, len(expectedBytes1)+len(expectedBytes3), n)
	assert.Equal(t, append(expectedBytes1, expectedBytes3...), nested.Bytes())
}

func TestProcessingStreamWriter_LongerThanBuffer(t *testing.T) {
	expectedMsg1 := "foofoofoo"
	expectedBytes1 := []byte(expectedMsg1)
	expectedMsg2 := "barbarbar"
	expectedBytes2 := []byte(expectedMsg2)
	expectedMsg3 := "END"
	expectedBytes3 := []byte(expectedMsg3)

	endProcesser := func(in []byte, err error) ([]byte, error) {
		out := append(in, expectedBytes3...)
		return out, nil
	}

	nested := &bytes.Buffer{}
	pr := ProcessingStreamWriter{Nested: nested}
	pr.AddProcesser(endProcesser)
	n, err := pr.Write(expectedBytes1)
	require.NoError(t, err)
	assert.Equal(t, 12, n)
	assert.Equal(t, "foofoofooEND", nested.String())
	n, err = pr.Write(expectedBytes2)
	require.NoError(t, err)
	assert.Equal(t, 12, n)
	require.Equal(t, "foofoofooENDbarbarbarEND", nested.String())
	n, err = pr.Write(nil)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, "foofoofooENDbarbarbarEND", nested.String())
}

func TestProcessingBufferWriter_NoProcesser(t *testing.T) {
	expectedMsg1 := "foo"
	expecterdBytes1 := []byte(expectedMsg1)
	nested := &bytes.Buffer{}

	pr := ProcessingBufferWriter{Nested: nested}

	n, err := pr.Write(expecterdBytes1)
	require.NoError(t, err)
	assert.Equal(t, len(expecterdBytes1), n)
	assert.Equal(t, expecterdBytes1, nested.Bytes())
}

func TestProcessingBufferWriter(t *testing.T) {
	expectedMsg1 := "foo"
	expectedBytes1 := []byte(expectedMsg1)
	expectedMsg2 := "bar"
	expectedBytes2 := []byte(expectedMsg2)
	expectedMsg3 := "END"
	expectedBytes3 := []byte(expectedMsg3)
	errProcesser := func(in []byte, err error) ([]byte, error) {
		return nil, fmt.Errorf("errProcesserError")
	}
	barProcesser := func(in []byte, err error) ([]byte, error) {
		return expectedBytes2, nil
	}
	endProcesser := func(in []byte, err error) ([]byte, error) {
		out := append(in, expectedBytes3...)
		return out, nil
	}

	nested := &bytes.Buffer{}
	pr := ProcessingBufferWriter{Nested: nested}
	pr.AddProcesser(errProcesser)
	n, err := pr.Write(expectedBytes1)
	require.Error(t, err)
	assert.Equal(t, 0, n)

	nested = &bytes.Buffer{}
	pr = ProcessingBufferWriter{Nested: nested}
	pr.AddProcesser(barProcesser)
	n, err = pr.Write(expectedBytes1)
	require.NoError(t, err)
	assert.Equal(t, len(expectedBytes2), n)
	assert.Equal(t, expectedBytes2, nested.Bytes())

	nested = &bytes.Buffer{}
	pr = ProcessingBufferWriter{Nested: nested}
	pr.AddProcesser(endProcesser)
	n, err = pr.Write(expectedBytes1)
	require.NoError(t, err)
	assert.Equal(t, len(expectedBytes1)+len(expectedBytes3), n)
	assert.Equal(t, append(expectedBytes1, expectedBytes3...), nested.Bytes())
}

func TestProcessingBufferWriter_LongerThanBuffer(t *testing.T) {
	expectedMsg1 := "foofoofoo"
	expectedBytes1 := []byte(expectedMsg1)
	expectedMsg2 := "barbarbar"
	expectedBytes2 := []byte(expectedMsg2)
	expectedMsg3 := "END"
	expectedBytes3 := []byte(expectedMsg3)
	//expectedMsg4 := "bazbazbaz"
	//expectedBytes4 := []byte(expectedMsg4)
	//EOL := []byte{0}

	endProcesser := func(in []byte, err error) ([]byte, error) {
		out := append(in, expectedBytes3...)
		return out, nil
	}

	nested := &bytes.Buffer{}
	pr := ProcessingBufferWriter{Nested: nested}
	pr.AddProcesser(endProcesser)
	n, err := pr.Write(expectedBytes1)
	require.NoError(t, err)
	assert.Equal(t, 12, n)
	assert.Equal(t, expectedMsg1+expectedMsg3, nested.String())
	n, err = pr.Write(expectedBytes2)
	require.NoError(t, err)
	assert.Equal(t, 12, n)
	assert.Equal(t, expectedMsg1+expectedMsg3+expectedMsg2+expectedMsg3, nested.String())
	n, err = pr.Write(nil)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, expectedMsg1+expectedMsg3+expectedMsg2+expectedMsg3, nested.String())

	n, err = pr.Write([]byte(""))
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, expectedMsg1+expectedMsg3+expectedMsg2+expectedMsg3, nested.String())

}
