package cmdz

import (
	"io"
	"strings"

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

	// TODO describe a sequence of // and serial Exec to execute
	execs          []Executer
	failFast       bool
	fallbackConfig *config

	status int
}

func (s *seq) StdoutRecord() string {
	stdouts := collections.Map[Executer, string](s.execs, func(e Executer) string {
		return e.StdoutRecord()
	})
	return strings.Join(stdouts, "")
}

func (s *seq) StderrRecord() string {
	stderrs := collections.Map[Executer, string](s.execs, func(e Executer) string {
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

func (s *serialSeq) FailFast(enabled bool) *serialSeq {
	s.failFast = enabled
	return s
}

func (s *serialSeq) Retries(count, delayInMs int) *serialSeq {
	s.config.retries = count
	s.config.retryDelayInMs = delayInMs
	return s
}

func (s *serialSeq) Timeout(delayInMs int) *serialSeq {
	s.config.timeout = delayInMs
	return s
}

func (s *serialSeq) Outputs(stdout, stderr io.Writer) *serialSeq {
	s.config.stdout = stdout
	s.config.stderr = stderr
	return s
}

func (s *serialSeq) ErrorOnFailure(enable bool) *serialSeq {
	s.config.errorOnFailure = enable
	return s
}

func (s *serialSeq) FailOnError() Executer {
	return s.ErrorOnFailure(true)
}

func (s *serialSeq) Add(execs ...Executer) *serialSeq {
	s.execs = append(s.execs, execs...)
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

type parallelSeq struct {
	*seq

	forkCount int
}

func (s *parallelSeq) FailFast(enabled bool) *parallelSeq {
	s.failFast = enabled
	return s
}

func (s *parallelSeq) Retries(count, delayInMs int) *parallelSeq {
	s.config.retries = count
	s.config.retryDelayInMs = delayInMs
	return s
}

func (s *parallelSeq) Timeout(delayInMs int) *parallelSeq {
	s.config.timeout = delayInMs
	return s
}

func (s *parallelSeq) Outputs(stdout, stderr io.Writer) *parallelSeq {
	s.config.stdout = stdout
	s.config.stderr = stderr
	return s
}

func (s *parallelSeq) ErrorOnFailure(enable bool) *parallelSeq {
	s.config.errorOnFailure = enable
	return s
}

func (s *parallelSeq) FailOnError() Executer {
	return s.ErrorOnFailure(true)
}

func (s *parallelSeq) Fork(count int) *parallelSeq {
	s.forkCount = count
	return s
}

func (s *parallelSeq) Add(execs ...Executer) *parallelSeq {
	s.execs = append(s.execs, execs...)
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

func Serial(execs ...Executer) *serialSeq {
	s := &serialSeq{seq: &seq{config: config{}}}
	s.Add(execs...)
	return s
}

func Parallel(execs ...Executer) *parallelSeq {
	s := &parallelSeq{seq: &seq{config: config{}}}
	s.Add(execs...)
	return s
}
