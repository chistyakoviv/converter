package logger

import (
	"net/http"
	"time"

	"log/slog"

	"github.com/chistyakoviv/converter/internal/constants"
	"github.com/go-chi/chi/v5/middleware"
)

func New(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		logger := logger.With(
			slog.String("component", "middleware/logger"),
		)

		logger.Info("logger middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := logger.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String(constants.RequestID, middleware.GetReqID(r.Context())),
			)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				entry.Info("request completed",
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", time.Since(t1).String()),
				)
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
