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
	taskService service.TaskService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoratedLogger := loggerDecorator.LoggerDecorator("handlers.conversion.New", logger, r)

		var req request.ConversionRequest

		err := validationrDecorator.ValidationDecorator(decoratedLogger, validation, &req, w, r)
		if err != nil {
			return
		}

		id, err := conversionService.Add(ctx, converter.ToConversionInfoFromRequest(req))
		if errors.Is(err, conversionq.ErrPathAlreadyExist) || errors.Is(err, conversionq.ErrFilestemAlreadyExist) {
			decoratedLogger.Debug("file with the specified path or filestem already exists", slog.String("path", req.Path))

			render.Status(r, http.StatusConflict) // 409
			render.JSON(w, r, resp.Error("file with the specified path already exists"))

			return
		}
		if errors.Is(err, conversionq.ErrFileDoesNotExist) {
			decoratedLogger.Debug("file does not exist", slog.String("path", req.Path))

			render.Status(r, http.StatusNotFound) // 404
			render.JSON(w, r, resp.Error("file does not exist"))

			return
		}
		if errors.Is(err, conversionq.ErrFileTypeNotSupported) {
			decoratedLogger.Debug("file type not supported", slog.String("path", req.Path))

			render.Status(r, http.StatusBadRequest) // 400
			render.JSON(w, r, resp.Error("file type not supported"))

			return
		}
		if errors.Is(err, conversionq.ErrFailedDetermineFileType) {
			decoratedLogger.Debug("failed to determine file type", slog.String("path", req.Path))

			render.Status(r, http.StatusUnprocessableEntity) // 422
			render.JSON(w, r, resp.Error("failed to determine file type"))

			return
		}
		if errors.Is(err, conversionq.ErrInvalidConversionFormat) {
			decoratedLogger.Debug("cannot convert to the specified format", slog.String("path", req.Path))

			render.Status(r, http.StatusBadRequest) // 400
			render.JSON(w, r, resp.Error("cannot convert to the specified format"))

			return
		}
		if errors.Is(err, conversionq.ErrEmptyTargetFormatList) {
			decoratedLogger.Debug("target format list is empty", slog.String("path", req.Path))

			render.Status(r, http.StatusBadRequest) // 400
			render.JSON(w, r, resp.Error("target format list is empty"))

			return
		}
		if err != nil {
			decoratedLogger.Error("failed to add file to conversion queue", slogger.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to add file to conversion queue"))

			return
		}

		decoratedLogger.Debug("file added", slog.Int64("id", id))

		// Try to process the file immediately
		taskService.TryQueueConversion()

		render.JSON(w, r, ConversionResponse{
			Response: resp.OK(),
			Id:       id,
		})
	}
}
