package engine

import "round_table/apps/server/internal/stream"

// TeeProgressLogger forwards progress lines to multiple loggers.
type TeeProgressLogger struct {
	Loggers []ProgressLogger
}

func (t TeeProgressLogger) Logf(format string, args ...any) {
	for _, l := range t.Loggers {
		if l != nil {
			l.Logf(format, args...)
		}
	}
}

// TeeStreamLogger forwards LLM stream events to multiple loggers.
type TeeStreamLogger struct {
	Loggers []StreamLogger
}

func (t TeeStreamLogger) Start(meta stream.Meta) {
	for _, l := range t.Loggers {
		if l != nil {
			l.Start(meta)
		}
	}
}

func (t TeeStreamLogger) Delta(delta string) {
	for _, l := range t.Loggers {
		if l != nil {
			l.Delta(delta)
		}
	}
}

func (t TeeStreamLogger) End() {
	for _, l := range t.Loggers {
		if l != nil {
			l.End()
		}
	}
}
