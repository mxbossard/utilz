package zlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/mxbossard/utilz/anzi"
	"github.com/mxbossard/utilz/formatz"
)

func levelAnsiColor(l slog.Level) (string, string) {
	switch {
	case l < LevelPerf:
		return string(anzi.BoldHiWhite), string(anzi.HiWhite)
	case l < LevelDebug:
		return string(anzi.BoldHiGreen), string(anzi.HiGreen)
	case l < LevelInfo:
		return string(anzi.BoldHiCyan), string(anzi.HiCyan)
	case l < LevelWarn:
		return string(anzi.BoldBlue), string(anzi.HiBlue)
	case l < LevelError:
		return string(anzi.BoldHiYellow), string(anzi.HiYellow)
	case l < LevelFatal:
		return string(anzi.BoldHiRed), string(anzi.HiRed)
	default:
		return string(anzi.BoldHiPurple), string(anzi.HiPurple)
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
	state.appendString(" " + hiColor + levelShortLabel(lvl) + string(anzi.Reset) + " ")
	if h.uh.qualifier != "" || defaultPart != "" {
		part := ""
		if defaultPart != "" {
			part = fmt.Sprintf("%s:", defaultPart)
		}
		qualifier := formatz.PadLeft(h.uh.qualifier, QualifierPadding)
		qualifier = formatz.TruncateLeftPrefix(qualifier, QualifierPadding, "...")
		q := fmt.Sprintf("[%s%s%s%s] ", part, string(anzi.BoldWhite), qualifier, string(anzi.Reset))
		state.appendString(q)
	}

	state.appendString(color + r.Message + string(anzi.Reset))

	state.appendNonBuiltIns(r)

	// source
	if h.uh.ch.opts.AddSource {
		state.appendAttr(slog.Any(slog.SourceKey, source(r, h.uh.packageName)))
	}

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
