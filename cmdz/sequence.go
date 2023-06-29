package cmdz

import (
	"mby.fr/utils/stringz"
)

type seq struct {
	// TODO describe a sequence of // and serial Exec to execute
	execs []Executer
}

func (s *seq) Serial(execs ...Executer) *serialSeq {
	s.execs = execs
	return &serialSeq{s}
}

func (s *seq) Parallel(execs ...Executer) *parallelSeq {
	s.execs = execs
	return &parallelSeq{s}
}

func Sequence() seq {
	return seq{}
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

func (s serialSeq) ReportError(exec Exec) string {

}

func (s *serialSeq) BlockRun() (rc int, err error) {

}

func (s *serialSeq) AsyncRun() *ExecPromise {

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

func (s parallelSeq) ReportError(exec Exec) string {

}

func (s *parallelSeq) BlockRun() (rc int, err error) {

}

func (s *parallelSeq) AsyncRun() *ExecPromise {

}
