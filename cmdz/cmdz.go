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

type Executer interface {
	reset()

	//Retries(int, int) Executer
	//Timeout(int) Executer
	//Outputs(io.Writer, io.Writer) Executer
	//Context(context.Context) Executer

	String() string
	ReportError() string
	BlockRun() (int, error)
	AsyncRun() *execPromise
	StdoutRecord() string
	StderrRecord() string
}

type config struct {
	retries        int
	retryDelayInMs int
	timeout        int
	stdout         io.Writer
	stderr         io.Writer
}

type cmdz struct {
	*exec.Cmd
	config

	clone exec.Cmd

	stdoutRecord inout.RecordingWriter
	stderrRecord inout.RecordingWriter

	ResultsCodes []int
	// FIXME: replace ResultsCodes by Executions
	Executions []*exec.Cmd
}

func (e *cmdz) Retries(count, delayInMs int) *cmdz {
	e.config.retries = count
	e.config.retryDelayInMs = delayInMs
	return e
}

func (e *cmdz) Timeout(delayInMs int) *cmdz {
	e.config.timeout = delayInMs
	return e
}

func (e *cmdz) Outputs(stdout, stderr io.Writer) *cmdz {
	e.stdout = stdout
	e.stderr = stderr
	e.recordingOutputs(stdout, stderr)
	return e
}

func (e *cmdz) recordingOutputs(stdout, stderr io.Writer) {
	if stdout == nil {
		stdout = &strings.Builder{}
	}
	if stderr == nil {
		stderr = &strings.Builder{}
	}

	e.stdoutRecord.Nested = stdout
	e.Stdout = &e.stdoutRecord
	e.stderrRecord.Nested = stderr
	e.Stderr = &e.stderrRecord
}

func (e *cmdz) AddEnv(key, value string) {
	entry := fmt.Sprintf("%s=%s", key, value)
	e.Env = append(e.Env, entry)
}

func (e *cmdz) AddArgs(args ...string) {
	e.Args = append(e.Args, args...)
}

func (e *cmdz) StdoutRecord() string {
	return e.stdoutRecord.String()
}

func (e *cmdz) StderrRecord() string {
	return e.stderrRecord.String()
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
	e.stdoutRecord.Reset()
	e.stderrRecord.Reset()
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
	stderr := e.stderrRecord.String()
	errorMessage := fmt.Sprintf("Exec failed after %d attempt(s): [%s] !\nRC=%d ERR> %s", attempts, execCmdSummary, status, strings.TrimSpace(stderr))
	return errorMessage
}

func (e *cmdz) BlockRun() (rc int, err error) {
	e.reset()
	rc = -1
	for i := 0; i <= e.config.retries && rc != 0; i++ {

		if i > 0 {
			// Wait between retries
			time.Sleep(time.Duration(e.config.retryDelayInMs) * time.Millisecond)
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

func Cmd(binary string, args ...string) *cmdz {
	cmd := exec.Command(binary, args...)
	e := cmdz{Cmd: cmd}
	e.recordingOutputs(cmd.Stdout, cmd.Stderr)
	e.checkpoint()
	return &e
}

func CmdCtx(ctx context.Context, binary string, args ...string) *cmdz {
	cmd := exec.CommandContext(ctx, binary, args...)
	e := cmdz{Cmd: cmd}
	e.recordingOutputs(cmd.Stdout, cmd.Stderr)
	e.checkpoint()
	return &e
}
