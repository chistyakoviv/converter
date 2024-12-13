package dummy

import (
	"context"

	"log/slog"
)

func NewDummyLogger() *slog.Logger {
	return slog.New(NewDummyHandler())
}

type DummyHandler struct{}

func NewDummyHandler() *DummyHandler {
	return &DummyHandler{}
}

func (h *DummyHandler) Handle(_ context.Context, _ slog.Record) error {
	// ingnore the record
	return nil
}

func (h *DummyHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	// return unchanged handler
	return h
}

func (h *DummyHandler) WithGroup(_ string) slog.Handler {
	// return unchanged handler
	return h
}

func (h *DummyHandler) Enabled(_ context.Context, _ slog.Level) bool {
	// do nothing
	return false
}
