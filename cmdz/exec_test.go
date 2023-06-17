package cmdz

import (
	//"fmt"
	//"io"
	"context"
	//"log"
	//"os/exec"
	"strings"
	"testing"

	"mby.fr/utils/promise"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddArgs(t *testing.T) {
	e := Execution("echo")
	assert.Equal(t, []string{"echo"}, e.Args)
	e.AddArgs("foo")
	assert.Equal(t, []string{"echo", "foo"}, e.Args)
	e.AddArgs("bar", "baz")
	assert.Equal(t, []string{"echo", "foo", "bar", "baz"}, e.Args)
}

func TestBlockRun(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg := "foobar"
	stdout := strings.Builder{}
	stderr := strings.Builder{}
	e := ExecutionOutputs(&stdout, &stderr, echoBinary, echoArg)

	rc, err := e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, rc, 0)
	sout := stdout.String()
	serr := stderr.String()
	assert.Equal(t, echoArg+"\n", sout)
	assert.Equal(t, "", serr)
}

func TestAsynckRun(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := ExecutionOutputs(&stdout1, &stderr1, echoBinary, echoArg1)

	echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := ExecutionOutputs(&stdout2, &stderr2, echoBinary, echoArg2)

	p1 := e1.AsyncRun()
	p2 := e2.AsyncRun()

	assert.Equal(t, "", stdout1.String())
	assert.Equal(t, "", stderr1.String())
	assert.Equal(t, "", stdout2.String())
	assert.Equal(t, "", stderr2.String())

	ctx := context.Background()
	p := promise.All(ctx, p1, p2)

	val, err := p.Await(ctx)
	require.NoError(t, err)
	require.NotNil(t, val)
	assert.Equal(t, []int{0, 0}, *val)

	assert.Equal(t, echoArg1+"\n", stdout1.String())
	assert.Equal(t, "", stderr1.String())
	assert.Equal(t, echoArg2+"\n", stdout2.String())
	assert.Equal(t, "", stderr2.String())
}

func TestAsynckRunAll(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := ExecutionOutputs(&stdout1, &stderr1, echoBinary, echoArg1)

	echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := ExecutionOutputs(&stdout2, &stderr2, echoBinary, echoArg2)

	p := AsyncRunAll(e1, e2)

	val, err := WaitAllResults(p)
	require.NoError(t, err)
	require.NotNil(t, val)
	assert.Equal(t, []int{0, 0}, *val)
}

func TestAsynckRunAll_WithFailure(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := ExecutionOutputs(&stdout1, &stderr1, echoBinary, echoArg1)

	//echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := ExecutionOutputs(&stdout2, &stderr2, "/bin/false")

	p := AsyncRunAll(e1, e2)

	val, err := WaitAllResults(p)
	require.NoError(t, err)
	assert.Equal(t, []int{0, 1}, *val)
}

func TestAsynckRunBest(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := ExecutionOutputs(&stdout1, &stderr1, echoBinary, echoArg1)

	echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := ExecutionOutputs(&stdout2, &stderr2, echoBinary, echoArg2)

	p := AsyncRunBest(e1, e2)

	br, err := WaitBestResults(p)
	require.NoError(t, err)
	require.NotNil(t, br)
	assert.False(t, br.DidError())
	assert.Equal(t, 2, br.Len())

	val, err := br.Result(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, val)

	val, err = br.Result(1)
	assert.NoError(t, err)
	assert.Equal(t, 0, val)
}

func TestAsynckRunBest_WithFailure(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := ExecutionOutputs(&stdout1, &stderr1, echoBinary, echoArg1)

	//echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := ExecutionOutputs(&stdout2, &stderr2, "/bin/false")

	p := AsyncRunBest(e1, e2)

	br, err := WaitBestResults(p)
	require.NoError(t, err)
	require.NotNil(t, br)
	assert.False(t, br.DidError())
	assert.Equal(t, 2, br.Len())

	val, err := br.Result(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, val)

	val, err = br.Result(1)
	assert.NoError(t, err)
	assert.Equal(t, 1, val)

}
