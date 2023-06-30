package cmdz

import (
	"strings"

	"mby.fr/utils/collections"
	"mby.fr/utils/promise"
	"mby.fr/utils/stringz"
)

func Sequence() *seq {
	return &seq{}
}

/*
func Serial() *serialSeq {
	return Sequence().Serial()
}

func Parallel() *parallelSeq {
	return Sequence().Parallel()
}
*/

type seq struct {
	// TODO describe a sequence of // and serial Exec to execute
	execs    []Executer
	failFast bool
	status   int
}

func (s *seq) Serial(execs ...Executer) *serialSeq {
	s.execs = execs
	return &serialSeq{s}
}

func (s *seq) Parallel(execs ...Executer) *parallelSeq {
	s.execs = execs
	return &parallelSeq{s}
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

func (s *serialSeq) Serial(execs ...Executer) *serialSeq {
	s.seq.execs = append(s.seq.execs, execs...)
	return s
}

func (s *serialSeq) Parallel(execs ...Executer) *parallelSeq {
	p := Sequence().Parallel(execs...)
	s.seq.execs = append(s.seq.execs, p)
	return p
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
}

func (p *parallelSeq) Parallel(execs ...Executer) *parallelSeq {
	p.seq.execs = append(p.seq.execs, execs...)
	return p
}

func (p *parallelSeq) Serial(execs ...Executer) *serialSeq {
	s := Sequence().Serial(execs...)
	p.seq.execs = append(p.seq.execs, s)
	return s
}

func (p parallelSeq) String() string {
	return stringz.JoinStringers(p.seq.execs, "\n")
}

func (p parallelSeq) ReportError() string {
	errors := collections.Map[Executer, string](p.seq.execs, func(e Executer) string {
		return e.ReportError()
	})
	return strings.Join(errors, "\n")
}

func (p *parallelSeq) BlockRun() (rc int, err error) {
	p.reset()
	err = BlockParallel(-1, p.seq.execs...)
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func (p *parallelSeq) AsyncRun() *execPromise {
	return promise.New(func(resolve func(int), reject func(error)) {
		rc, err := p.BlockRun()
		if err != nil {
			reject(err)
		}
		resolve(rc)
	})
}
