package cmdz

import (
	"strings"

	"mby.fr/utils/collections"
	"mby.fr/utils/promise"
	"mby.fr/utils/stringz"
)

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

func Sequence() *seq {
	return &seq{}
}

type serialSeq struct {
	*seq
}

func (s *serialSeq) Serial(execs ...Executer) *serialSeq {
	s.seq.execs = append(s.seq.execs, execs...)
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

func (s *parallelSeq) Parallel(execs ...Executer) *parallelSeq {
	s.seq.execs = append(s.seq.execs, execs...)
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
	err = BlockParallel(-1, s.seq.execs...)
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func (s *parallelSeq) AsyncRun() *execPromise {
	p := promise.New(func(resolve func(int), reject func(error)) {
		rc, err := s.BlockRun()
		if err != nil {
			reject(err)
		}
		resolve(rc)
	})
	return p
}
