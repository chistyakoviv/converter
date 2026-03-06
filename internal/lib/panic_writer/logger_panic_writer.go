package panic_writer

import (
	"io"
	"log/slog"
)

type loggerPanicWriter struct {
	logger *slog.Logger
}

func NewLoggerPanicWriter(logger *slog.Logger) io.Writer {
	return &loggerPanicWriter{
		logger: logger,
	}
}

func (l *loggerPanicWriter) Write(p []byte) (n int, err error) {
	l.logger.Error("panic", slog.String("stack", string(p)))
	return len(p), nil
}
