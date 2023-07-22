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
	"mby.fr/utils/stringz"
)

type execPromise = promise.Promise[int]
type execsPromise = promise.Promise[[]int]

type failure struct {
	Rc  int
	Cmd *cmdz
}

func (f failure) Error() string {
	stderrSummary := stringz.SummaryRatio(f.Cmd.StderrRecord(), 128, 0.2)
	return fmt.Sprintf("Failing with ResultCode: %d executing: [%s] ! stderr: %s", f.Rc, f.Cmd.String(), stderrSummary)
}

type Executer interface {
	reset()
	fallback(*config)

	String() string
	ReportError() string
	BlockRun() (int, error)
	AsyncRun() *execPromise
	StdoutRecord() string
	StderrRecord() string
	FailOnError() Executer
}

type config struct {
	retries        int
	retryDelayInMs int
	timeout        int
	stdout         io.Writer
	stderr         io.Writer
	errorOnFailure bool
}

// Merge lower priority config into higher priority
func mergeConfigs(higher, lower *config) *config {
	merged := *higher
	if lower == nil {
		return &merged
	}
	if merged.retries == 0 {
		merged.retries = lower.retries
	}
	if merged.retryDelayInMs == 0 {
		merged.retryDelayInMs = lower.retryDelayInMs
	}
	if merged.timeout == 0 {
		merged.timeout = lower.timeout
	}
	if merged.stdout == nil {
		merged.stdout = lower.stdout
	}
	if merged.stderr == nil {
		merged.stderr = lower.stderr
	}
	if !merged.errorOnFailure {
		merged.errorOnFailure = lower.errorOnFailure
	}
	return &merged
}

type cmdz struct {
	*exec.Cmd
	config

	cmdCheckpoint  exec.Cmd
	fallbackConfig *config

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
	e.config.stdout = stdout
	e.config.stderr = stderr
	e.recordingOutputs(stdout, stderr)
	return e
}

func (e *cmdz) ErrorOnFailure(enable bool) *cmdz {
	e.config.errorOnFailure = enable
	return e
}

func (e *cmdz) FailOnError() Executer {
	return e.ErrorOnFailure(true)
}

func (e *cmdz) recordingOutputs(stdout, stderr io.Writer) {
	if stdout == nil {
		stdout = &strings.Builder{}
	}
	if stderr == nil {
		stderr = &strings.Builder{}
	}

	e.stdoutRecord.Nested = stdout
	e.Cmd.Stdout = &e.stdoutRecord
	e.stderrRecord.Nested = stderr
	e.Cmd.Stderr = &e.stderrRecord
}

func (e *cmdz) AddEnv(key, value string) *cmdz {
	entry := fmt.Sprintf("%s=%s", key, value)
	e.Cmd.Env = append(e.Env, entry)
	e.checkpoint()
	return e
}

func (e *cmdz) AddArgs(args ...string) *cmdz {
	e.Cmd.Args = append(e.Args, args...)
	e.checkpoint()
	return e
}

func (e *cmdz) StdoutRecord() string {
	return e.stdoutRecord.String()
}

func (e *cmdz) StderrRecord() string {
	return e.stderrRecord.String()
}

func (e *cmdz) checkpoint() {
	e.cmdCheckpoint = *e.Cmd
}

func (e *cmdz) rollback() {
	// Clone checkpoint
	newClone := e.cmdCheckpoint
	e.Cmd = &newClone
}

func (e *cmdz) reset() {
	e.stdoutRecord.Reset()
	e.stderrRecord.Reset()
	e.ResultsCodes = nil
	e.Executions = nil
	e.rollback()
}

func (e *cmdz) fallback(cfg *config) {
	e.fallbackConfig = cfg
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
	config := mergeConfigs(&e.config, e.fallbackConfig)
	e.recordingOutputs(config.stdout, config.stderr)
	e.checkpoint()
	rc = -1
	for i := 0; i <= config.retries && rc != 0; i++ {

		if i > 0 {
			// Wait between retries
			time.Sleep(time.Duration(config.retryDelayInMs) * time.Millisecond)
		}
		if commandMock == nil {
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
		} else {
			// Replace execution by mocking function
			rc = commandMock.Mock(e.Cmd)
		}
		e.ResultsCodes = append(e.ResultsCodes, rc)
		e.Executions = append(e.Executions, e.Cmd)
		e.rollback()
	}
	if e.errorOnFailure && rc > 0 {
		err = failure{rc, e}
		rc = -1
	}
	return
}

func (e *cmdz) AsyncRun() *execPromise {
	p := promise.New(func(resolve func(int), reject func(error)) {
		rc, err := e.BlockRun()
		if err != nil {
			reject(err)
		}
		if e.errorOnFailure && rc > 0 {
			err = failure{rc, e}
			reject(err)
		}
		resolve(rc)
	})
	return p
}

func Cmd(binary string, args ...string) *cmdz {
	cmd := exec.Command(binary, args...)
	cmd.Env = make([]string, 8)
	e := cmdz{Cmd: cmd}
	e.checkpoint()
	return &e
}

func CmdCtx(ctx context.Context, binary string, args ...string) *cmdz {
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = make([]string, 8)
	e := cmdz{Cmd: cmd}
	e.checkpoint()
	return &e
}

//var mockingCommand func(exec.Cmd) (int, io.Reader, io.Reader)
