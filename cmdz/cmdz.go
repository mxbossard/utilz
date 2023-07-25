package cmdz

import (
	//"bufio"

	"bytes"
	"context"

	//"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"mby.fr/utils/inout"
	"mby.fr/utils/promise"
	"mby.fr/utils/stringz"
)

type failure struct {
	Rc   int
	Exec Executer
}

func (f failure) Error() (msg string) {
	if f.Exec != nil {
		stderrSummary := stringz.SummaryRatio(f.Exec.StderrRecord(), 128, 0.2)
		msg = fmt.Sprintf("Failing with ResultCode: %d executing: [%s] ! stderr: %s", f.Rc, f.Exec.String(), stderrSummary)
	}
	return
}

type config struct {
	retries        int
	retryDelayInMs int
	timeout        int
	stdin          io.Reader
	stdout         io.Writer
	stderr         io.Writer
	combinedOuts   bool
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

	stdinRecord  inout.RecordingReader
	stdoutRecord inout.RecordingWriter
	stderrRecord inout.RecordingWriter

	resultsCodes []int
	// FIXME: replace ResultsCodes by Executions
	Executions []*exec.Cmd

	feeder      *cmdz
	pipedInput  bool
	pipedOutput bool
	pipeFail    bool

	processers []OutProcesser
}

func (e *cmdz) Retries(count, delayInMs int) Executer {
	e.config.retries = count
	e.config.retryDelayInMs = delayInMs
	return e
}

func (e *cmdz) Timeout(delayInMs int) Executer {
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

func (e *cmdz) CombineOutputs() Executer {
	e.config.combinedOuts = true
	return e
}

func (e *cmdz) ErrorOnFailure(enable bool) Executer {
	e.config.errorOnFailure = enable
	return e
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

	e.stdoutRecord.Nested = stdout
	e.Cmd.Stdout = &e.stdoutRecord

	if e.config.combinedOuts {
		e.stderrRecord.Nested = stdout
		e.Cmd.Stderr = &e.stdoutRecord
	} else {
		if stderr == nil {
			stderr = &strings.Builder{}
		}
		e.stderrRecord.Nested = stderr
		e.Cmd.Stderr = &e.stderrRecord
	}
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
	e.resultsCodes = nil
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
	attempts := len(e.resultsCodes)
	status := e.resultsCodes[attempts-1]
	stderr := e.stderrRecord.String()
	errorMessage := fmt.Sprintf("Exec failed after %d attempt(s): [%s] !\nRC=%d ERR> %s", attempts, execCmdSummary, status, strings.TrimSpace(stderr))
	return errorMessage
}

func (e *cmdz) Pipe(c *cmdz) *cmdz {
	c.feeder = e
	return c
}

func (e *cmdz) PipeFail(c *cmdz) *cmdz {
	e.pipeFail = true
	return e.Pipe(c)
}

func (e *cmdz) BlockRun() (rc int, err error) {
	f := e.feeder
	var originalStdout io.Writer
	var originalStdin io.Reader
	if f != nil {
		originalStdout = f.config.stdout
		originalStdin = f.config.stdin
		b := bytes.Buffer{}

		// Replace configured stdin / stdout temporarilly
		f.pipedOutput = true
		f.config.stdout = &b

		e.pipedInput = true
		e.config.stdin = &b

		frc, ferr := e.feeder.BlockRun()
		if _, ok := ferr.(failure); ferr != nil && !ok {
			// If feeder error is not a failure{} return it immediately
			return frc, ferr
		}

		if e.feeder.pipeFail && frc > 0 {
			// if pipefail enabled and rc > 0 fail shortly
			return frc, ferr
		}
	}

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
		e.resultsCodes = append(e.resultsCodes, rc)
		e.Executions = append(e.Executions, e.Cmd)
		e.rollback()
	}
	if e.errorOnFailure && rc > 0 {
		err = failure{rc, e}
		rc = -1
	}
	if f != nil {
		f.config.stdout = originalStdout
		e.config.stdin = originalStdin
	}
	return
}

func (e *cmdz) ResultCodes() []int {
	return e.resultsCodes
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
