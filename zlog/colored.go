package zlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"mby.fr/utils/ansi"
)

func levelAnsiColor(l slog.Level) (string, string) {
	switch {
	case l < LevelPerf:
		return string(ansi.BoldHiWhite), string(ansi.HiWhite)
	case l < LevelDebug:
		return string(ansi.BoldHiGreen), string(ansi.HiGreen)
	case l < LevelInfo:
		return string(ansi.BoldHiCyan), string(ansi.HiCyan)
	case l < LevelWarn:
		return string(ansi.BoldBlue), string(ansi.HiBlue)
	case l < LevelError:
		return string(ansi.BoldHiYellow), string(ansi.HiYellow)
	case l < LevelFatal:
		return string(ansi.BoldHiRed), string(ansi.HiRed)
	default:
		return string(ansi.BoldHiPurple), string(ansi.HiPurple)
	}
}

type coloredHandler struct {
	uh *unstructuredHandler
}

func (h *coloredHandler) Enabled(c context.Context, l slog.Level) bool {
	return h.uh.Enabled(c, l)
}

func (h *coloredHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := NewBuffer()
	state := h.uh.ch.newHandleState(buf, true, " ")
	defer state.free()
	// time
	if !r.Time.IsZero() {
		val := r.Time.Round(0) // strip monotonic to match Attr behavior
		state.appendString(val.Format("15:04:05,000"))
	}

	lvl := r.Level
	hiColor, color := levelAnsiColor(lvl)
	state.appendString(" " + hiColor + levelLabel(lvl) + string(ansi.Reset) + " ")
	if h.uh.qualifier != "" {
		q := fmt.Sprintf("[%s%s%s] ", string(ansi.BoldWhite), h.uh.qualifier, string(ansi.Reset))
		state.appendString(q)
	}
	state.appendString(color + r.Message + string(ansi.Reset))
	defer state.free()
	state.appendNonBuiltIns(r)
	state.appendString("\n")
	return h.uh.output(r.PC, *buf)
}

func (h coloredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &coloredHandler{
		uh: h.uh.WithAttrs(attrs).(*unstructuredHandler),
	}
}

func (h coloredHandler) WithGroup(name string) slog.Handler {
	return &coloredHandler{
		uh: h.uh.WithGroup(name).(*unstructuredHandler),
	}
}

func NewColoredHandler(w io.Writer, opts *slog.HandlerOptions) *coloredHandler {
	return &coloredHandler{
		uh: NewUnstructuredHandler(w, opts),
	}
}
