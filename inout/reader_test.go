package inout

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordingReader(t *testing.T) {
	expectedMsg1 := "foo"
	nested := strings.NewReader(expectedMsg1)
	rr := RecordingReader{}
	rr.Nested = nested

	buffer := make([]byte, len(expectedMsg1)+5)
	n, err := rr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, len(expectedMsg1), n)
	assert.Equal(t, expectedMsg1, string(buffer[0:n]))
	assert.Equal(t, expectedMsg1, rr.String())
}

func TestProcessingStreamReader_NoProcesser(t *testing.T) {
	expectedMsg1 := "foo"
	expecterdBytes1 := []byte(expectedMsg1)
	nested := strings.NewReader(expectedMsg1)
	buffer := make([]byte, len(expectedMsg1)+5)

	pr := ProcessingStreamReader{Nested: nested}

	n, err := pr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, len(expecterdBytes1), n)
	assert.Equal(t, expecterdBytes1, buffer[0:n])
}

func TestProcessingStreamReader(t *testing.T) {
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
	buffer := make([]byte, len(expectedMsg1)+5)

	nested := strings.NewReader(expectedMsg1)
	pr := ProcessingStreamReader{Nested: nested}
	pr.AddProcesser(errProcesser)
	n, err := pr.Read(buffer)
	require.Error(t, err)
	assert.Equal(t, 0, n)

	nested = strings.NewReader(expectedMsg1)
	pr = ProcessingStreamReader{Nested: nested}
	pr.AddProcesser(barProcesser)
	n, err = pr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, len(expectedBytes2), n)
	assert.Equal(t, expectedBytes2, buffer[0:n])

	nested = strings.NewReader(expectedMsg1)
	pr = ProcessingStreamReader{Nested: nested}
	pr.AddProcesser(endProcesser)
	n, err = pr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, len(expectedBytes1)+len(expectedBytes3), n)
	assert.Equal(t, append(expectedBytes1, expectedBytes3...), buffer[0:n])
}

func TestProcessingStreamReader_LongerThanBuffer(t *testing.T) {
	expectedMsg1 := "foofoofoobarbarbar"
	expectedMsg3 := "END"
	expectedBytes3 := []byte(expectedMsg3)

	endProcesser := func(in []byte, err error) ([]byte, error) {
		out := append(in, expectedBytes3...)
		return out, nil
	}
	buffer := make([]byte, 9)

	nested := strings.NewReader(expectedMsg1)
	pr := ProcessingStreamReader{Nested: nested}
	pr.AddProcesser(endProcesser)
	n, err := pr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, 9, n)
	assert.Equal(t, "foofoofoo", string(buffer))
	n, err = pr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, 9, n)
	require.Equal(t, "ENDbarbar", string(buffer))
	n, err = pr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.Equal(t, "barEND", string(buffer[0:n]))
	n, err = pr.Read(buffer)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.ErrorIs(t, err, io.EOF)
}

func TestProcessingBufferReader_NoProcesser(t *testing.T) {
	expectedMsg1 := "foo"
	expecterdBytes1 := []byte(expectedMsg1)
	nested := strings.NewReader(expectedMsg1)
	buffer := make([]byte, len(expectedMsg1)+5)

	pr := ProcessingBufferReader{Nested: nested}

	n, err := pr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, len(expecterdBytes1), n)
	assert.Equal(t, expecterdBytes1, buffer[0:n])
}

func TestProcessingBufferReader(t *testing.T) {
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
	buffer := make([]byte, len(expectedMsg1)+16)

	nested := strings.NewReader(expectedMsg1)
	pr := ProcessingBufferReader{Nested: nested}
	pr.AddProcesser(errProcesser)
	n, err := pr.Read(buffer)
	require.Error(t, err)
	assert.Equal(t, 0, n)

	nested = strings.NewReader(expectedMsg1)
	pr = ProcessingBufferReader{Nested: nested}
	pr.AddProcesser(barProcesser)
	n, err = pr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, len(expectedBytes2), n)
	assert.Equal(t, expectedBytes2, buffer[0:n])

	nested = strings.NewReader(expectedMsg1)
	pr = ProcessingBufferReader{Nested: nested}
	pr.AddProcesser(endProcesser)
	n, err = pr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, len(expectedBytes1)+len(expectedBytes3), n)
	assert.Equal(t, append(expectedBytes1, expectedBytes3...), buffer[0:n])
}

func TestProcessingBufferReader_LongerThanBuffer(t *testing.T) {
	expectedMsg1 := "foofoofoobarbarbar"
	expectedMsg3 := "END"
	expectedBytes3 := []byte(expectedMsg3)

	endProcesser := func(in []byte, err error) ([]byte, error) {
		out := append(in, expectedBytes3...)
		return out, nil
	}
	buffer := make([]byte, 9)

	nested := strings.NewReader(expectedMsg1)
	pr := ProcessingBufferReader{Nested: nested}
	pr.AddProcesser(endProcesser)
	n, err := pr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, 9, n)
	assert.Equal(t, expectedMsg1[0:9], string(buffer))
	n, err = pr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, 9, n)
	assert.Equal(t, expectedMsg1[9:18], string(buffer))
	n, err = pr.Read(buffer)
	require.NoError(t, err)
	require.Equal(t, len(expectedBytes3), n)
	assert.Equal(t, expectedBytes3, buffer[0:3])

}

func TestLineProcesser(t *testing.T) {
	addEndProc := func(buffer *[]byte, sizeIn int) (sizeOut int, err error) {
		*buffer = append(*buffer, []byte("END")...)
		return sizeIn + 3, nil
	}

	p := LineProcesser(addEndProc)

	buffer := []byte("foo\nbar\nbaz")
	sizeOut, err := p(&buffer, 11)
	require.NoError(t, err)
	assert.Equal(t, 14, sizeOut)
	assert.Equal(t, []byte("fooEND\nbarEND\n"), buffer[0:sizeOut])
}
func TestLineStringProcesser(t *testing.T) {
	addEndProc := func(in string) (out string, err error) {
		return in + "END", nil
	}

	p := LineStringProcesser(addEndProc)

	message := "foo\nbar\nbaz"
	//buffer := make([]byte, len(message), 128)
	buffer := []byte(message)
	copy(buffer, []byte(message))
	sizeOut, err := p(buffer, 11)
	require.NoError(t, err)
	assert.Equal(t, 14, sizeOut)
	assert.Equal(t, []byte("fooEND\nbarEND\n"), buffer[0:sizeOut])
}
