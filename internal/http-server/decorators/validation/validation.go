package validation

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	resp "github.com/chistyakoviv/converter/internal/lib/http/response"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

func ValidationDecorator(
	logger *slog.Logger,
	validation *validator.Validate,
	data interface{},
	w http.ResponseWriter,
	r *http.Request,
) error {
	err := render.DecodeJSON(r.Body, data)
	if errors.Is(err, io.EOF) {
		// Encounter such error if request body is empty
		// Handle it separately
		logger.Error("request body is empty")

		render.JSON(w, r, resp.Error("empty request"))

		return errors.New("empty request")
	}
	if err != nil {
		logger.Error("failed to decode request body", slogger.Err(err))

		render.JSON(w, r, resp.Error("failed to decode request"))

		return errors.New("failed to decode request")
	}

	logger.Debug("request body is decoded", slog.Any("request", data))

	if err := validation.Struct(data); err != nil {
		validationErr := err.(validator.ValidationErrors)

		logger.Error("invalid request", slogger.Err(err))

		render.JSON(w, r, resp.ValidationError(validationErr))

		return errors.New("invalid request")
	}

	return nil
}
