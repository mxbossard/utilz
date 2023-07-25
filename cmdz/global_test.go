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
	e1 := Cmd(echoBinary, echoArg1).Outputs(&stdout1, &stderr1).Retries(2, 100)

	echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := Cmd(echoBinary, echoArg2).Outputs(&stdout2, &stderr2).Retries(2, 100)

	p := AsyncRunAll(e1, e2)

	val, err := WaitAllResults(p)
	require.NoError(t, err)
	require.NotNil(t, val)
	assert.Equal(t, []int{0, 0}, *val)
	assert.Equal(t, []int{0}, e1.ResultCodes())
	assert.Equal(t, []int{0}, e2.ResultCodes())
}

func TestAsynkRunAll_WithFailure(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := Cmd(echoBinary, echoArg1).Outputs(&stdout1, &stderr1).Retries(2, 100)

	//echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := Cmd("/bin/false").Outputs(&stdout2, &stderr2).Retries(2, 100)

	p := AsyncRunAll(e1, e2)

	val, err := WaitAllResults(p)
	require.NoError(t, err)
	assert.Equal(t, []int{0, 1}, *val)
	assert.Equal(t, []int{0}, e1.ResultCodes())
	assert.Equal(t, []int{1, 1, 1}, e2.ResultCodes())
}

func TestBlockParallelRunAll(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := Cmd(echoBinary, echoArg1).Outputs(&stdout1, &stderr1).Retries(2, 100)

	echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := Cmd(echoBinary, echoArg2).Outputs(&stdout2, &stderr2).Retries(2, 100)

	val, err := blockParallelRunAll(-1, e1, e2)
	require.NoError(t, err)
	require.NotNil(t, val)
	assert.Equal(t, []int{0, 0}, val)
	assert.Equal(t, []int{0}, e1.ResultCodes())
	assert.Equal(t, []int{0}, e2.ResultCodes())
}

func TestBlockParallelRunAll_WithFailure(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := Cmd(echoBinary, echoArg1).Outputs(&stdout1, &stderr1).Retries(2, 100)

	//echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := Cmd("/bin/false").Outputs(&stdout2, &stderr2).Retries(2, 100)

	val, err := blockParallelRunAll(-1, e1, e2)
	require.NoError(t, err)
	require.NotNil(t, val)
	assert.Equal(t, []int{0, 1}, val)
	assert.Equal(t, []int{0}, e1.ResultCodes())
	assert.Equal(t, []int{1, 1, 1}, e2.ResultCodes())
}

func TestAsyncRunBest(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := Cmd(echoBinary, echoArg1).Outputs(&stdout1, &stderr1).Retries(2, 100)

	echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := Cmd(echoBinary, echoArg2).Outputs(&stdout2, &stderr2).Retries(2, 100)

	p := AsyncRunBest(e1, e2)

	br, err := WaitBestResults(p)
	require.NoError(t, err)
	require.NotNil(t, br)
	assert.False(t, br.DidError())
	assert.Equal(t, 2, br.Len())

	val, err := br.Result(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, val)
	assert.Equal(t, []int{0}, e1.ResultCodes())

	val, err = br.Result(1)
	assert.NoError(t, err)
	assert.Equal(t, 0, val)
	assert.Equal(t, []int{0}, e2.ResultCodes())
}

func TestAsyncRunBest_WithFailure(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := Cmd(echoBinary, echoArg1).Outputs(&stdout1, &stderr1).Retries(2, 100)

	//echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := Cmd("/bin/false").Outputs(&stdout2, &stderr2).Retries(2, 100)

	p := AsyncRunBest(e1, e2)

	br, err := WaitBestResults(p)
	require.NoError(t, err)
	require.NotNil(t, br)
	assert.False(t, br.DidError())
	assert.Equal(t, 2, br.Len())

	val, err := br.Result(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, val)
	assert.Equal(t, []int{0}, e1.ResultCodes())

	val, err = br.Result(1)
	assert.NoError(t, err)
	assert.Equal(t, 1, val)
	assert.Equal(t, []int{1, 1, 1}, e2.ResultCodes())
}

func TestBlockSerial(t *testing.T) {
	e1 := Cmd("echo", "foo")
	e2 := Cmd("echo", "bar")
	f1 := Cmd("false")

	rc, err := blockSerial(false, e1, e2)
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo\n", e1.StdoutRecord())
	assert.Equal(t, "bar\n", e2.StdoutRecord())

	e1.reset()
	f1.reset()
	e2.reset()
	rc, err = blockSerial(false, f1, e1)
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo\n", e1.StdoutRecord())

	e1.reset()
	f1.reset()
	e2.reset()
	rc, err = blockSerial(false, e1, f1)
	require.NoError(t, err)
	assert.Equal(t, 1, rc)
	assert.Equal(t, "foo\n", e1.StdoutRecord())

	e1.reset()
	f1.reset()
	e2.reset()

	rc, err = blockSerial(true, f1, e1)
	require.NoError(t, err)
	assert.Equal(t, 1, rc)
	assert.Equal(t, "", e1.StdoutRecord())
}

func TestBlockParallel(t *testing.T) {
	e1 := Cmd("/bin/sh", "-c", "sleep 0.1 ; echo foo")
	e2 := Cmd("/bin/sh", "-c", "sleep 0.1 ; echo bar")
	f1 := Cmd("false")

	rc, err := blockParallel(false, -1, e1, e2)
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo\n", e1.StdoutRecord())
	assert.Equal(t, "bar\n", e2.StdoutRecord())

	e1.reset()
	f1.reset()
	e2.reset()
	rc, err = blockParallel(false, -1, f1, e1, e2)
	require.NoError(t, err)
	assert.Equal(t, 1, rc)
	assert.Equal(t, "foo\n", e1.StdoutRecord())
	assert.Equal(t, "bar\n", e2.StdoutRecord())

	e1.reset()
	f1.reset()
	e2.reset()
	rc, err = blockParallel(false, -1, e1, f1, e2)
	require.NoError(t, err)
	assert.Equal(t, 1, rc)
	assert.Equal(t, "foo\n", e1.StdoutRecord())
	assert.Equal(t, "bar\n", e2.StdoutRecord())

	e1.reset()
	f1.reset()
	e2.reset()
	// Should fail fast before echoing to stdout
	rc, err = blockParallel(true, -1, e1, e2, f1)
	require.NoError(t, err)
	assert.Equal(t, 1, rc)
	assert.Equal(t, "", e1.StdoutRecord())
	assert.Equal(t, "", e2.StdoutRecord())

}
