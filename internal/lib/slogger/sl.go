package slogger

import (
	"fmt"
	"log/slog"
)

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

// Helper function to create slog.Attrs from an array of errors
func GroupErr(errs []error) slog.Attr {
	fields := make([]slog.Attr, len(errs))
	for i, err := range errs {
		fields[i] = slog.Attr{
			Key:   fmt.Sprintf("error_%d", i+1),
			Value: slog.StringValue(err.Error()),
		}
	}
	return slog.Attr{
		Key:   "errors",
		Value: slog.GroupValue(fields...),
	}
}
