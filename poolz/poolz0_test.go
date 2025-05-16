package poolz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummyOpenerCloser0 struct {
	opened chan bool
	closed chan bool
}

func (o dummyOpenerCloser0) Open() error {
	o.opened <- true
	return nil
}

func (o dummyOpenerCloser0) Close() error {
	o.opened <- true
	return nil
}

func (o dummyOpenerCloser0) PoolClose() error {
	return o.Close()
}

func buildDummyOc0() dummyOpenerCloser0 {
	return dummyOpenerCloser0{
		opened: make(chan bool, 5),
		closed: make(chan bool, 5),
	}
}

func TestNewBasic0(t *testing.T) {
	expectedMaxSize := 10
	factory := func() (dummyOpenerCloser0, error) {
		return buildDummyOc0(), nil
	}
	var p Pool0[dummyOpenerCloser0]
	_ = p
	p = NewBasic0(expectedMaxSize, factory)
	require.NotNil(t, p)
	require.Implements(t, (*Pool0[dummyOpenerCloser0])(nil), p)
}

func TestBasic0OpenClose(t *testing.T) {
	t.Skip()
	expectedMaxSize := 10

	var opened []*dummyOpenerCloser0
	factory := func() (dummyOpenerCloser0, error) {
		oc := buildDummyOc0()
		opened = append(opened, &oc)
		return oc, nil
	}
	p := NewBasic0(expectedMaxSize, factory)

	o, err := p.Open()
	assert.NoError(t, err)
	require.NotNil(t, o)
	// dummy should have been open
	require.Len(t, opened, 1)
	assert.Len(t, opened[0].opened, 1)
	assert.Len(t, opened[0].closed, 0)
	assert.Len(t, p.available, 0)
	assert.Equal(t, 1, p.inUse)
	// assert.NotNil(t, o.Get())

	err = o.PoolClose()
	assert.NoError(t, err)
	// dummy should not have been closed bu stay in pool
	assert.Len(t, opened[0].opened, 1)
	assert.Len(t, opened[0].closed, 0)
	assert.Len(t, p.available, 1)
	assert.Equal(t, 0, p.inUse)

	o, err = p.Open()
	assert.NoError(t, err)
	require.NotNil(t, o)
	// dummy should have been open once
	require.Len(t, opened, 1)
	assert.Len(t, opened[0].opened, 1)
	assert.Len(t, opened[0].closed, 0)
	assert.Len(t, p.available, 0)
	assert.Equal(t, 1, p.inUse)
	// assert.NotNil(t, o.Get())

	o2, err := p.Open()
	assert.NoError(t, err)
	require.NotNil(t, o2)
	// dummy should have been open once
	require.Len(t, opened, 2)
	assert.Len(t, opened[0].opened, 1)
	assert.Len(t, opened[0].closed, 0)
	assert.Len(t, opened[1].opened, 1)
	assert.Len(t, opened[1].closed, 0)
	assert.Len(t, p.available, 0)
	assert.Equal(t, 2, p.inUse)
	// assert.NotNil(t, o2.Get())

	err = o.PoolClose()
	assert.NoError(t, err)
	// dummy should not have been closed bu stay in pool
	assert.Len(t, opened[0].opened, 1)
	assert.Len(t, opened[0].closed, 0)
	assert.Len(t, opened[1].opened, 1)
	assert.Len(t, opened[1].closed, 0)
	assert.Len(t, p.available, 1)
	assert.Equal(t, 1, p.inUse)

	err = o2.PoolClose()
	assert.NoError(t, err)
	// dummy should not have been closed bu stay in pool
	assert.Len(t, opened[0].opened, 1)
	assert.Len(t, opened[0].closed, 0)
	assert.Len(t, opened[1].opened, 1)
	assert.Len(t, opened[1].closed, 0)
	assert.Len(t, p.available, 2)
	assert.Equal(t, 0, p.inUse)
}
