package cmdz

import (
	//"fmt"
	//"io"
	"bytes"
	"context"

	//"log"
	//"os/exec"
	"strings"
	"testing"
	"time"

	"mby.fr/utils/inout"
	"mby.fr/utils/promise"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmd(t *testing.T) {
	c := Cmd("echo", "foo")
	assert.NotNil(t, c)
	//assert.NotNil(t, c.cmd)
	assert.NotNil(t, c.config)
	assert.NotNil(t, c.stdoutRecord)
	assert.NotNil(t, c.stderrRecord)
	assert.Len(t, c.ResultCodes(), 0)
	assert.Len(t, c.Executions(), 0)
}

func TestCmdCtx(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	c := CmdCtx(ctx, "echo", "foo")
	assert.NotNil(t, c)
	//assert.NotNil(t, c.cmd)
	assert.NotNil(t, c.config)
	assert.NotNil(t, c.stdoutRecord)
	assert.NotNil(t, c.stderrRecord)
	assert.Len(t, c.ResultCodes(), 0)
	assert.Len(t, c.Executions(), 0)
}

func TestString(t *testing.T) {
	c := Cmd("echo", "foo")
	assert.Equal(t, "echo foo", c.String())

	c.AddArgs("bar", "$val")
	assert.Equal(t, "echo foo bar $val", c.String())

	c.AddEnv("val", "baz")
	assert.Equal(t, "echo foo bar $val", c.String())
}

func TestBlockRun(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg := "foobar"
	e := Cmd(echoBinary, echoArg)

	rc, err := e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e.ResultCodes())

	e2 := Cmd("false")
	rc2, err2 := e2.BlockRun()
	require.NoError(t, err2, "should not error")
	assert.Equal(t, 1, rc2)
	assert.Equal(t, []int{1}, e2.ResultCodes())
}

func TestBlockRun_WithInput(t *testing.T) {
	echoBinary := "/bin/cat"
	input1 := "foo"
	input2 := "bar"
	reader := bytes.NewBufferString(input1)
	e := Cmd(echoBinary, "-").SetInput(reader)

	rc, err := e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e.ResultCodes())
	sin := e.StdinRecord()
	sout := e.StdoutRecord()
	serr := e.StderrRecord()
	assert.Equal(t, input1, sin)
	assert.Equal(t, input1, sout)
	assert.Equal(t, "", serr)

	reader.Reset()
	reader.WriteString(input2)
	e.reset()
	rc, err = e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, 0, rc)
	sin = e.StdinRecord()
	sout = e.StdoutRecord()
	serr = e.StderrRecord()
	assert.Equal(t, input2, sin)
	assert.Equal(t, input2, sout)
	assert.Equal(t, "", serr)
}

func TestBlockRun_WithOutputs(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg := "foobar"
	stdout := strings.Builder{}
	stderr := strings.Builder{}
	e := Cmd(echoBinary, echoArg).SetOutputs(&stdout, &stderr)

	rc, err := e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e.ResultCodes())
	sout := stdout.String()
	serr := stderr.String()
	assert.Equal(t, echoArg+"\n", sout)
	assert.Equal(t, "", serr)

	// Re run
	rc, err = e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e.ResultCodes())
	sout = stdout.String()
	serr = stderr.String()
	assert.Equal(t, echoArg+"\n"+echoArg+"\n", sout)
	assert.Equal(t, "", serr)
}

func TestBlockRun_RecordingOutputs(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg := "foobar"
	e := Cmd(echoBinary, echoArg)

	rc, err := e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e.ResultCodes())
	sout := e.StdoutRecord()
	serr := e.StderrRecord()
	assert.Equal(t, echoArg+"\n", sout)
	assert.Equal(t, "", serr)

	// Re run
	rc, err = e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e.ResultCodes())
	sout = e.StdoutRecord()
	serr = e.StderrRecord()
	assert.Equal(t, echoArg+"\n", sout)
	assert.Equal(t, "", serr)
}

func TestBlockRun_AddArgs(t *testing.T) {
	echoBinary := "/bin/echo"
	e := Cmd(echoBinary).AddArgs("foo", "bar")

	rc, err := e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e.ResultCodes())
	assert.Equal(t, "foo bar\n", e.StdoutRecord())
	assert.Equal(t, "", e.StderrRecord())
}

func TestBlockRun_AddEnv(t *testing.T) {
	e := Cmd("/bin/sh", "-c", "echo foo $VALUE").AddEnv("VALUE", "baz")

	rc, err := e.BlockRun()
	require.NoError(t, err, "should not error")
	assert.Equal(t, 0, rc)
	assert.Equal(t, []int{0}, e.ResultCodes())
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
	assert.Equal(t, []int{0}, e.ResultCodes())
	sout := e.StdoutRecord()
	serr := e.StderrRecord()
	assert.Equal(t, echoArg+"\n", sout)
	assert.Equal(t, "", serr)

	rc3, err3 := e.BlockRun()
	require.NoError(t, err3, "should not error")
	assert.Equal(t, 0, rc3)
	assert.Equal(t, []int{0}, e.ResultCodes())
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
	assert.Equal(t, []int{1, 1, 1}, e.ResultCodes())
	sout := e.StdoutRecord()
	serr := e.StderrRecord()
	assert.Equal(t, "", sout)
	assert.Equal(t, "", serr)
}

func TestBlockRun_ErrorOnFailure(t *testing.T) {
	f := Cmd("/bin/false").ErrorOnFailure(true)
	rc, err := f.BlockRun()
	require.Error(t, err, "should error")
	assert.Equal(t, -1, rc)
}

func TestBlockRun_CombinedOutputs(t *testing.T) {
	f := Sh("echo foo ; sleep 0.02 ; >&2 echo bar").CombinedOutputs()
	rc, err := f.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo\nbar\n", f.StdoutRecord())
	assert.Equal(t, "", f.StderrRecord())
}

func TestBlockRun_Timeout_ok(t *testing.T) {
	f := Sh("sleep 0.02 ; echo bar")
	f.Timeout(100)
	rc, err := f.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "bar\n", f.StdoutRecord())
}

func TestBlockRun_Timeout_ko(t *testing.T) {
	f := Sh("sleep 0.02 ; echo bar")
	f.Timeout(10)
	rc, err := f.BlockRun()
	require.Error(t, err)
	assert.Equal(t, -1, rc)
	assert.Equal(t, "", f.StdoutRecord())
}

func TestAsyncRun(t *testing.T) {
	echoBinary := "/bin/echo"
	echoArg1 := "foobar"
	stdout1 := strings.Builder{}
	stderr1 := strings.Builder{}
	e1 := Cmd(echoBinary, echoArg1).SetOutputs(&stdout1, &stderr1).Retries(2, 100)

	echoArg2 := "foobaz"
	stdout2 := strings.Builder{}
	stderr2 := strings.Builder{}
	e2 := Cmd(echoBinary, echoArg2).SetOutputs(&stdout2, &stderr2).Retries(3, 100)

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
	assert.Equal(t, []int{0}, e1.ResultCodes())
	assert.Equal(t, []int{0}, e2.ResultCodes())

	assert.Equal(t, echoArg1+"\n", stdout1.String())
	assert.Equal(t, "", stderr1.String())
	assert.Equal(t, echoArg2+"\n", stdout2.String())
	assert.Equal(t, "", stderr2.String())
}

func TestPipe_OutputString(t *testing.T) {
	echo := Cmd("/bin/echo", "-n", "foo")
	sed := Cmd("/bin/sed", "-e", "s/o/a/")
	p := echo.Pipe(sed)

	o, err := Outputted(p).OutputString()
	require.NoError(t, err)
	assert.Equal(t, "fao", o)
	assert.Equal(t, "foo", echo.StdoutRecord())
	assert.Equal(t, "foo", sed.StdinRecord())
	assert.Equal(t, "fao", sed.StdoutRecord())
	assert.Equal(t, "fao", p.StdoutRecord())

	fail := Cmd("/bin/false")
	echo = Cmd("/bin/echo", "-n", "foo")
	p = fail.Pipe(echo)

	o, err = Outputted(p).OutputString()
	require.NoError(t, err)
	assert.Equal(t, "foo", o)
	assert.Equal(t, "", fail.StdoutRecord())
	assert.Equal(t, "", echo.StdinRecord())
	assert.Equal(t, "foo", echo.StdoutRecord())
	assert.Equal(t, "foo", p.StdoutRecord())

	errCmd := Cmd("doNotExists")
	echo = Cmd("/bin/echo", "-n", "foo")
	p = errCmd.Pipe(echo)

	o, err = Outputted(p).OutputString()
	require.Error(t, err)
	assert.ErrorContains(t, err, "doNotExists")
	assert.Equal(t, "", o)
	assert.Equal(t, "", errCmd.StdoutRecord())
	assert.Equal(t, "", echo.StdinRecord())
	assert.Equal(t, "", echo.StdoutRecord())
	assert.Equal(t, "", p.StdoutRecord())
}

func TestPipeFail_OutputString(t *testing.T) {
	echo := Cmd("/bin/echo", "-n", "foo")
	sed := Cmd("/bin/sed", "-e", "s/o/a/")
	p := echo.PipeFail(sed)

	o, err := Outputted(p).OutputString()
	require.NoError(t, err)
	assert.Equal(t, "fao", o)
	assert.Equal(t, "foo", echo.StdoutRecord())
	assert.Equal(t, "foo", sed.StdinRecord())
	assert.Equal(t, "fao", sed.StdoutRecord())
	assert.Equal(t, "fao", p.StdoutRecord())

	fail := Cmd("/bin/false")
	echo = Cmd("/bin/echo", "-n", "foo")
	p = fail.PipeFail(echo)

	o, err = Outputted(p).OutputString()
	require.Error(t, err)
	assert.IsType(t, err, failure{})
	assert.Equal(t, "", o)
	assert.Equal(t, "", fail.StdoutRecord())
	assert.Equal(t, "", echo.StdinRecord())
	assert.Equal(t, "", echo.StdoutRecord())
	assert.Equal(t, "", p.StdoutRecord())

	errCmd := Cmd("doNotExists")
	echo = Cmd("/bin/echo", "-n", "foo")
	p = errCmd.Pipe(echo)

	o, err = Outputted(p).OutputString()
	require.Error(t, err)
	assert.ErrorContains(t, err, "doNotExists")
	assert.Equal(t, "", o)
	assert.Equal(t, "", errCmd.StdoutRecord())
	assert.Equal(t, "", echo.StdinRecord())
	assert.Equal(t, "", echo.StdoutRecord())
	assert.Equal(t, "", p.StdoutRecord())
}

func TestReportError(t *testing.T) {
	// TODO
}

func TestSh(t *testing.T) {
	c := Sh(`echo -e "foo\nbar\nbaz" | grep 'bar'`)
	rc, err := c.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "bar\n", c.StdoutRecord())

	c = Sh(`true || echo "foo"`)
	rc, err = c.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "", c.StdoutRecord())

	c = Sh(`false || echo "foo"`)
	rc, err = c.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo\n", c.StdoutRecord())

	c = Sh(`true && echo "bar"`)
	rc, err = c.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "bar\n", c.StdoutRecord())

	c = Sh(`false && echo "bar"`)
	rc, err = c.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 1, rc)
	assert.Equal(t, "", c.StdoutRecord())

	c = Sh(`>&2 echo "bar"`)
	rc, err = c.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "", c.StdoutRecord())
	assert.Equal(t, "bar\n", c.StderrRecord())
}

func TestInputProcesser(t *testing.T) {
	// No input processer
	c := Cmd("cat")
	r := strings.NewReader("foo")
	c.SetInput(r)
	rc, err := c.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo", c.StdinRecord())
	assert.Equal(t, "foo", c.StdoutRecord())
	assert.Equal(t, "", c.StderrRecord())

	// No input processer
	p := Cmd("cat")
	r = strings.NewReader("foo")
	p.SetInput(r)
	rc, err = p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "foo", p.StdinRecord())
	assert.Equal(t, "foo", p.StdoutRecord())
	assert.Equal(t, "", p.StderrRecord())

	// Simple prefix processer
	prefixProcessor := inout.StringLineProcesser(func(in string) (out string, err error) {
		return "PREFIX" + in, nil
	})
	p = Cmd("cat")
	r = strings.NewReader("foo\n")
	p.ProcessIn(prefixProcessor)
	p.SetInput(r)
	rc, err = p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "PREFIXfoo\n", p.StdinRecord())
	assert.Equal(t, "PREFIXfoo\n", p.StdoutRecord())
	assert.Equal(t, "", p.StderrRecord())

	// Reverse config order
	prefixProcessor = inout.StringLineProcesser(func(in string) (out string, err error) {
		return "PREFIX" + in, nil
	})
	p = Cmd("cat")
	r = strings.NewReader("foo\n")
	p.SetInput(r)
	p.ProcessIn(prefixProcessor)
	rc, err = p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "PREFIXfoo\n", p.StdinRecord())
	assert.Equal(t, "PREFIXfoo\n", p.StdoutRecord())
	assert.Equal(t, "", p.StderrRecord())
}

func TestOutputProcesser(t *testing.T) {
	// No input processer
	c := Sh("echo foo")
	wout := strings.Builder{}
	werr := strings.Builder{}
	c.SetOutputs(&wout, &werr)
	rc, err := c.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "", c.StdinRecord())
	assert.Equal(t, "foo\n", c.StdoutRecord())
	assert.Equal(t, "", c.StderrRecord())
	assert.Equal(t, "foo\n", wout.String())
	assert.Equal(t, "", werr.String())

	// No output processer
	p := Sh("echo foo")
	wout = strings.Builder{}
	werr = strings.Builder{}
	p.SetOutputs(&wout, &werr)
	rc, err = p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "", p.StdinRecord())
	assert.Equal(t, "foo\n", p.StdoutRecord())
	assert.Equal(t, "", p.StderrRecord())
	assert.Equal(t, "foo\n", wout.String())
	assert.Equal(t, "", werr.String())

	// Simple prefix processer
	prefixProcessor := inout.StringLineProcesser(func(in string) (out string, err error) {
		return "PREFIX" + in, nil
	})
	p = Sh("echo bar; >&2 echo baz")
	wout = strings.Builder{}
	werr = strings.Builder{}
	p.ProcessOut(prefixProcessor)
	p.SetOutputs(&wout, &werr)
	rc, err = p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "", p.StdinRecord())
	assert.Equal(t, "PREFIXbar\n", p.StdoutRecord())
	assert.Equal(t, "baz\n", p.StderrRecord())
	assert.Equal(t, "PREFIXbar\n", wout.String())
	assert.Equal(t, "baz\n", werr.String())

	// Reverse config order
	prefixProcessor = inout.StringLineProcesser(func(in string) (out string, err error) {
		return "PREFIX" + in, nil
	})
	p = Sh("echo bar")
	wout = strings.Builder{}
	werr = strings.Builder{}
	p.SetOutputs(&wout, &werr)
	p.ProcessOut(prefixProcessor)
	rc, err = p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "", p.StdinRecord())
	assert.Equal(t, "PREFIXbar\n", p.StdoutRecord())
	assert.Equal(t, "", p.StderrRecord())
	assert.Equal(t, "PREFIXbar\n", wout.String())
	assert.Equal(t, "", werr.String())

	// Simple prefix processer
	prefixProcessor = inout.StringLineProcesser(func(in string) (out string, err error) {
		return "PREFIX" + in, nil
	})
	p = Sh(">&2 echo baz")
	wout = strings.Builder{}
	werr = strings.Builder{}
	p.ProcessErr(prefixProcessor)
	p.SetOutputs(&wout, &werr)
	rc, err = p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "", p.StdinRecord())
	assert.Equal(t, "", p.StdoutRecord())
	assert.Equal(t, "PREFIXbaz\n", p.StderrRecord())
	assert.Equal(t, "", wout.String())
	assert.Equal(t, "PREFIXbaz\n", werr.String())

	// Reverse config order
	prefixProcessor = inout.StringLineProcesser(func(in string) (out string, err error) {
		return "PREFIX" + in, nil
	})
	p = Sh(">&2 echo baz")
	wout = strings.Builder{}
	werr = strings.Builder{}
	p.SetOutputs(&wout, &werr)
	p.ProcessErr(prefixProcessor)
	rc, err = p.BlockRun()
	require.NoError(t, err)
	assert.Equal(t, 0, rc)
	assert.Equal(t, "", p.StdinRecord())
	assert.Equal(t, "", p.StdoutRecord())
	assert.Equal(t, "PREFIXbaz\n", p.StderrRecord())
	assert.Equal(t, "", wout.String())
	assert.Equal(t, "PREFIXbaz\n", werr.String())
}
