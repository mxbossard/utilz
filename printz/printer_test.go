package printz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	outW, errW, outs := NewStringOutputs()

	p := New(outs)

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	p.Out("foo")

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	p.Flush()

	assert.Equal(t, "foo", outW.String())
	assert.Empty(t, errW.String())

	p.Err("bar")

	assert.Equal(t, "foo", outW.String())
	assert.Empty(t, errW.String())

	p.Flush()

	assert.Equal(t, "foo", outW.String())
	assert.Equal(t, "bar", errW.String())
}

func TestNewUnbuffured(t *testing.T) {
	outW, errW, outs := NewStringOutputs()

	p := NewUnbuffured(outs)

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	p.Out("foo")

	assert.Equal(t, "foo", outW.String())
	assert.Empty(t, errW.String())

	p.Flush()

	assert.Equal(t, "foo", outW.String())
	assert.Empty(t, errW.String())

	p.Err("bar")

	assert.Equal(t, "foo", outW.String())
	assert.Equal(t, "bar", errW.String())

	p.Flush()

	assert.Equal(t, "foo", outW.String())
	assert.Equal(t, "bar", errW.String())
}

func TestBuffured_1(t *testing.T) {
	outW, errW, outs := NewStringOutputs()

	p := Buffered(NewUnbuffured(outs))

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	p.Out("foo")

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	p.Flush()

	assert.Equal(t, "foo", outW.String())
	assert.Empty(t, errW.String())

	p.Err("bar")

	assert.Equal(t, "foo", outW.String())
	assert.Empty(t, errW.String())

	p.Flush()

	assert.Equal(t, "foo", outW.String())
	assert.Equal(t, "bar", errW.String())
}

func TestBuffured_2(t *testing.T) {
	outW, errW, outs := NewStringOutputs()

	p1 := New(outs)
	p2 := Buffered(p1)

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	p2.Out("foo")

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	p1.Flush()

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	p2.Flush()

	assert.Empty(t, outW.String())
	assert.Empty(t, errW.String())

	p1.Flush()

	assert.Equal(t, "foo", outW.String())
	assert.Empty(t, errW.String())

	p2.Err("bar")

	assert.Equal(t, "foo", outW.String())
	assert.Empty(t, errW.String())

	p1.Flush()

	assert.Equal(t, "foo", outW.String())
	assert.Empty(t, errW.String())

	p2.Flush()

	assert.Equal(t, "foo", outW.String())
	assert.Empty(t, errW.String())

	p1.Flush()

	assert.Equal(t, "foo", outW.String())
	assert.Equal(t, "bar", errW.String())
}
