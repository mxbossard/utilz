package error

import (
	"testing"
	"fmt"

        "github.com/stretchr/testify/assert"
        //"github.com/stretchr/testify/require"

)

func TestAggregated(t *testing.T) {
	err1 := fmt.Errorf("err1")
	err2 := fmt.Errorf("err2")
	err3 := fmt.Errorf("err3")
	err4 := fmt.Errorf("err4")
	err5 := fmt.Errorf("err1")

	agg := Aggregated{}
	agg.Add(err1)
	agg.Add(err2)
	agg.Add(err3)

	assert.Equal(t, "err1\nerr2\nerr3\n", agg.Error())
	assert.True(t, agg.Is(err1))
	assert.True(t, agg.Is(err2))
	assert.True(t, agg.Is(err3))
	assert.False(t, agg.Is(err4))
	assert.False(t, agg.Is(err5))

	assert.Equal(t, []error{err1}, agg.Get(err1))
	assert.Equal(t, []error{err2}, agg.Get(err2))
	assert.Equal(t, []error{err3}, agg.Get(err3))
	assert.Nil(t, agg.Get(err4))
	assert.Nil(t, agg.Get(err5))
}
