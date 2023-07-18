package serializ

import (
	_ "io"
	_ "net/http"
	_ "net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	emptyMapTree   = map[string]any{}
	simpleMapTree  = map[string]any{"string": "foo", "int": 4}
	complexMapTree = map[string]any{
		"string": "foo",
		"int":    4,
		"map": map[string]any{
			"k1": "v1",
			"k2": "v2",
		},
		"array": []string{
			"foo",
			"bar",
			"baz",
		},
	}
)

func TestParentPath(t *testing.T) {
	res := parentPath("")
	assert.Equal(t, "/", res)

	res = parentPath("/")
	assert.Equal(t, "/", res)

	res = parentPath("/foo")
	assert.Equal(t, "/", res)

	res = parentPath("/foo/")
	assert.Equal(t, "/", res)

	res = parentPath("/foo/bar")
	assert.Equal(t, "/foo", res)

	res = parentPath("/foo/bar/")
	assert.Equal(t, "/foo", res)
}

func TestLastChild(t *testing.T) {
	res := lastChild("")
	assert.Equal(t, "", res)

	res = lastChild("/")
	assert.Equal(t, "", res)

	res = lastChild("/foo")
	assert.Equal(t, "foo", res)

	res = lastChild("/foo/")
	assert.Equal(t, "foo", res)

	res = lastChild("/foo/bar")
	assert.Equal(t, "bar", res)

	res = lastChild("/foo/bar/")
	assert.Equal(t, "bar", res)
}

func TestTreeLeaf_BadPath(t *testing.T) {
	_, err := treeLeaf[string](simpleMapTree, "")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBatPathFormat)

	_, err = treeLeaf[string](simpleMapTree, "foo")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBatPathFormat)
}

func TestTreeLeaf_Success(t *testing.T) {
	r0, err := treeLeaf[map[string]any](simpleMapTree, "/")
	require.NoError(t, err)
	assert.NotNil(t, r0)
	assert.Equal(t, "foo", r0["string"])

	r1, err := treeLeaf[string](simpleMapTree, "/string")
	require.NoError(t, err)
	assert.Equal(t, "foo", r1)

	r2, err := treeLeaf[int](simpleMapTree, "/int")
	require.NoError(t, err)
	assert.Equal(t, 4, r2)

	r3, err := treeLeaf[string](complexMapTree, "/map/k1")
	require.NoError(t, err)
	assert.Equal(t, "v1", r3)

	r4, err := treeLeaf[[]string](complexMapTree, "/array")
	require.NoError(t, err)
	assert.NotNil(t, r4)
	assert.Len(t, r4, 3)
	assert.Contains(t, r4, "foo")
	assert.Contains(t, r4, "bar")
	assert.Contains(t, r4, "baz")
}

func TestTreeAdd(t *testing.T) {
	in := emptyMapTree

	// Test empty args
	_, err := treeAdd(nil, "", nil)
	require.Error(t, err)

	_, err = treeAdd(in, "", nil)
	require.Error(t, err)

	_, err = treeAdd(nil, "/foo", nil)
	require.Error(t, err)

	_, err = treeAdd(in, "/foo", nil)
	require.Error(t, err)

	// Test nil input map
	res, err := treeAdd(nil, "/foo", "bar")
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"foo": "bar"}, res)

	res, err = treeAdd(nil, "/", simpleMapTree)
	require.NoError(t, err)
	assert.Equal(t, simpleMapTree, res)
	assert.NotSame(t, simpleMapTree, res)

	// Test bad value type for map replacement
	_, err = treeAdd(nil, "", "bar")
	require.Error(t, err)

	_, err = treeAdd(in, "", "foo")
	require.Error(t, err)

	_, err = treeAdd(in, "/", "foo")
	require.Error(t, err)

	// Test adding empty string
	require.NotContains(t, in, "key")
	res, err = treeAdd(in, "/key", "")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Contains(t, res, "key")
	assert.Equal(t, "", res["key"])
	require.NotContains(t, in, "key")

	// Test adding string
	res, err = treeAdd(in, "/key", "foo")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Contains(t, res, "key")
	assert.Equal(t, "foo", res["key"])
	require.NotContains(t, in, "key")

	// Test adding int
	res, err = treeAdd(in, "/key", 42)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Contains(t, res, "key")
	assert.Equal(t, 42, res["key"])
	require.NotContains(t, in, "key")

	// Test adding map
	res, err = treeAdd(in, "/key", simpleMapTree)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Contains(t, res, "key")
	assert.Equal(t, simpleMapTree, res["key"])
	require.NotContains(t, in, "key")

	// Test overriding document
	res, err = treeAdd(in, "/", simpleMapTree)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, simpleMapTree, res)
	require.Equal(t, emptyMapTree, in)
	require.NotEqual(t, emptyMapTree, simpleMapTree)

	// Test adding to array
	in = complexMapTree
	require.NotContains(t, in["array"], "pif")
	res, err = treeAdd(in, "/array", "pif")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res["array"])
	assert.Len(t, res["array"], 4)
	assert.Contains(t, res["array"], "pif")
	require.Len(t, in["array"], 3)

	// Test adding slice into array
	in = complexMapTree
	res, err = treeAdd(in, "/array", []string{"pif"})
	require.Error(t, err)

	// Test adding wrong type into array
	in = complexMapTree
	res, err = treeAdd(in, "/array", 4)
	require.Error(t, err)
}

func TestTreeRemove(t *testing.T) {
	// Test empty args
	res, err := treeRemove(nil, "")
	require.NoError(t, err)
	assert.Equal(t, emptyMapTree, res)

	res, err = treeRemove(simpleMapTree, "")
	require.NoError(t, err)
	assert.Equal(t, emptyMapTree, res)

	res, err = treeRemove(nil, "/")
	require.NoError(t, err)
	assert.Equal(t, emptyMapTree, res)

	// Test removing
	res, err = treeRemove(simpleMapTree, "/notExists")
	require.NoError(t, err)
	assert.Equal(t, simpleMapTree, res)

	res, err = treeRemove(simpleMapTree, "/string")
	require.NoError(t, err)
	assert.NotEqual(t, simpleMapTree, res)
	assert.Len(t, res, 1)
	assert.NotContains(t, res, "string")
	assert.Contains(t, res, "int")
	assert.Len(t, simpleMapTree, 2)
	assert.Contains(t, simpleMapTree, "string")
	assert.Contains(t, simpleMapTree, "int")

	// Test adding empty string
	res, err = treeRemove(simpleMapTree, "/")
	require.NoError(t, err)
	assert.Equal(t, emptyMapTree, res)
	assert.NotEqual(t, simpleMapTree, res)
	assert.Len(t, res, 0)
	assert.Len(t, simpleMapTree, 2)
	assert.Contains(t, simpleMapTree, "string")
	assert.Contains(t, simpleMapTree, "int")

}

func TestTransform_JsonAdd(t *testing.T) {
	op := OpAdd("/foo", 4)

	in := emptyMapTree
	out, err := op.Transform(in)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Contains(t, out, "foo")
	assert.Equal(t, out["foo"], 4)

	in = simpleMapTree
	out, err = op.Transform(in)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Contains(t, out, "string")
	assert.Contains(t, out, "int")
	assert.Contains(t, out, "foo")
	assert.Equal(t, out["string"], "foo")
	assert.Equal(t, out["int"], 4)
	assert.Equal(t, out["foo"], 4)

	op = OpAdd("/", simpleMapTree)
	in = emptyMapTree
	out, err = op.Transform(in)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, simpleMapTree, out)
}
