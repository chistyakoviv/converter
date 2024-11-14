package logger

import (
	"log/slog"
	"net/http"

	"github.com/chistyakoviv/converter/internal/constants"
	"github.com/go-chi/chi/v5/middleware"
)

func LoggerDecorator(
	op string,
	logger *slog.Logger,
	r *http.Request,
) *slog.Logger {
	return logger.With(
		slog.String("op", op),
		slog.String(constants.RequestID, middleware.GetReqID(r.Context())),
	)
}
