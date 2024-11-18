package convert

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/chistyakoviv/converter/internal/http-server/converter"
	loggerDecorator "github.com/chistyakoviv/converter/internal/http-server/decorators/logger"
	validationrDecorator "github.com/chistyakoviv/converter/internal/http-server/decorators/validation"
	"github.com/chistyakoviv/converter/internal/http-server/request"
	resp "github.com/chistyakoviv/converter/internal/lib/http/response"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
	"github.com/chistyakoviv/converter/internal/service"
	"github.com/chistyakoviv/converter/internal/service/conversionq"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type ConversionResponse struct {
	resp.Response
	Id int64 `json:"id"`
}

func New(
	ctx context.Context,
	logger *slog.Logger,
	validation *validator.Validate,
	conversionService service.ConversionQueueService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoratedLogger := loggerDecorator.LoggerDecorator("handlers.conversion.New", logger, r)

		var req request.ConversionRequest

		err := validationrDecorator.ValidationDecorator(decoratedLogger, validation, &req, w, r)
		if err != nil {
			return
		}

		id, err := conversionService.Add(ctx, converter.ToConversionInfoFromRequest(req))
		if errors.Is(err, conversionq.ErrPathAlreadyExist) {
			decoratedLogger.Debug("file with the specified path already exists", slog.String("path", req.Path))

			render.JSON(w, r, resp.Error("file with the specified path already exists"))

			return
		}
		if err != nil {
			decoratedLogger.Error("failed to add file to conversion queue", slogger.Err(err))

			render.JSON(w, r, resp.Error("failed to add file to conversion queue"))

			return
		}

		decoratedLogger.Debug("file added", slog.Int64("id", id))

		render.JSON(w, r, ConversionResponse{
			Response: resp.OK(),
			Id:       id,
		})
	}
}
