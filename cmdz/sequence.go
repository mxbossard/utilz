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
Executer(binary, ...args)
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
	.Add(Executer(...)...)
	.Add(
		Executer(...)...,
		Executer(...)...
	)
	.Add(
		Executer(...)...,
		Serial(
			Executer(...)...,
			Executer(...)...
		)
	)

# execution serie
- config: timeout, retries?
Serial()
	.Retries(count)
	.Timeout(deadline)
	.Outputs(stdout, stderr)
	.Context(ctx)
	.Add(Executer(...)...)
	.Add(
		Executer(...)...,
		Executer(...)...
	)
	.Add(
		Executer(...)...,
		Parallel(
			Executer(...)...,
			Executer(...)...
		)
	)
*/

type seq struct {
	config

	// TODO describe a sequence of // and serial Exec to execute
	execs    []Executer
	failFast bool
	status   int
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

type serialSeq struct {
	*seq
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
	s.reset()
	err = BlockSerial(s.seq.execs...)
	if err != nil {
		return 1, err
	}
	return 0, nil
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
	s.reset()
	err = BlockParallel(-1, s.seq.execs...)
	if err != nil {
		return 1, err
	}
	return 0, nil
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
