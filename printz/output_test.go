package printz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStringOutputs(t *testing.T) {
	outW, errW, outs := NewStringOutputs()

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	expectedOut := "foo"
	expectedErr := "bar"

	outs.Out().Write([]byte(expectedOut))

	assert.Equal(t, expectedOut, outW.String())
	assert.Empty(t, errW.String())

	outs.Flush()

	assert.Equal(t, expectedOut, outW.String())
	assert.Empty(t, errW.String())

	outs.Err().Write([]byte(expectedErr))

	assert.Equal(t, expectedOut, outW.String())
	assert.Equal(t, expectedErr, errW.String())

	outs.Flush()

	assert.Equal(t, expectedOut, outW.String())
	assert.Equal(t, expectedErr, errW.String())
}

func TestNewBufferedOutputs_1(t *testing.T) {

	outW, errW, outs := NewStringOutputs()

	bouts := NewBufferedOutputs(outs)

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	expectedOut := "foo"
	expectedErr := "bar"

	bouts.Out().Write([]byte(expectedOut))

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	outs.Flush()

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	bouts.Flush()

	assert.Equal(t, expectedOut, outW.String())
	assert.Empty(t, errW.String())

	bouts.Err().Write([]byte(expectedErr))

	assert.Equal(t, expectedOut, outW.String())
	assert.Empty(t, errW.String())

	outs.Flush()

	assert.Equal(t, expectedOut, outW.String())
	assert.Empty(t, errW.String())

	bouts.Flush()

	assert.Equal(t, expectedOut, outW.String())
	assert.Equal(t, expectedErr, errW.String())
}

func TestNewBufferedOutputs_2(t *testing.T) {

	outW, errW, outs := NewStringOutputs()

	bouts1 := NewBufferedOutputs(outs)
	bouts2 := NewBufferedOutputs(bouts1)

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	expectedOut := "foo"
	expectedErr := "bar"

	bouts2.Out().Write([]byte(expectedOut))

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	bouts1.Flush()

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	bouts2.Flush()

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	bouts1.Flush()

	assert.Equal(t, expectedOut, outW.String())
	assert.Empty(t, errW.String())

	bouts2.Err().Write([]byte(expectedErr))

	assert.Equal(t, expectedOut, outW.String())
	assert.Empty(t, errW.String())

	bouts1.Flush()

	assert.Equal(t, expectedOut, outW.String())
	assert.Empty(t, errW.String())

	bouts2.Flush()

	assert.Equal(t, expectedOut, outW.String())
	assert.Empty(t, errW.String())

	bouts1.Flush()

	assert.Equal(t, expectedOut, outW.String())
	assert.Equal(t, expectedErr, errW.String())
}
