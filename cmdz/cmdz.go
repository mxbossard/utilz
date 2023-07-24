package cmdz

import (
	//"bufio"

	"bytes"
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

type config struct {
	retries        int
	retryDelayInMs int
	timeout        int
	stdin          io.Reader
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

	stdinRecord    inout.RecordingReader
	stdoutRecord   inout.RecordingWriter
	stderrRecord   inout.RecordingWriter

	ResultsCodes   []int
	// FIXME: replace ResultsCodes by Executions
	Executions     []*exec.Cmd

	pipedInput	   bool
	pipedOutput	   bool
	pipeFail	   bool
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

func (e *cmdz) Input(stdin io.Reader) *cmdz {
	if !e.pipedInput {
		e.config.stdin = stdin
		e.recordingInput(stdin)
	}
	return e
}

func (e *cmdz) Outputs(stdout, stderr io.Writer) *cmdz {
	if e.pipedOutput {
		stdout = e.config.stdout
	}
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

func (e *cmdz) recordingInput(stdin io.Reader) {
	if stdin == nil {
		stdin = strings.NewReader("")
	}

	e.stdinRecord.Nested = stdin
	e.Cmd.Stdin = &e.stdinRecord
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

func (e *cmdz) StdinRecord() string {
	return e.stdinRecord.String()
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
	e.stdinRecord.Reset()
	e.stdoutRecord.Reset()
	e.stderrRecord.Reset()
	e.ResultsCodes = nil
	e.Executions = nil
	e.rollback()
}

func (e *cmdz) fallback(cfg *config) {
	e.fallbackConfig = cfg
}

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
	e.recordingInput(config.stdin)
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

func (e *cmdz) Pipe(c *cmdz) (*cmdz) {
	e.pipedOutput = true
	c.pipedInput = true
	b := bytes.Buffer{}
	c.config.stdin = &b
	c.stdinRecord.Nested = &b
	c.config.stdout = &b
	e.stdoutRecord.Nested = &b
	return e
}

func (e *cmdz) PipeFail(c *cmdz) (*cmdz) {
	e.pipeFail = true
	return e
}

func (e *cmdz) Output() (string, error) {
	rc, err := e.ErrorOnFailure(true).BlockRun()
	if err != nil {
		return "", err
	}
	if rc > 0 {
		err = failure{rc, e}
		return "", err
	}
	return e.StdoutRecord(), nil
}

func (e *cmdz) CombinedOutput() (string, error) {
	if e.config.stdout == nil {
		e.config.stdout = &strings.Builder{}
	}
	combinedWriter := inout.RecordingWriter{Nested: e.config.stdout}
	rc, err := e.ErrorOnFailure(true).Outputs(&combinedWriter, &combinedWriter).BlockRun()
	if err != nil {
		return "", err
	}
	if rc > 0 {
		err = failure{rc, e}
		return "", err
	}
	return combinedWriter.String(), nil
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
