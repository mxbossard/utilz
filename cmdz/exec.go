package cmdz

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"mby.fr/utils/promise"
)

type ExecPromise promise.Promise[int]
type ExecPromises promise.Promise[[]int]

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

func (e Exec) String() (t string) {
	t = strings.Join(e.Args, " ")
	return
}

func (e Exec) WithStdOuts() (rc int, err error) {
	return
}

func (e Exec) BlockRun() (rc int, err error) {
	err = e.Run()
	rc = e.ProcessState.ExitCode()
	return
}

func (e Exec) AsyncRun() *promise.Promise[int] {
	p := promise.New(func(resolve func(int), reject func(error)) {
		rc, err := e.BlockRun()
		if err != nil {
			reject(err)
		}
		resolve(rc)
	})
	return p
}

func Execution(binary string, args ...string) Exec {
	cmd := exec.Command(binary, args...)
	e := Exec{Cmd: cmd}
	return e
}

func ExecutionOutputs(stdout, stderr io.Writer, binary string, args ...string) Exec {
	cmd := exec.Command(binary, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	e := Exec{Cmd: cmd}
	return e
}

func ExecutionContext(ctx context.Context, binary string, args ...string) Exec {
	cmd := exec.CommandContext(ctx, binary, args...)
	e := Exec{Cmd: cmd}
	return e
}

func AsyncRunAll(execs ...Exec) *promise.Promise[[]int] {
	var promises []*promise.Promise[int]
	for _, e := range execs {
		p := e.AsyncRun()
		promises = append(promises, p)
	}

	ctx := context.Background()
	p := promise.All[int](ctx, promises...)
	return p
}

func WaitAllResults(p *promise.Promise[[]int]) (*[]int, error) {
	ctx := context.Background()
	return p.Await(ctx)
}

func AsyncRunBest(execs ...Exec) *promise.Promise[promise.BestResults[int]] {
	var promises []*promise.Promise[int]
	for _, e := range execs {
		p := e.AsyncRun()
		promises = append(promises, p)
	}

	ctx := context.Background()
	p := promise.Best[int](ctx, promises...)
	return p
}

func WaitBestResults(p *promise.Promise[promise.BestResults[int]]) (*promise.BestResults[int], error) {
	ctx := context.Background()
	br, err := p.Await(ctx)
	return br, err
}
