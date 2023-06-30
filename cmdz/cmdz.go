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

type execPromise = promise.Promise[int]
type execsPromise = promise.Promise[[]int]

type cmdz struct {
	*exec.Cmd

	clone exec.Cmd

	StdoutRecord inout.RecordingWriter
	StderrRecord inout.RecordingWriter

	Retries        int
	RetryDelayInMs int64

	ResultsCodes []int
	// FIXME: replace ResultsCodes by Executions
	Executions []*exec.Cmd
}

func (e *cmdz) AddEnv(key, value string) {
	entry := fmt.Sprintf("%s=%s", key, value)
	e.Env = append(e.Env, entry)
}

func (e *cmdz) AddArgs(args ...string) {
	e.Args = append(e.Args, args...)
}

func (e *cmdz) RecordingOutputs(stdout, stderr io.Writer) {
	if stdout != nil {
		e.StdoutRecord.Nested = stdout
		e.Stdout = &e.StdoutRecord
	}
	if stderr != nil {
		e.StderrRecord.Nested = stderr
		e.Stderr = &e.StderrRecord
	}
}

func (e *cmdz) checkpoint() {
	e.clone = *e.Cmd
}

func (e *cmdz) rollback() {
	// Clone checkpoint
	newClone := e.clone
	e.Cmd = &newClone
}

func (e *cmdz) reset() {
	e.StdoutRecord.Reset()
	e.StderrRecord.Reset()
	e.ResultsCodes = nil
	e.Executions = nil
	e.rollback()
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

func (e cmdz) String() (t string) {
	t = strings.Join(e.Args, " ")
	return
}

func (e cmdz) ReportError() string {
	execCmdSummary := e.String()
	attempts := len(e.ResultsCodes)
	status := e.ResultsCodes[attempts-1]
	stderr := e.StderrRecord.String()
	errorMessage := fmt.Sprintf("Exec failed after %d attempt(s): [%s] !\nRC=%d ERR> %s", attempts, execCmdSummary, status, strings.TrimSpace(stderr))
	return errorMessage
}

func (e *cmdz) BlockRun() (rc int, err error) {
	e.reset()
	rc = -1
	for i := 0; i <= e.Retries && rc != 0; i++ {

		if i > 0 {
			// Wait between retries
			time.Sleep(time.Duration(e.RetryDelayInMs) * time.Millisecond)
		}
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
		e.rollback()
	}
	return rc, nil
}

func (e *cmdz) AsyncRun() *execPromise {
	p := promise.New(func(resolve func(int), reject func(error)) {
		rc, err := e.BlockRun()
		if err != nil {
			reject(err)
		}
		resolve(rc)
	})
	return p
}

func Execution(binary string, args ...string) *cmdz {
	cmd := exec.Command(binary, args...)
	e := cmdz{Cmd: cmd, RetryDelayInMs: 100}
	stdout := &strings.Builder{}
	stderr := &strings.Builder{}
	e.RecordingOutputs(stdout, stderr)
	e.checkpoint()
	return &e
}

func ExecutionOutputs(stdout, stderr io.Writer, binary string, args ...string) *cmdz {
	e := Execution(binary, args...)
	e.RecordingOutputs(stdout, stderr)
	return e
}

func ExecutionContext(ctx context.Context, binary string, args ...string) *cmdz {
	e := Execution(binary, args...)
	cmd := exec.CommandContext(ctx, binary, args...)
	e.Cmd = cmd
	e.checkpoint()
	return e
}
