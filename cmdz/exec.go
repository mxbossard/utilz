package cmdz

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

type FutureExec struct {
	exec Exec
}

func (f FutureExec) Wait() (rc int, err error) {
	return
}

type Exec struct {
	*exec.Cmd
}

func (e *Exec) AddEnv(key, value string) {
	entry := fmt.Sprintf("%s=%s", key, value)
	e.Env = append(e.Env, entry)
}

func (e *Exec) AddArgs(args ...string) {
	e.Args = append(e.Args, args...)
}

func (e Exec) Trace() (t string) {
	return
}

func (e Exec) WithStdOuts() (rc int, err error) {
	return
}

func (e Exec) RunBlocking(stdout, stderr io.Writer) (rc int, err error) {
	e.Stdout = stdout
	e.Stderr = stderr

	err = e.Run()
	rc = e.ProcessState.ExitCode()
	return
}

func (e Exec) RunAsync(stdout, stderr io.Writer) (future FutureExec) {
	return
}

func Execution(binary string, args ...string) Exec {
	cmd := exec.Command(binary, args...)
	e := Exec{Cmd: cmd}
	return e
}

func ExecutionContext(ctx context.Context, binary string, args ...string) Exec {
	cmd := exec.CommandContext(ctx, binary, args...)
	e := Exec{Cmd: cmd}
	return e
}
