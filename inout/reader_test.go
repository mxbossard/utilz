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

	pr := NewProcessingStreamReader(nested)

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
	errProcesser := BasicProcesser(func(b *[]byte, i int) (o int, err error) {
		return 0, fmt.Errorf("errProcesserError")
	})
	barProcesser := BasicProcesser(func(b *[]byte, i int) (o int, err error) {
		copy(*b, expectedBytes2)
		return len(expectedBytes2), nil
	})
	endProcesser := BasicProcesser(func(b *[]byte, i int) (o int, err error) {
		*b = (*b)[:i+3]
		copy((*b)[i:], expectedBytes3)
		return i + 3, nil
	})
	buffer := make([]byte, len(expectedMsg1)+5)

	nested := strings.NewReader(expectedMsg1)
	pr := NewProcessingStreamReader(nested)
	pr.Add(errProcesser)
	n, err := pr.Read(buffer)
	require.Error(t, err)
	assert.Equal(t, 0, n)

	nested = strings.NewReader(expectedMsg1)
	pr = NewProcessingStreamReader(nested)
	pr.Add(barProcesser)
	n, err = pr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, len(expectedBytes2), n)
	assert.Equal(t, expectedBytes2, buffer[0:n])

	nested = strings.NewReader(expectedMsg1)
	pr = NewProcessingStreamReader(nested)
	pr.Add(endProcesser)
	n, err = pr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, len(expectedBytes1)+len(expectedBytes3), n)
	assert.Equal(t, append(expectedBytes1, expectedBytes3...), buffer[0:n])
}

func TestProcessingStreamReader_LongerThanBuffer(t *testing.T) {
	expectedMsg1 := "foofoofoobarbarbar"
	expectedMsg3 := "END"
	expectedBytes3 := []byte(expectedMsg3)

	endProcesser := BasicProcesser(func(b *[]byte, i int) (o int, err error) {
		// set buffer length
		*b = (*b)[:i+3]
		copy((*b)[i:], expectedBytes3)
		return i + 3, nil
	})
	buffer := make([]byte, 9, 32)

	nested := strings.NewReader(expectedMsg1)
	pr := NewProcessingStreamReader(nested)
	pr.Add(endProcesser)
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

	pr := NewProcessingBufferReader(nested)

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
	errProcesser := BasicProcesser(func(b *[]byte, i int) (o int, err error) {
		return 0, fmt.Errorf("errProcesserError")
	})
	barProcesser := BasicProcesser(func(b *[]byte, i int) (o int, err error) {
		copy(*b, expectedBytes2)
		return len(expectedBytes2), nil
	})
	endProcesser := BasicProcesser(func(b *[]byte, i int) (o int, err error) {
		*b = (*b)[:i+3]
		copy((*b)[i:], expectedBytes3)
		return i + 3, nil
	})
	buffer := make([]byte, len(expectedMsg1)+16)

	nested := strings.NewReader(expectedMsg1)
	pr := NewProcessingBufferReader(nested)
	pr.Add(errProcesser)
	n, err := pr.Read(buffer)
	require.Error(t, err)
	assert.Equal(t, 0, n)

	nested = strings.NewReader(expectedMsg1)
	pr = NewProcessingBufferReader(nested)
	pr.Add(barProcesser)
	n, err = pr.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, len(expectedBytes2), n)
	assert.Equal(t, expectedBytes2, buffer[0:n])

	nested = strings.NewReader(expectedMsg1)
	pr = NewProcessingBufferReader(nested)
	pr.Add(endProcesser)
	n, err = pr.Read(buffer[0:16])
	require.NoError(t, err)
	assert.Equal(t, len(expectedBytes1)+len(expectedBytes3), n)
	assert.Equal(t, append(expectedBytes1, expectedBytes3...), buffer[0:n])
}

func TestProcessingBufferReader_LongerThanBuffer(t *testing.T) {
	expectedMsg1 := "foofoofoobarbarbar"
	expectedMsg3 := "END"
	expectedBytes3 := []byte(expectedMsg3)

	endProcesser := BasicProcesser(func(b *[]byte, i int) (o int, err error) {
		// set buffer length
		*b = (*b)[:i+3]
		copy((*b)[i:], expectedBytes3)
		return i + 3, nil
	})
	buffer := make([]byte, 9)

	nested := strings.NewReader(expectedMsg1)
	pr := NewProcessingBufferReader(nested)
	pr.Add(endProcesser)
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
