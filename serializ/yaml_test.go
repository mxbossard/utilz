package serializ

import (
	_ "strings"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	expectedJson1 = `{"k1":"v1","k2":"v2"}`
	expectedJson2 = `{"a1":[1,2,3],"k1":"v1","m1":{"a":"foo","b":"bar"}}`

	expectedYaml1 = `k1: v1
k2: v2
`
	expectedYaml2 = `a1:
    - 1
    - 2
    - 3
k1: v1
m1:
    a: foo
    b: bar
`
)

func TestYamlToJsonString(t *testing.T) {
	// Can convert json to json because json is valid yaml
	res, err := YamlToJsonString(expectedJson1)
	require.NoError(t, err)
	assert.Equal(t, expectedJson1, res)

	res, err = YamlToJsonString(expectedYaml1)
	require.NoError(t, err)
	assert.Equal(t, expectedJson1, res)
	
	res, err = YamlToJsonString(expectedYaml2)
	require.NoError(t, err)
	assert.Equal(t, expectedJson2, res)
	
}

func TestJsonToYamlString(t *testing.T) {
	// Should error bad format
	res, err := JsonToYamlString(expectedYaml1)
	require.Error(t, err)

	res, err = JsonToYamlString(expectedJson1)
	require.NoError(t, err)
	assert.Equal(t, expectedYaml1, res)

	res, err = JsonToYamlString(expectedJson2)
	require.NoError(t, err)
	assert.Equal(t, expectedYaml2, res)

}
