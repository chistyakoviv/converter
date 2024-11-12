package logger

import (
	"log/slog"
	"net/http"

	"github.com/chistyakoviv/converter/internal/constants"
	"github.com/chistyakoviv/converter/internal/http-server/deps"
	"github.com/chistyakoviv/converter/internal/pipe"
	"github.com/go-chi/chi/v5/middleware"
)

func LoggerDecorator(op string) pipe.HandlerFn[deps.ConversionDeps, http.HandlerFunc] {
	return func(
		d *deps.ConversionDeps,
		handler http.HandlerFunc,
	) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			d.Logger = d.Logger.With(
				slog.String("op", op),
				slog.String(constants.RequestID, middleware.GetReqID(r.Context())),
			)

			handler(w, r)
		}
	}
}
