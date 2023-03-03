package cmdz

import (
	//"fmt"
	//"io"
	"context"
	//"log"
	//"os/exec"
	"strings"
	"testing"
	"time"

	"mby.fr/utils/promise"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmd(t *testing.T) {
	c := Cmd("echo", "foo")
	assert.NotNil(t, c)
	assert.NotNil(t, c.Cmd)
	assert.NotNil(t, c.config)
	assert.NotNil(t, c.stdoutRecord)
	assert.NotNil(t, c.stderrRecord)
	assert.Len(t, c.ResultsCodes, 0)
	assert.Len(t, c.Executions, 0)
}

func TestCmdCtx(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	c := CmdCtx(ctx, "echo", "foo")
	assert.NotNil(t, c)
	assert.NotNil(t, c.Cmd)
	assert.NotNil(t, c.config)
	assert.NotNil(t, c.stdoutRecord)
	assert.NotNil(t, c.stderrRecord)
	assert.Len(t, c.ResultsCodes, 0)
	assert.Len(t, c.Executions, 0)
}

func TestString(t *testing.T) {
	c := Cmd("echo", "foo")
	assert.Equal(t, "echo foo", c.String())

	c.AddArgs("bar", "$val")
	assert.Equal(t, "echo foo bar $val", c.String())

	c.AddEnv("val", "baz")
	assert.Equal(t, "echo foo bar $val", c.String())
}

func TestAddArgs(t *testing.T) {
	e := Cmd("echo")
	assert.Equal(t, []string{"echo"}, e.Args)
	e.AddArgs("foo")
	assert.Equal(t, []string{"echo", "foo"}, e.Args)
	e.AddArgs("bar", "baz")
	assert.Equal(t, []string{"echo", "foo", "bar", "baz"}, e.Args)
}

func TestBlockRun(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg := "foobar"
	e := Cmd(echoBinary, echoArg)

	rc, err := e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e.ResultsCodes)

	e2 := Cmd("false")
	rc2, err2 := e2.BlockRun()
	require.NoError(t, err2, "should not error")
	assert.Equal(t, 1, rc2)
	assert.Equal(t, []int{1}, e2.ResultsCodes)
}

func TestBlockRun_WithOutputs(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg := "foobar"
	stdout := strings.Builder{}
	stderr := strings.Builder{}
	e := Cmd(echoBinary, echoArg).Outputs(&stdout, &stderr)

	rc, err := e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e.ResultsCodes)
	sout := stdout.String()
	serr := stderr.String()
	assert.Equal(t, echoArg+"\n", sout)
	assert.Equal(t, "", serr)
}

func TestBlockRun_RecordingOutputs(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg := "foobar"
	e := Cmd(echoBinary, echoArg)

	rc, err := e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e.ResultsCodes)
	sout := e.StdoutRecord()
	serr := e.StderrRecord()
	assert.Equal(t, echoArg+"\n", sout)
	assert.Equal(t, "", serr)
}

func TestBlockRun_WithEnv(t *testing.T) {
	e := Cmd("/bin/sh", "-c", "echo foo $VALUE").AddEnv("VALUE", "baz")
	//e := Cmd("/bin/sh", "-c", "export").AddEnv("VALUE", "baz")

	rc, err := e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e.ResultsCodes)
	assert.Equal(t, "foo baz\n", e.StdoutRecord())
	assert.Equal(t, "", e.StderrRecord())
}

func TesBlockRun_ReRun(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg := "foobar"
	e := Cmd(echoBinary, echoArg)

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
	e := Cmd(echoBinary).Retries(2, 100)

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
	e1 := Cmd(echoBinary, echoArg1).Outputs(&stdout1, &stderr1).Retries(2, 100)

	echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := Cmd(echoBinary, echoArg2).Outputs(&stdout2, &stderr2).Retries(3, 100)

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
