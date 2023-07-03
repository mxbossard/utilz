package cmdz

import (
	//"fmt"
	//"io"
	//"context"
	//"log"
	//"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncRunAll(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := New(echoBinary, echoArg1).Outputs(&stdout1, &stderr1).Retries(2, 100)

	echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := New(echoBinary, echoArg2).Outputs(&stdout2, &stderr2).Retries(2, 100)

	p := AsyncRunAll(e1, e2)

	val, err := WaitAllResults(p)
	require.NoError(t, err)
	require.NotNil(t, val)
	assert.Equal(t, []int{0, 0}, *val)
	assert.Equal(t, []int{0}, e1.ResultsCodes)
	assert.Equal(t, []int{0}, e2.ResultsCodes)
}

func TestAsynkRunAll_WithFailure(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := New(echoBinary, echoArg1).Outputs(&stdout1, &stderr1).Retries(2, 100)

	//echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := New("/bin/false").Outputs(&stdout2, &stderr2).Retries(2, 100)

	p := AsyncRunAll(e1, e2)

	val, err := WaitAllResults(p)
	require.NoError(t, err)
	assert.Equal(t, []int{0, 1}, *val)
	assert.Equal(t, []int{0}, e1.ResultsCodes)
	assert.Equal(t, []int{1, 1, 1}, e2.ResultsCodes)
}

func TestBlockParallelRunAll(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := New(echoBinary, echoArg1).Outputs(&stdout1, &stderr1).Retries(2, 100)

	echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := New(echoBinary, echoArg2).Outputs(&stdout2, &stderr2).Retries(2, 100)

	val, err := BlockParallelRunAll(-1, e1, e2)
	require.NoError(t, err)
	require.NotNil(t, val)
	assert.Equal(t, []int{0, 0}, val)
	assert.Equal(t, []int{0}, e1.ResultsCodes)
	assert.Equal(t, []int{0}, e2.ResultsCodes)
}

func TestBlockParallelRunAll_WithFailure(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := New(echoBinary, echoArg1).Outputs(&stdout1, &stderr1).Retries(2, 100)

	//echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := New("/bin/false").Outputs(&stdout2, &stderr2).Retries(2, 100)

	val, err := BlockParallelRunAll(-1, e1, e2)
	require.NoError(t, err)
	require.NotNil(t, val)
	assert.Equal(t, []int{0, 1}, val)
	assert.Equal(t, []int{0}, e1.ResultsCodes)
	assert.Equal(t, []int{1, 1, 1}, e2.ResultsCodes)
}

func TestAsyncRunBest(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := New(echoBinary, echoArg1).Outputs(&stdout1, &stderr1).Retries(2, 100)

	echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := New(echoBinary, echoArg2).Outputs(&stdout2, &stderr2).Retries(2, 100)

	p := AsyncRunBest(e1, e2)

	br, err := WaitBestResults(p)
	require.NoError(t, err)
	require.NotNil(t, br)
	assert.False(t, br.DidError())
	assert.Equal(t, 2, br.Len())

	val, err := br.Result(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, val)
	assert.Equal(t, []int{0}, e1.ResultsCodes)

	val, err = br.Result(1)
	assert.NoError(t, err)
	assert.Equal(t, 0, val)
	assert.Equal(t, []int{0}, e2.ResultsCodes)
}

func TestAsyncRunBest_WithFailure(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := New(echoBinary, echoArg1).Outputs(&stdout1, &stderr1).Retries(2, 100)

	//echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := New("/bin/false").Outputs(&stdout2, &stderr2).Retries(2, 100)

	p := AsyncRunBest(e1, e2)

	br, err := WaitBestResults(p)
	require.NoError(t, err)
	require.NotNil(t, br)
	assert.False(t, br.DidError())
	assert.Equal(t, 2, br.Len())

	val, err := br.Result(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, val)
	assert.Equal(t, []int{0}, e1.ResultsCodes)

	val, err = br.Result(1)
	assert.NoError(t, err)
	assert.Equal(t, 1, val)
	assert.Equal(t, []int{1, 1, 1}, e2.ResultsCodes)

}
