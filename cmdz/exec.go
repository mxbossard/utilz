package cmdz

import (
	//"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"mby.fr/utils/inout"
	"mby.fr/utils/promise"
)

type ExecPromise = promise.Promise[int]
type ExecsPromise = promise.Promise[[]int]

type Exec struct {
	*exec.Cmd

	StdoutRecord inout.RecordingWriter
	StderrRecord inout.RecordingWriter

	Retries        int
	RetryDelayInMs int64
	ResultsCodes   []int
}

func (e *Exec) AddEnv(key, value string) {
	entry := fmt.Sprintf("%s=%s", key, value)
	e.Env = append(e.Env, entry)
}

func (e *Exec) AddArgs(args ...string) {
	e.Args = append(e.Args, args...)
}

func (e *Exec) RecordingOutputs(stdout, stderr io.Writer) {
	e.StdoutRecord.Nested = stdout
	e.StderrRecord.Nested = stderr
	e.Stdout = &e.StdoutRecord
	e.Stderr = &e.StderrRecord
}

/*
func (e Exec) FlushOutputs() (err error) {
	if e.StdoutBuffer != nil {
		err = e.StdoutBuffer.Flush()
		if err != nil {
			return
		}
	}
	if e.StderrBuffer != nil {
		err = e.StderrBuffer.Flush()
		if err != nil {
			return
		}
	}
	return
}
*/

func (e Exec) String() (t string) {
	t = strings.Join(e.Args, " ")
	return
}

func (e *Exec) BlockRun() (rc int, err error) {
	err = e.Run()
	rc = -1
	for i := 0; i <= e.Retries && rc != 0; i++ {
		if i > 0 {
			// Wait between retries
			time.Sleep(time.Duration(e.RetryDelayInMs) * time.Millisecond)
		}
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				rc = exitErr.ProcessState.ExitCode()
				err = nil
			} else {
				return -1, err
			}
		} else {
			rc = e.ProcessState.ExitCode()
		}
		e.ResultsCodes = append(e.ResultsCodes, rc)
	}
	//e.FlushOutputs()
	return rc, nil
}

func (e *Exec) AsyncRun() *ExecPromise {
	p := promise.New(func(resolve func(int), reject func(error)) {
		rc, err := e.BlockRun()
		if err != nil {
			reject(err)
		}
		resolve(rc)
	})
	return p
}

func Execution(binary string, args ...string) *Exec {
	cmd := exec.Command(binary, args...)
	e := Exec{Cmd: cmd, RetryDelayInMs: 100}
	e.RecordingOutputs(cmd.Stdout, cmd.Stderr)
	return &e
}

func ExecutionOutputs(stdout, stderr io.Writer, binary string, args ...string) *Exec {
	e := Execution(binary, args...)
	e.RecordingOutputs(stdout, stderr)
	return e
}

func ExecutionContext(ctx context.Context, binary string, args ...string) *Exec {
	cmd := exec.CommandContext(ctx, binary, args...)
	e := Exec{Cmd: cmd}
	return &e
}

func AsyncRunAll(execs ...*Exec) *ExecsPromise {
	var promises []*promise.Promise[int]
	for _, e := range execs {
		p := e.AsyncRun()
		promises = append(promises, p)
	}

	ctx := context.Background()
	p := promise.All[int](ctx, promises...)
	return p
}

func WaitAllResults(p *ExecsPromise) (*[]int, error) {
	ctx := context.Background()
	return p.Await(ctx)
}

func ParallelRunAll(forkCount int, execs ...*Exec) ([]int, error) {
	p := AsyncRunAll(execs...)
	br, err := WaitAllResults(p)
	if err != nil {
		return nil, err
	}

	return *br, nil
}

func AsyncRunBest(execs ...*Exec) *promise.Promise[promise.BestResults[int]] {
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
	if err != nil {
		return nil, err
	}
	if br.DidError() {
		return nil, br.AggError()
	}
	return br, err
}

func Failed(resultCodes ...int) bool {
	for _, rc := range resultCodes {
		if rc != 0 {
			return true
		}
	}
	return false
}

func Succeed(resultCodes ...int) bool {
	return !Failed(resultCodes...)
}
