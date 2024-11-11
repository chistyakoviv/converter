package convert

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/chistyakoviv/converter/internal/constants"
	"github.com/chistyakoviv/converter/internal/http-server/converter"
	"github.com/chistyakoviv/converter/internal/http-server/requests"
	resp "github.com/chistyakoviv/converter/internal/lib/http/response"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
	"github.com/chistyakoviv/converter/internal/repository/conversion"
	"github.com/chistyakoviv/converter/internal/service"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Response struct {
	resp.Response
	Id int64 `json:"id"`
}

func New(
	ctx context.Context,
	logger *slog.Logger,
	validation *validator.Validate,
	conversionService service.ConversionService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.conversion.New"

		logger := logger.With(
			slog.String("op", op),
			slog.String(constants.RequestID, middleware.GetReqID(r.Context())),
		)

		var req requests.ConversionRequest

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			// Encounter such error if request body is empty
			// Handle it separately
			logger.Error("request body is empty")

			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			logger.Error("failed to decode request body", slogger.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		logger.Debug("request body is decoded", slog.Any("request", req))

		// TODO: wrap handler in decorator to make validation automatically
		if err := validation.Struct(req); err != nil {
			validationErr := err.(validator.ValidationErrors)

			logger.Error("invalid request", slogger.Err(err))

			render.JSON(w, r, resp.ValidationError(validationErr))

			return
		}

		id, err := conversionService.Convert(ctx, converter.ToConversionInfoFromRequest(req))
		if errors.Is(err, conversion.ErrPathAlreadyExist) {
			logger.Debug("path already exists", slog.String("path", req.Path))

			render.JSON(w, r, resp.Error("path already exists"))

			return
		}
		if err != nil {
			logger.Error("failed to add file to conversion queue", slogger.Err(err))

			render.JSON(w, r, resp.Error("failed to add file to conversion queue"))

			return
		}

		logger.Debug("file added", slog.Int64("id", id))

		render.JSON(w, r, Response{
			Response: resp.OK(),
			Id:       id,
		})
	}
}
