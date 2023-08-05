package structz

import (
	_ "io"
	_ "net/http"
	_ "net/http/httptest"
	_ "os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mby.fr/utils/serializ"
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

func TestYamlStringExplorer_Empty(t *testing.T) {
	/*
		exp := YamlStringExplorer(jsonEmpty0).Path("")
		require.NotNil(t, exp)
		_, err := exp.Resolve()
		require.Error(t, err)
	*/

	exp := YamlStringExplorer(jsonEmpty1)
	require.NotNil(t, exp)
	res, err := exp.Explore("")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, map[string]any{}, res)

	require.NotNil(t, exp)
	_, err = exp.Explore("foo")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrBadPathFormat)

	require.NotNil(t, exp)
	_, err = exp.Explore("/foo")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrPathDontExists)
}

func TestYamlStringExplorer_yaml0(t *testing.T) {
	yaml0, err := serializ.JsonToYamlString(json0)
	require.NoError(t, err)
	exp := YamlStringExplorer(yaml0)
	require.NotNil(t, exp)
	res, err := exp.Explore("")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, map[string]any{"key": "value"}, res)

	require.NotNil(t, exp)
	res, err = exp.Explore("/key")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "value", res)

	require.NotNil(t, exp)
	res, err = exp.Explore("/notExist")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrPathDontExists)
	assert.Nil(t, res)
}

func TestYamlStringExplorer_json0(t *testing.T) {
	exp := YamlStringExplorer(json0)
	require.NotNil(t, exp)
	res, err := exp.Explore("")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, map[string]any{"key": "value"}, res)

	require.NotNil(t, exp)
	res, err = exp.Explore("/key")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "value", res)

	require.NotNil(t, exp)
	res, err = exp.Explore("/notExist")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrPathDontExists)
	assert.Nil(t, res)
}

func TestYamlStringExplorer_json1(t *testing.T) {
	exp := YamlStringExplorer(json1)
	require.NotNil(t, exp)
	res, err := exp.Explore("/s1")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "foo", res)

	exp = YamlStringExplorer(json1)
	require.NotNil(t, exp)
	res, err = exp.Explore("/a1")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, []any{"bar", "baz"}, res)

	exp = YamlStringExplorer(json1)
	require.NotNil(t, exp)
	res, err = exp.Explore("/m1")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, map[string]any{"K1": "pif", "k2": "paf"}, res)

	exp = YamlStringExplorer(json1)
	require.NotNil(t, exp)
	res, err = exp.Explore("/m1/K1")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "pif", res)

	exp = YamlStringExplorer(json1)
	require.NotNil(t, exp)
	res, err = exp.Explore("/m1/k7")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPathDontExists)
	assert.Nil(t, res)
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

func TestYamlStringResolver_Empty(t *testing.T) {
	/*
		exp := YamlStringResolver[map[string]any](jsonEmpty0, "")
		require.NotNil(t, exp)
		_, err := exp.Resolve()
		require.Error(t, err)
	*/
	exp := YamlStringExplorer(jsonEmpty1)
	require.NotNil(t, exp)
	res, err := Resolve[map[string]any](exp, "")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, map[string]any{}, res)

	_, err = Resolve[map[string]any](exp, "/foo")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrPathDontExists)
}

func TestYamlStringResolver_Success(t *testing.T) {
	exp0 := YamlStringExplorer(json0)
	require.NotNil(t, exp0)
	res0, err := Resolve[type0](exp0, "")
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson0, res0)

	exp1 := YamlStringExplorer(json1)
	require.NotNil(t, exp1)
	res1, err := Resolve[type1](exp1, "")
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson1, res1)

	res2, err := Resolve[string](exp1, "/s1")
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson1.S1, res2)

	res3, err := ResolveArray[type1b](exp1, "/a1")
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson1.A1, res3)

	res4, err := Resolve[type1c](exp1, "/m1")
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson1.M1, res4)

	res5, err := Resolve[string](exp1, "/m1/K1")
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson1.M1.K1, res5)

	res6, err := Resolve[type1b](exp1, "/m1/k2")
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson1.M1.K2, res6)

	exp2 := YamlStringExplorer(json2)
	require.NotNil(t, exp2)
	res7, err := Resolve[type2](exp2, "")
	require.NoError(t, err)
	assert.Equal(t, expectedResolvedJson2, res7)

	res8, err := ResolveArray[type1](exp2, "/A2")
	require.NoError(t, err)
	assert.Equal(t, []type1{expectedResolvedJson1}, res8)

	res9, err := ResolveArray[type1c](exp2, "/A3")
	require.NoError(t, err)
	assert.Equal(t, []type1c{expectedResolvedJson1c, expectedResolvedJson1c2}, res9)

	exp3 := YamlStringExplorer(json3)
	require.NotNil(t, exp3)
	res10, err := Resolve[type3](exp3, "")
	require.NoError(t, err)
	assert.NotNil(t, res10)
	assert.Len(t, res10.A4, 1)

	res11, err := ResolveArray[type3](exp3, "/A4")
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
