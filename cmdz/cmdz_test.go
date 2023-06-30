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
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e.ResultsCodes)
	sout := stdout.String()
	serr := stderr.String()
	assert.Equal(t, echoArg+"\n", sout)
	assert.Equal(t, "", serr)
}

func TesBlockRun_ReRun(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg := "foobar"
	e := Execution(echoBinary, echoArg)

	rc, err := e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e.ResultsCodes)
	sout := e.StdoutRecord()
	serr := e.StderrRecord()
	assert.Equal(t, echoArg+"\n", sout)
	assert.Equal(t, "", serr)

	rc3, err3 := e.BlockRun()
	require.NoError(t, err3, "should not error")
	assert.Equal(t, 0, rc3)
	assert.Equal(t, []int{0}, e.ResultsCodes)
	sout3 := e.StdoutRecord()
	serr3 := e.StderrRecord()
	assert.Equal(t, echoArg+"\n", sout3)
	assert.Equal(t, "", serr3)
}

func TestBlockRun_Retries(t *testing.T) {
	echoBinary := "/bin/false"
	e := Execution(echoBinary)
	e.Retries = 2

	rc, err := e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, 1, rc)
	assert.Equal(t, []int{1, 1, 1}, e.ResultsCodes)
	sout := e.StdoutRecord()
	serr := e.StderrRecord()
	assert.Equal(t, "", sout)
	assert.Equal(t, "", serr)
}

func TestAsyncRun(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := ExecutionOutputs(&stdout1, &stderr1, echoBinary, echoArg1)
	e1.Retries = 2

	echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := ExecutionOutputs(&stdout2, &stderr2, echoBinary, echoArg2)
	e2.Retries = 3

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
	assert.Equal(t, []int{0}, e1.ResultsCodes)
	assert.Equal(t, []int{0}, e2.ResultsCodes)

	assert.Equal(t, echoArg1+"\n", stdout1.String())
	assert.Equal(t, "", stderr1.String())
	assert.Equal(t, echoArg2+"\n", stdout2.String())
	assert.Equal(t, "", stderr2.String())
}
