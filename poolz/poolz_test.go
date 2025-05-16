package poolz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummyOpenerCloser struct {
	PoolCloser
	opened chan bool
	closed chan bool
}

func (o dummyOpenerCloser) Open() error {
	o.opened <- true
	return nil
}

func (o dummyOpenerCloser) Close() error {
	o.opened <- true
	return nil
}

func (o *dummyOpenerCloser) SetPoolCloser(pc PoolCloser) {
	o.PoolCloser = pc
}

func buildDummyOc() *dummyOpenerCloser {
	return &dummyOpenerCloser{
		opened: make(chan bool, 5),
		closed: make(chan bool, 5),
	}
}

func TestNewBasic(t *testing.T) {
	expectedMaxSize := 10
	factory := func() (*dummyOpenerCloser, error) {
		return buildDummyOc(), nil
	}
	var p Pool[*dummyOpenerCloser]
	_ = p
	p = NewBasic(expectedMaxSize, factory)
	require.NotNil(t, p)
	require.Implements(t, (*Pool[*dummyOpenerCloser])(nil), p)
}

func TestOpenClose(t *testing.T) {
	expectedMaxSize := 10

	var opened []*dummyOpenerCloser
	factory := func() (*dummyOpenerCloser, error) {
		oc := buildDummyOc()
		opened = append(opened, oc)
		return oc, nil
	}
	p := NewBasic(expectedMaxSize, factory)

	o, err := p.Open()
	assert.NoError(t, err)
	require.NotNil(t, o)
	require.NotNil(t, o.PoolCloser)
	require.NotNil(t, o.PoolCloser.wrapper)
	require.NotNil(t, o.PoolCloser.wrapper.onClose)

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
