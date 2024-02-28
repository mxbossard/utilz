package utilz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionalMerge(t *testing.T) {
	var left, right Optional[int]

	left = EmptyOptional[int]()
	right = EmptyOptional[int]()
	left.Merge(right)
	require.False(t, left.IsPresent())
	require.False(t, right.IsPresent())

	left = OptionalOf(42)
	right = EmptyOptional[int]()
	left.Merge(right)
	require.True(t, left.IsPresent())
	require.False(t, right.IsPresent())
	assert.Equal(t, 42, left.Get())

	left = EmptyOptional[int]()
	right = OptionalOf(43)
	left.Merge(right)
	require.True(t, left.IsPresent())
	require.True(t, right.IsPresent())
	assert.Equal(t, 43, left.Get())
	assert.Equal(t, 43, right.Get())

	left = OptionalOf(42)
	right = OptionalOf(43)
	left.Merge(right)
	require.True(t, left.IsPresent())
	require.True(t, right.IsPresent())
	assert.Equal(t, 43, left.Get())
	assert.Equal(t, 43, right.Get())

}
