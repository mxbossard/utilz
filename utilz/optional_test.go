package utilz

import (
	"bytes"
	"encoding/gob"
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

func TestGobEncode_Then_Decode(t *testing.T) {
	var err error
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	buf.Reset()
	expectedInt := OptionalOf(42)
	//expectedInt.Default(23)
	var gotInt Optional[int]
	err = enc.Encode(expectedInt)
	require.NoError(t, err)

	err = dec.Decode(&gotInt)
	require.NoError(t, err)
	assert.NotNil(t, gotInt.Value)
	assert.Nil(t, gotInt.Def)
	assert.Equal(t, 42, *gotInt.Value)
	require.True(t, gotInt.IsPresent())
	assert.Equal(t, expectedInt, gotInt)

	buf.Reset()
	expectedString := OptionalOf("foo")
	var gotString Optional[string]
	err = enc.Encode(expectedString)
	require.NoError(t, err)

	err = dec.Decode(&gotString)
	require.NoError(t, err)
	require.True(t, gotString.IsPresent())
	assert.Equal(t, expectedString, gotString)

	buf.Reset()
	expectedBool := OptionalOf(false)
	expectedBool.Default(true)
	var gotBool Optional[bool]
	err = enc.Encode(expectedBool)
	require.NoError(t, err)

	err = dec.Decode(&gotBool)
	require.NoError(t, err)
	assert.NotNil(t, gotBool.Value)
	assert.NotNil(t, gotBool.Def)
	assert.Equal(t, false, *gotBool.Value)
	assert.Equal(t, true, *gotBool.Def)
	require.True(t, gotBool.IsPresent())
	assert.Equal(t, expectedBool, gotBool)

}
