package inout

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrowOrCopy(t *testing.T) {
	l1c5 := make([]byte, 1, 5)
	l1c5[0] = 10
	GrowOrCopy(&l1c5, 4, 5)
	assert.Len(t, l1c5, 4)
	assert.Equal(t, 5, cap(l1c5))
	assert.Equal(t, []byte{10, 0, 0, 0}, l1c5)

	GrowOrCopy(&l1c5, 3, 5)
	assert.Len(t, l1c5, 3)
	assert.Equal(t, 5, cap(l1c5))
	assert.Equal(t, []byte{10, 0, 0}, l1c5)

	GrowOrCopy(&l1c5, 7, 10)
	assert.Len(t, l1c5, 7)
	assert.Equal(t, 10, cap(l1c5))
	assert.Equal(t, []byte{10, 0, 0, 0, 0, 0, 0}, l1c5)

	GrowOrCopy(&l1c5, 3, 5)
	assert.Len(t, l1c5, 3)
	assert.Equal(t, 10, cap(l1c5))
	assert.Equal(t, []byte{10, 0, 0}, l1c5)
}

func TestLineProcesser(t *testing.T) {
	addEndProc := func(buffer *[]byte, sizeIn int) (sizeOut int, err error) {
		*buffer = (*buffer)[:sizeIn+4]
		copy((*buffer)[sizeIn:], []byte("END"))
		return sizeIn + 3, nil
	}

	p := LineProcesser(addEndProc)

	message := "foo\nbar\nbaz"
	// Make a buffer with a too small capacity
	buffer := make([]byte, len(message), 11)
	copy(buffer, []byte(message))
	sizeOut, err := p.Process(&buffer, 11)
	require.NoError(t, err)
	assert.Equal(t, 14, sizeOut)
	assert.Equal(t, []byte("fooEND\nbarEND\n"), buffer)

	buffer, err = p.AvailableBuffer()
	require.NoError(t, err)
	assert.Len(t, buffer, 7)
	assert.Equal(t, []byte("bazEND\n"), buffer)
}

func TestStringLineProcesser(t *testing.T) {
	addEndProc := func(in string) (out string, err error) {
		return in + "END", nil
	}

	p := StringLineProcesser(addEndProc)

	message := "foo\nbar\nbaz"
	// Make a buffer with a too small capacity
	buffer := make([]byte, len(message), 11)
	copy(buffer, []byte(message))
	//b := &buffer
	sizeOut, err := p.Process(&buffer, 11)
	require.NoError(t, err)
	assert.Equal(t, 14, sizeOut)
	assert.Equal(t, []byte("fooEND\nbarEND\n"), buffer)
}
