package validation

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/chistyakoviv/converter/internal/http-server/deps"
	resp "github.com/chistyakoviv/converter/internal/lib/http/response"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

func ValidationDecorator(
	d *deps.ConversionDeps,
	handler http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := render.DecodeJSON(r.Body, &d.Request)
		if errors.Is(err, io.EOF) {
			// Encounter such error if request body is empty
			// Handle it separately
			d.Logger.Error("request body is empty")

			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			d.Logger.Error("failed to decode request body", slogger.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		d.Logger.Debug("request body is decoded", slog.Any("request", d.Request))

		if err := d.Validator.Struct(d.Request); err != nil {
			validationErr := err.(validator.ValidationErrors)

			d.Logger.Error("invalid request", slogger.Err(err))

			render.JSON(w, r, resp.ValidationError(validationErr))

			return
		}

		handler(w, r)
	}
}
