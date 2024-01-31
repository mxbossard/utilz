package cmdz

import (
	"io"
	"log"
	"strings"
	"time"

	"mby.fr/utils/collections"
	"mby.fr/utils/promise"
	"mby.fr/utils/stringz"
)

// IDEAS: Sequence() => Config() Parallel(config) Serial(config) AddExecuter(config)
/*

# 1 seul execution
- config: retries, outputs, context, timeout
Cmd(binary, ...args)
	.Retries(count)
	.Timeout(deadline)
	.Outputs(stdout, stderr)
	.Context(ctx)

# execution //
- config: fork, timeout
Parallel()
	.Retries(count)
	.Timeout(deadline)
	.Outputs(stdout, stderr)
	.Context(ctx)
	.Fork(fork)
	.Add(Cmd(...)...)
	.Add(
		Cmd(...)...,
		Cmd(...)...
	)
	.Add(
		Cmd(...)...,
		Cmd(
			Cmd(...)...,
			Cmd(...)...
		)
	)

# execution serie
- config: timeout, retries?
Serial()
	.Retries(count)
	.Timeout(deadline)
	.Outputs(stdout, stderr)
	.Context(ctx)
	.Add(Cmd(...)...)
	.Add(
		Cmd(...)...,
		Cmd(...)...
	)
	.Add(
		Cmd(...)...,
		Parallel(
			Cmd(...)...,
			Cmd(...)...
		)
	)
*/

type seq struct {
	config

	inners []Executer
	outers []Executer

	// TODO describe a sequence of // and serial Exec to execute
	execs          []Executer
	failFast       bool
	fallbackConfig *config

	status int
}

func (e *seq) getConfig() config {
	return e.config
}

func (e *seq) setStdin(stdin io.Reader) {
	if e.pipedInput {
		log.Fatal("Input is piped cannot change it !")
	}
	e.config.stdin = stdin
	for _, i := range e.execs {
		i.SetInput(stdin)
	}
}

func (e *seq) setStdout(stdout io.Writer) {
	if e.pipedOutput {
		log.Fatal("Output is piped cannot change it !")
	}
	e.config.stdout = stdout
	for _, o := range e.outers {
		o.SetStdout(stdout)
	}
}

func (e *seq) setStderr(stderr io.Writer) {
	e.config.stderr = stderr
	for _, o := range e.outers {
		o.SetStderr(stderr)
	}
}

// ----- InOuter methods -----
func (e *seq) Stdin() io.Reader {
	return e.config.stdin
}
func (e *seq) Stdout() io.Writer {
	return e.config.stdout
}
func (e *seq) Stderr() io.Writer {
	return e.config.stdout
}

// ----- Recorder methods -----
func (s *seq) StdinRecord() string {
	stdins := collections.Map[Executer, string](s.inners, func(e Executer) string {
		return e.StdinRecord()
	})
	return strings.Join(stdins, "")
}

func (s *seq) StdoutRecord() string {
	stdouts := collections.Map[Executer, string](s.outers, func(e Executer) string {
		return e.StdoutRecord()
	})
	return strings.Join(stdouts, "")
}

func (s *seq) StderrRecord() string {
	stderrs := collections.Map[Executer, string](s.outers, func(e Executer) string {
		return e.StderrRecord()
	})
	return strings.Join(stderrs, "")
}

func (s *seq) reset() {
	for _, e := range s.execs {
		e.reset()
	}
}

func (s *seq) fallback(cfg *config) {
	s.fallbackConfig = cfg
}

type serialSeq struct {
	*seq
}

func (e *serialSeq) SetInput(stdin io.Reader) Executer {
	e.seq.setStdin(stdin)
	return e
}

func (e *serialSeq) SetStdout(stdout io.Writer) Executer {
	e.seq.setStdout(stdout)
	return e
}

func (e *serialSeq) SetStderr(stderr io.Writer) Executer {
	e.seq.setStderr(stderr)
	return e
}

func (e *serialSeq) SetOutputs(stdout, stderr io.Writer) Executer {
	e.SetStdout(stdout)
	e.SetStderr(stderr)
	return e
}

func (s *serialSeq) FailFast(enabled bool) *serialSeq {
	s.failFast = enabled
	return s
}

func (s *serialSeq) Retries(count, delayInMs int) Executer {
	s.config.retries = count
	s.config.retryDelayInMs = delayInMs
	return s
}

func (s *serialSeq) Timeout(delayInMs int) Executer {
	s.config.timeout = delayInMs
	return s
}

func (s *serialSeq) CombinedOutputs() Executer {
	for _, e := range s.outers {
		e.CombinedOutputs()
	}
	return s
}

func (s *serialSeq) ErrorOnFailure(enable bool) Executer {
	s.config.errorOnFailure = enable
	return s
}

func (s *serialSeq) Add(execs ...Executer) *serialSeq {
	s.execs = append(s.execs, execs...)
	if len(s.execs) > 0 {
		s.inners = []Executer{s.execs[0]}
	}
	s.outers = s.execs
	return s
}

func (s serialSeq) String() string {
	return stringz.JoinStringers(s.seq.execs, "\n")
}

func (s serialSeq) ReportError() string {
	errors := collections.Map[Executer, string](s.seq.execs, func(e Executer) string {
		return e.ReportError()
	})
	return strings.Join(errors, "\n")
}

func (s *serialSeq) BlockRun() (rc int, err error) {
	mergedConfig := mergeConfigs(&s.config, s.fallbackConfig)
	for _, exec := range s.seq.execs {
		exec.fallback(mergedConfig)
	}
	s.reset()
	if len(s.seq.execs) > 0 {
		rc, err = blockSerial(s.failFast, s.seq.execs...)
	}
	return
}

func (s *serialSeq) AsyncRun() *execPromise {
	p := promise.New(func(resolve func(int), reject func(error)) {
		rc, err := s.BlockRun()
		if err != nil {
			reject(err)
		}
		resolve(rc)
	})
	return p
}

func (e *serialSeq) ResultCodes() (codes []int) {
	for _, exec := range e.execs {
		codes = append(codes, exec.ResultCodes()...)
	}
	return
}

func (e *serialSeq) ExitCode() int {
	codes := e.ResultCodes()
	return codes[len(codes)-1]
}

func (e *serialSeq) StartTimes() (times []time.Time) {
	for _, exec := range e.execs {
		times = append(times, exec.StartTime())
	}
	return
}

func (e *serialSeq) StartTime() time.Time {
	return e.StartTimes()[0]
}

func (e *serialSeq) Durations() (durations []time.Duration) {
	for _, exec := range e.execs {
		durations = append(durations, exec.Duration())
	}
	return
}

func (e *serialSeq) Duration() time.Duration {
	startTimes := e.StartTimes()
	lastStart := startTimes[len(startTimes)-1]
	lastDuration := e.Durations()[len(startTimes)-1]
	totalDuration := lastStart.Add(lastDuration).Sub(startTimes[0])
	return totalDuration
}

type andSeq struct {
	serialSeq
}

func (s andSeq) String() string {
	return stringz.JoinStringers(s.seq.execs, " && ")
}

type orSeq struct {
	*seq
}

func (e *orSeq) SetInput(stdin io.Reader) Executer {
	e.seq.setStdin(stdin)
	return e
}

func (e *orSeq) SetStdout(stdout io.Writer) Executer {
	e.seq.setStdout(stdout)
	return e
}

func (e *orSeq) SetStderr(stderr io.Writer) Executer {
	e.seq.setStderr(stderr)
	return e
}

func (e *orSeq) SetOutputs(stdout, stderr io.Writer) Executer {
	if e.pipedOutput {
		log.Fatal("Output is piped cannot change it !")
	}
	e.config.stdout = stdout
	e.config.stderr = stderr
	for _, o := range e.outers {
		o.SetOutputs(stdout, stderr)
	}
	return e
}

func (s orSeq) String() string {
	return stringz.JoinStringers(s.seq.execs, " || ")
}

func (s *orSeq) Retries(count, delayInMs int) Executer {
	s.config.retries = count
	s.config.retryDelayInMs = delayInMs
	return s
}

func (s *orSeq) Timeout(delayInMs int) Executer {
	s.config.timeout = delayInMs
	return s
}

func (s *orSeq) CombinedOutputs() Executer {
	for _, e := range s.execs {
		e.CombinedOutputs()
	}
	return s
}

func (s *orSeq) ErrorOnFailure(enable bool) Executer {
	s.config.errorOnFailure = enable
	return s
}

func (s *orSeq) Add(execs ...Executer) *orSeq {
	s.execs = append(s.execs, execs...)
	s.inners = s.execs
	s.outers = s.execs
	return s
}

func (s orSeq) ReportError() string {
	errors := collections.Map[Executer, string](s.seq.execs, func(e Executer) string {
		return e.ReportError()
	})
	return strings.Join(errors, "\n")
}

func (s *orSeq) BlockRun() (rc int, err error) {
	mergedConfig := mergeConfigs(&s.config, s.fallbackConfig)
	for _, exec := range s.seq.execs {
		exec.fallback(mergedConfig)
	}
	s.reset()
	if len(s.seq.execs) > 0 {
		rc, err = blockOr(s.seq.execs...)
	}
	return
}

func (s *orSeq) AsyncRun() *execPromise {
	p := promise.New(func(resolve func(int), reject func(error)) {
		rc, err := s.BlockRun()
		if err != nil {
			reject(err)
		}
		resolve(rc)
	})
	return p
}

func (e *orSeq) ResultCodes() (codes []int) {
	for _, exec := range e.execs {
		codes = append(codes, exec.ResultCodes()...)
	}
	return
}

func (e *orSeq) ExitCode() int {
	codes := e.ResultCodes()
	return codes[len(codes)-1]
}

func (e *orSeq) StartTimes() (times []time.Time) {
	for _, exec := range e.execs {
		times = append(times, exec.StartTime())
	}
	return
}

func (e *orSeq) StartTime() time.Time {
	return e.StartTimes()[0]
}

func (e *orSeq) Durations() (durations []time.Duration) {
	for _, exec := range e.execs {
		durations = append(durations, exec.Duration())
	}
	return
}

func (e *orSeq) Duration() time.Duration {
	startTimes := e.StartTimes()
	lastStart := startTimes[len(startTimes)-1]
	lastDuration := e.Durations()[len(startTimes)-1]
	totalDuration := lastStart.Add(lastDuration).Sub(startTimes[0])
	return totalDuration
}

type parallelSeq struct {
	*seq

	forkCount int
}

func (e *parallelSeq) SetInput(stdin io.Reader) Executer {
	e.seq.setStdin(stdin)
	return e
}

func (e *parallelSeq) SetStdout(stdout io.Writer) Executer {
	e.seq.setStdout(stdout)
	return e
}

func (e *parallelSeq) SetStderr(stderr io.Writer) Executer {
	e.seq.setStderr(stderr)
	return e
}

func (e *parallelSeq) SetOutputs(stdout, stderr io.Writer) Executer {
	if e.pipedOutput {
		log.Fatal("Output is piped cannot change it !")
	}
	e.config.stdout = stdout
	e.config.stderr = stderr
	for _, o := range e.execs {
		o.SetOutputs(stdout, stderr)
	}
	return e
}

// ----- Configurer methods -----

func (s *parallelSeq) ErrorOnFailure(enable bool) Executer {
	s.config.errorOnFailure = enable
	return s
}

func (s *parallelSeq) Retries(count, delayInMs int) Executer {
	s.config.retries = count
	s.config.retryDelayInMs = delayInMs
	return s
}

func (s *parallelSeq) Timeout(delayInMs int) Executer {
	s.config.timeout = delayInMs
	return s
}

func (s *parallelSeq) FailFast(enabled bool) *parallelSeq {
	s.failFast = enabled
	return s
}

func (s *parallelSeq) CombinedOutputs() Executer {
	for _, e := range s.execs {
		e.CombinedOutputs()
	}
	return s
}

func (s *parallelSeq) Fork(count int) *parallelSeq {
	s.forkCount = count
	return s
}

func (s *parallelSeq) Add(execs ...Executer) *parallelSeq {
	s.execs = append(s.execs, execs...)
	s.inners = s.execs
	s.outers = s.execs
	return s
}

func (s parallelSeq) String() string {
	return stringz.JoinStringers(s.seq.execs, "\n")
}

func (s parallelSeq) ReportError() string {
	errors := collections.Map[Executer, string](s.seq.execs, func(e Executer) string {
		return e.ReportError()
	})
	return strings.Join(errors, "\n")
}

func (s *parallelSeq) BlockRun() (rc int, err error) {
	for _, exec := range s.seq.execs {
		exec.fallback(&s.config)
	}
	s.reset()
	if len(s.seq.execs) > 0 {
		rc, err = blockParallel(s.failFast, s.forkCount, s.seq.execs...)
	}
	return
}

func (s *parallelSeq) AsyncRun() *execPromise {
	return promise.New(func(resolve func(int), reject func(error)) {
		rc, err := s.BlockRun()
		if err != nil {
			reject(err)
		}
		resolve(rc)
	})
}

func (e *parallelSeq) ResultCodes() (codes []int) {
	for _, exec := range e.execs {
		codes = append(codes, exec.ResultCodes()...)
	}
	return
}

func (e *parallelSeq) ExitCode() int {
	codes := e.ResultCodes()
	return codes[len(codes)-1]
}

func (e *parallelSeq) StartTimes() (times []time.Time) {
	for _, exec := range e.execs {
		times = append(times, exec.StartTime())
	}
	return
}

func (e *parallelSeq) StartTime() time.Time {
	return e.StartTimes()[0]
}

func (e *parallelSeq) Durations() (durations []time.Duration) {
	for _, exec := range e.execs {
		durations = append(durations, exec.Duration())
	}
	return
}

func (e *parallelSeq) Duration() time.Duration {
	var d time.Duration
	for _, exec := range e.execs {
		if exec.Duration().Nanoseconds() > d.Nanoseconds() {
			d = exec.Duration()
		}
	}
	return d
}
