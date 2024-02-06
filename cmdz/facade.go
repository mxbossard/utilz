package cmdz

import (
	"context"
	"strings"
)

// ----- Commands -----
func Cmd(binaryAndArgs ...string) *cmdz {
	//cmd := exec.Command(binary, args...)
	//cmd.Env = make([]string, 8)
	var binary string
	var args []string
	if len(binaryAndArgs) > 0 {
		binary = binaryAndArgs[0]
	}
	if len(binaryAndArgs) > 1 {
		args = binaryAndArgs[1:]
	}
	e := cmdz{binary: binary, args: args, ctx: context.Background()}
	e.checkpoint()
	return &e
}

func CmdCtx(ctx context.Context, binaryAndArgs ...string) *cmdz {
	//cmd := exec.CommandContext(ctx, binary, args...)
	//cmd.Env = make([]string, 8)
	e := Cmd(binaryAndArgs...)
	e.ctx = ctx
	e.checkpoint()
	return e
}

func Sh(cmd ...string) *cmdz {
	return Cmd("sh", "-c", strings.Join(cmd, " "))
}

// ----- Executions -----
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

func And(execs ...Executer) *andSeq {
	s := &andSeq{serialSeq: *Serial(execs...)}
	s.Add(execs...)
	return s
}

func Or(execs ...Executer) *orSeq {
	s := &orSeq{seq: &seq{config: config{}}}
	s.Add(execs...)
	return s
}

// ----- Outputters -----
func Outputted(e Executer) Outputer {
	return &basicOutput{Executer: e}
}

func OutputtedCmd(binaryAndArgs ...string) Outputer {
	return Outputted(Cmd(binaryAndArgs...))
}

// ----- Formatters -----
func Formatted[O any](e Executer, f func(int, []byte, []byte) (O, error)) *basicFormat[O] {
	return &basicFormat[O]{Executer: e, outFormatter: f}
}

func FormattedCmd[O any](f func(int, []byte, []byte) (O, error), binaryAndArgs ...string) *basicFormat[O] {
	c := Cmd(binaryAndArgs...)
	return Formatted(c, f)
}

// ----- Pipers -----
func Pipable(i Executer) Piper {
	return &basicPipe{feeder: i}
}

func Pipe(i, o Executer) Piper {
	return &basicPipe{Executer: o, feeder: i}
}
