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

type Exec struct {
	*exec.Cmd

	StdoutRecord inout.RecordingWriter
	StderrRecord inout.RecordingWriter

	Retries        int
	RetryDelayInMs int64

	ResultsCodes []int
	// FIXME: replace ResultsCodes by Executions
	Executions []*exec.Cmd
}

func (e *Exec) AddEnv(key, value string) {
	entry := fmt.Sprintf("%s=%s", key, value)
	e.Env = append(e.Env, entry)
}

func (e *Exec) AddArgs(args ...string) {
	e.Args = append(e.Args, args...)
}

func (e *Exec) RecordingOutputs(stdout, stderr io.Writer) {
	if stdout != nil {
		e.StdoutRecord.Nested = stdout
		e.Stdout = &e.StdoutRecord
	}
	if stderr != nil {
		e.StderrRecord.Nested = stderr
		e.Stderr = &e.StderrRecord
	}
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

func (e Exec) ReportError() string {
	execCmdSummary := e.String()
	attempts := len(e.ResultsCodes)
	status := e.ResultsCodes[attempts-1]
	stderr := e.StderrRecord.String()
	errorMessage := fmt.Sprintf("Exec failed after %d attempt(s): [%s] !\nRC=%d ERR> %s", attempts, execCmdSummary, status, strings.TrimSpace(stderr))
	return errorMessage
}

func (e *Exec) BlockRun() (rc int, err error) {
	clone := *e.Cmd
	rc = -1
	for i := 0; i <= e.Retries && rc != 0; i++ {
		if i > 0 {
			// Wait between retries
			time.Sleep(time.Duration(e.RetryDelayInMs) * time.Millisecond)
		}
		newClone := clone
		e.Cmd = &newClone
		err = e.Start()
		if err != nil {
			return
		}
		err = e.Wait()
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
		e.Executions = append(e.Executions, e.Cmd)
	}
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
	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	e.RecordingOutputs(stdout, stderr)
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
