package cmdz

import (
	//"bufio"

	"bytes"
	"context"
	"log"

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
	Rc       int
	reporter Reporter
}

func (f failure) Error() (msg string) {
	if f.reporter != nil {
		stderrSummary := stringz.SummaryRatio(f.reporter.ReportError(), 128, 0.2)
		msg = fmt.Sprintf("Failing with ResultCode: %d executing: [%s] ! stderr: %s", f.Rc, f.reporter.String(), stderrSummary)
	}
	return
}

type config struct {
	retries        int
	retryDelayInMs int
	timeout        time.Duration
	stdin          io.Reader
	stdout         io.Writer
	stderr         io.Writer
	combinedOuts   bool
	errorOnFailure bool

	feeder      *cmdz
	pipedInput  bool
	pipedOutput bool
	pipeFail    bool
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
	cmd    *exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc

	binary      string
	args        []string
	environ     []string
	initialized bool
	stdin       io.Reader
	stdout      io.Writer
	stderr      io.Writer

	config

	cmdCheckpoint  exec.Cmd
	fallbackConfig *config

	stdinRecord  inout.RecordingReader
	stdoutRecord inout.RecordingWriter
	stderrRecord inout.RecordingWriter

	inProcesser  inout.ProcessingReader
	outProcesser inout.ProcessingWriter
	errProcesser inout.ProcessingWriter

	exitCodes []int
	// FIXME: replace exitCodes by Executions
	executions []*exec.Cmd
	startTimes []time.Time
	durations  []time.Duration
}

func (e *cmdz) getConfig() config {
	return e.config
}

// ----- InOuter methods -----
func (e *cmdz) Stdin() io.Reader {
	return e.stdin
}
func (e *cmdz) Stdout() io.Writer {
	return e.stdout
}
func (e *cmdz) Stderr() io.Writer {
	return e.stderr
}

func (e *cmdz) SetInput(stdin io.Reader) Executer {
	if e.pipedInput {
		log.Fatal("Input is piped cannot change it !")
	}
	//e.setupStdin(stdin)
	e.stdin = stdin
	e.config.stdin = stdin
	return e
}

func (e *cmdz) SetStdout(stdout io.Writer) Executer {
	if e.pipedOutput {
		log.Fatal("Output is piped cannot change it !")
	}
	//e.setupStdout(stdout)
	e.stdout = stdout
	e.config.stdout = stdout
	return e
}

func (e *cmdz) SetStderr(stderr io.Writer) Executer {
	//e.setupStderr(stderr)
	e.stderr = stderr
	e.config.stderr = stderr
	return e
}

func (e *cmdz) SetOutputs(stdout, stderr io.Writer) Executer {
	e.SetStdout(stdout)
	e.SetStderr(stderr)
	return e
}

func (e *cmdz) initProcessers() {
	if e.inProcesser == nil {
		e.inProcesser = inout.NewProcessingStreamReader(nil)
	}
	if e.outProcesser == nil {
		e.outProcesser = inout.NewProcessingStreamWriter(nil)
	}
	if e.errProcesser == nil {
		e.errProcesser = inout.NewProcessingStreamWriter(nil)
	}
}

func (e *cmdz) setupStdin(stdin io.Reader) {
	if stdin == nil {
		if e.stdin != nil {
			stdin = e.stdin
		} else {
			stdin = bytes.NewReader(nil)
		}
	}
	// Save supplied stdin in config
	// Decorate reader: ProcessingReader => RecordingReader => stdin
	e.initProcessers()
	e.config.stdin = stdin
	e.inProcesser.Nest(stdin)
	e.stdinRecord.Nested = e.inProcesser
	e.cmd.Stdin = &e.stdinRecord
}

func (e *cmdz) setupStdout(stdout io.Writer) {
	if stdout == nil {
		if e.stdout != nil {
			stdout = e.stdout
		} else {
			stdout = &bytes.Buffer{}
		}
	}
	// Save supplied stdout in config
	// Decorate writer: ProcessingWriter => RecordingWriter => stdout
	e.initProcessers()
	e.config.stdout = stdout
	e.stdoutRecord.Nested = stdout
	e.outProcesser.Nest(&e.stdoutRecord)
	e.cmd.Stdout = e.outProcesser
	if e.config.combinedOuts {
		e.cmd.Stderr = e.outProcesser
	}
}

func (e *cmdz) setupStderr(stderr io.Writer) {
	if e.config.combinedOuts {
		return
	}
	if stderr == nil {
		if e.stderr != nil {
			stderr = e.stderr
		} else {
			stderr = &strings.Builder{}
		}
	}
	// Save supplied stderr in config
	// Decorate writer: ProcessingWriter => RecordingWriter => stderr
	e.initProcessers()
	e.config.stderr = stderr
	e.stderrRecord.Nested = stderr
	e.errProcesser.Nest(&e.stderrRecord)
	e.cmd.Stderr = e.errProcesser
}

func (e *cmdz) ProcessIn(pcrs ...IOProcesser) *cmdz {
	e.initProcessers()
	e.inProcesser.Add(pcrs...)
	return e
}

func (e *cmdz) ProcessOut(pcrs ...IOProcesser) *cmdz {
	e.initProcessers()
	e.outProcesser.Add(pcrs...)
	return e
}

func (e *cmdz) ProcessErr(pcrs ...IOProcesser) *cmdz {
	e.initProcessers()
	e.errProcesser.Add(pcrs...)
	return e
}

// ----- Configurer methods -----
func (e *cmdz) ErrorOnFailure(enable bool) Executer {
	e.config.errorOnFailure = enable
	return e
}

func (e *cmdz) Retries(count, delayInMs int) Executer {
	e.config.retries = count
	e.config.retryDelayInMs = delayInMs
	return e
}

func (e *cmdz) Timeout(duration time.Duration) Executer {
	e.config.timeout = duration
	e.ctx, e.cancel = context.WithTimeout(context.Background(), duration)
	return e
}

func (e *cmdz) CombinedOutputs() Executer {
	e.config.combinedOuts = true
	return e
}

// ----- Recorder methods -----
func (e *cmdz) StdinRecord() string {
	return e.stdinRecord.String()
}

func (e *cmdz) StdoutRecord() string {
	return e.stdoutRecord.String()
}

func (e *cmdz) StderrRecord() string {
	return e.stderrRecord.String()
}

// ----- Reporter methods -----
func (e cmdz) String() (t string) {
	t = e.binary + " " + strings.Join(e.args, " ")
	return
}

func (e cmdz) ReportError() string {
	execCmdSummary := e.String()
	attempts := len(e.exitCodes)
	status := e.exitCodes[attempts-1]
	stderr := e.stderrRecord.String()
	errorMessage := fmt.Sprintf("Exec failed after %d attempt(s): [%s] !\nRC=%d ERR> %s", attempts, execCmdSummary, status, strings.TrimSpace(stderr))
	return errorMessage
}

// ----- Runner methods -----
func (e *cmdz) reset() {
	e.stdinRecord.Reset()
	e.stdoutRecord.Reset()
	e.stderrRecord.Reset()
	e.exitCodes = nil
	e.executions = nil
	e.startTimes = nil
	e.durations = nil
	e.rollback()
}

func (e *cmdz) fallback(cfg *config) {
	e.fallbackConfig = cfg
}

func (e *cmdz) init() {
	// Init internal exec.Cmd
	if !e.initialized {
		e.cmd = exec.CommandContext(e.ctx, e.binary, e.args...)
		e.cmd.Env = e.environ
	}
	//e.setupStdin(e.stdin)
	//e.setupStdout(e.stdout)
	//e.setupStderr(e.stderr)
	e.initialized = true
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

	e.init()

	config := mergeConfigs(&e.config, e.fallbackConfig)
	e.setupStdin(config.stdin)
	e.setupStdout(config.stdout)
	e.setupStderr(config.stderr)
	e.checkpoint()

	rc = -1
	for i := 0; i <= config.retries && rc != 0; i++ {
		var startTime time.Time
		var duration time.Duration
		if i > 0 {
			// Wait between retries
			time.Sleep(time.Duration(config.retryDelayInMs) * time.Millisecond)
		}
		if commandMock == nil {
			startTime = time.Now()
			e.startTimes = append(e.startTimes, startTime)
			err = e.cmd.Start()
			if err != nil {
				return
			}
			err = e.cmd.Wait()
			duration = time.Since(startTime)
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					rc = exitErr.ProcessState.ExitCode()
					err = nil
				} else {
					return -1, err
				}
			} else {
				rc = e.cmd.ProcessState.ExitCode()
			}
		} else {
			// Replace execution by mocking function
			rc = commandMock.Mock(e.cmd)
		}
		e.exitCodes = append(e.exitCodes, rc)
		e.executions = append(e.executions, e.cmd)
		e.durations = append(e.durations, duration)
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

func (e *cmdz) ResultCodes() []int {
	return e.exitCodes
}

func (e *cmdz) ExitCode() int {
	return e.exitCodes[len(e.exitCodes)-1]
}

func (e *cmdz) StartTimes() []time.Time {
	return e.startTimes
}

func (e *cmdz) StartTime() time.Time {
	return e.startTimes[len(e.startTimes)-1]
}

func (e *cmdz) Durations() []time.Duration {
	return e.durations
}

func (e *cmdz) Duration() time.Duration {
	return e.durations[len(e.durations)-1]
}

func (e *cmdz) Executions() []*exec.Cmd {
	// FIXME: do not return pointers
	return e.executions
}

func (e *cmdz) Execution() exec.Cmd {
	return *e.executions[len(e.executions)-1]
}

// ----- Piper methods -----
func (e *cmdz) Pipe(c *cmdz) *cmdz {
	c.feeder = e
	return c
}

func (e *cmdz) PipeFail(c *cmdz) *cmdz {
	e.pipeFail = true
	return e.Pipe(c)
}

func (e *cmdz) AddEnv(key, value string) *cmdz {
	entry := fmt.Sprintf("%s=%s", key, value)
	e.environ = append(e.environ, entry)
	e.checkpoint()
	return e
}

func (e *cmdz) AddEnviron(environ []string) *cmdz {
	e.environ = append(e.environ, environ...)
	e.checkpoint()
	return e
}

func (e *cmdz) AddArgs(args ...string) *cmdz {
	e.args = append(e.args, args...)
	e.checkpoint()
	return e
}

func (e *cmdz) checkpoint() {
	if e.cmd != nil {
		e.cmdCheckpoint = *e.cmd
	}
}

func (e *cmdz) rollback() {
	// Clone checkpoint
	newClone := e.cmdCheckpoint
	e.cmd = &newClone
}
