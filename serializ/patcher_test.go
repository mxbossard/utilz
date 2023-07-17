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
