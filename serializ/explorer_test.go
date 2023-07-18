package serializ

import (
	_ "io"
	_ "net/http"
	_ "net/http/httptest"
	_ "os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	jsonEmpty0 = ``
	jsonEmpty1 = `{}`
	json0      = `{"key": "value"}`
	json1c     = `{"K1": "pif", "k2": "paf"}`
	json1c2    = `{"k1": "puf", "K2": "pef"}`
	json1      = `{"s1": "foo", "a1": ["bar", "baz"], "m1": ` + json1c + `}`
	json2      = `{"A2": [` + json1 + `], "A3": [` + json1c + `, ` + json1c2 + `]}`
	json3      = `{"A4": [{"A4": [{"A4": [{"A4": []}]}]}]}`

	expectedResolvedJson0  = type0{Key: "value"}
	expectedResolvedJson1c = type1c{
		K1: "pif",
		K2: type1b("paf"),
	}
	expectedResolvedJson1c2 = type1c{
		K1: "puf",
		K2: type1b("pef"),
	}
	expectedResolvedJson1 = type1{
		S1: "foo",
		A1: []type1b{type1b("bar"), type1b("baz")},
		M1: expectedResolvedJson1c,
	}
	expectedResolvedJson2 = type2{
		A2: []type1{expectedResolvedJson1},
		A3: []type1c{
			expectedResolvedJson1c,
			expectedResolvedJson1c2,
		},
	}
)

type (
	type0 struct {
		Key string
	}

	type1 struct {
		S1 string
		A1 []type1b
		M1 type1c
	}

	type1b = string

	type1c struct {
		K1 string
		K2 type1b
	}

	type2 struct {
		A2 []type1
		A3 []type1c
	}

	type3 struct {
		A4 []type3
	}
)

func TestJsonStringExplorer_Empty(t *testing.T) {
	exp := JsonStringExplorer(jsonEmpty0).Path("")
	require.NotNil(t, exp)
	_, err := exp.Resolve()
	require.Error(t, err)

	exp = JsonStringExplorer(jsonEmpty1).Path("")
	require.NotNil(t, exp)
	res, err := exp.Resolve()
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, map[string]any{}, res)

	exp = JsonStringExplorer(jsonEmpty1).Path("foo")
	require.NotNil(t, exp)
	_, err = exp.Resolve()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrBadPathFormat)

	exp = JsonStringExplorer(jsonEmpty1).Path("/foo")
	require.NotNil(t, exp)
	_, err = exp.Resolve()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrPathDontExists)
}

func TestJsonStringExplorer_json0(t *testing.T) {
	exp := JsonStringExplorer(json0).Path("")
	require.NotNil(t, exp)
	res, err := exp.Resolve()
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, map[string]any{"key": "value"}, res)

	exp = JsonStringExplorer(json0).Path("/key")
	require.NotNil(t, exp)
	res, err = exp.Resolve()
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "value", res)

	exp = JsonStringExplorer(json0).Path("/notExist")
	require.NotNil(t, exp)
	res, err = exp.Resolve()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrPathDontExists)
}

func TestJsonStringExplorer_json1(t *testing.T) {
	exp := JsonStringExplorer(json1).Path("/s1")
	require.NotNil(t, exp)
	res, err := exp.Resolve()
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "foo", res)

	exp = JsonStringExplorer(json1).Path("/a1")
	require.NotNil(t, exp)
	res, err = exp.Resolve()
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, []any{"bar", "baz"}, res)

	exp = JsonStringExplorer(json1).Path("/m1")
	require.NotNil(t, exp)
	res, err = exp.Resolve()
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, map[string]any{"K1": "pif", "k2": "paf"}, res)

	exp = JsonStringExplorer(json1).Path("/m1/K1")
	require.NotNil(t, exp)
	res, err = exp.Resolve()
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "pif", res)

	exp = JsonStringExplorer(json1).Path("/m1/k7")
	require.NotNil(t, exp)
	res, err = exp.Resolve()
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPathDontExists)
}

func TestMap2Struct(t *testing.T) {
	in0 := map[string]any{}
	res0, err := map2Struct[type0](in0)
	require.NoError(t, err)
	require.NotNil(t, res0)
	assert.Equal(t, type0{}, res0)

	in1 := map[string]any{"key": "value"}
	res1, err := map2Struct[type0](in1)
	require.NoError(t, err)
	require.NotNil(t, res1)
	assert.Equal(t, expectedResolvedJson0, res1)

	in2 := map[string]any{"s1": "foo", "a1": []string{"bar", "baz"}, "m1": map[string]any{"K1": "pif", "k2": "paf"}}
	res2, err := map2Struct[type1](in2)
	require.NoError(t, err)
	require.NotNil(t, res2)
	assert.Equal(t, expectedResolvedJson1, res2)
}

func TestJsonStringResolver_Empty(t *testing.T) {
	exp := JsonStringResolver[map[string]any](jsonEmpty0, "")
	require.NotNil(t, exp)
	_, err := exp.Resolve()
	require.Error(t, err)

	exp = JsonStringResolver[map[string]any](jsonEmpty1, "")
	require.NotNil(t, exp)
	res, err := exp.Resolve()
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, map[string]any{}, res)

	exp = JsonStringResolver[map[string]any](jsonEmpty1, "/foo")
	require.NotNil(t, exp)
	_, err = exp.Resolve()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrPathDontExists)
}

func TestJsonStringResolver_Success(t *testing.T) {
	exp0 := JsonStringResolver[type0](json0, "")
	require.NotNil(t, exp0)
	res0, err := exp0.Resolve()
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson0, res0)

	exp1 := JsonStringResolver[type1](json1, "")
	require.NotNil(t, exp1)
	res1, err := exp1.Resolve()
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson1, res1)

	exp2 := JsonStringResolver[string](json1, "/s1")
	require.NotNil(t, exp2)
	res2, err := exp2.Resolve()
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson1.S1, res2)

	exp3 := JsonStringResolver[type1b](json1, "/a1")
	require.NotNil(t, exp3)
	res3, err := exp3.ResolveArray()
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson1.A1, res3)

	exp4 := JsonStringResolver[type1c](json1, "/m1")
	require.NotNil(t, exp4)
	res4, err := exp4.Resolve()
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson1.M1, res4)

	exp5 := JsonStringResolver[string](json1, "/m1/K1")
	require.NotNil(t, exp5)
	res5, err := exp5.Resolve()
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson1.M1.K1, res5)

	exp6 := JsonStringResolver[type1b](json1, "/m1/k2")
	require.NotNil(t, exp6)
	res6, err := exp6.Resolve()
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson1.M1.K2, res6)

	exp7 := JsonStringResolver[type2](json2, "")
	require.NotNil(t, exp7)
	res7, err := exp7.Resolve()
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson2, res7)

	exp8 := JsonStringResolver[type1](json2, "/A2")
	require.NotNil(t, exp8)
	res8, err := exp8.ResolveArray()
	require.NoError(t, err)
	assert.Equal(t, []type1{expectedResolvedJson1}, res8)

	exp9 := JsonStringResolver[type1c](json2, "/A3")
	require.NotNil(t, exp9)
	res9, err := exp9.ResolveArray()
	require.NoError(t, err)
	assert.Equal(t, []type1c{expectedResolvedJson1c, expectedResolvedJson1c2}, res9)

	exp10 := JsonStringResolver[type3](json3, "")
	require.NotNil(t, exp10)
	res10, err := exp10.Resolve()
	require.NoError(t, err)
	assert.NotNil(t, res10)
	assert.Len(t, res10.A4, 1)

	exp11 := JsonStringResolver[type3](json3, "/A4")
	require.NotNil(t, exp11)
	res11, err := exp11.ResolveArray()
	require.NoError(t, err)
	assert.NotNil(t, res11)
	require.Len(t, res11, 1)
	assert.NotNil(t, res11[0].A4)
	require.Len(t, res11[0].A4, 1)
	assert.NotNil(t, res11[0].A4[0])
	require.Len(t, res11[0].A4[0].A4, 1)
	assert.NotNil(t, res11[0].A4[0].A4[0])
	require.Len(t, res11[0].A4[0].A4[0].A4, 0)
}
