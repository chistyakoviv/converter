package delete

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
	"github.com/chistyakoviv/converter/internal/service/deletionq"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type DeletionResponse struct {
	resp.Response
	Id int64 `json:"id"`
}

func New(
	ctx context.Context,
	logger *slog.Logger,
	validation *validator.Validate,
	deletionService service.DeletionQueueService,
	taskService service.TaskService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoratedLogger := loggerDecorator.LoggerDecorator("handlers.deletion.New", logger, r)

		var req request.DeletionRequest

		err := validationrDecorator.ValidationDecorator(decoratedLogger, validation, &req, w, r)
		if err != nil {
			return
		}

		id, err := deletionService.Add(ctx, converter.ToDeletionInfoFromRequest(req))
		if errors.Is(err, deletionq.ErrPathAlreadyExist) {
			decoratedLogger.Debug("file with the specified path already exists in the deletion queue", slog.String("path", req.Path))

			render.Status(r, http.StatusConflict) // 409
			render.JSON(w, r, resp.Error("file with the specified path already exists in the deletion queue"))

			return
		}
		if errors.Is(err, deletionq.ErrFileDoesNotExist) {
			decoratedLogger.Debug("file does not exist", slog.String("path", req.Path))

			render.Status(r, http.StatusNotFound) // 404
			render.JSON(w, r, resp.Error("file does not exist"))

			return
		}
		if err != nil {
			decoratedLogger.Error("failed to add file to deletion queue", slogger.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to add file to deletion queue"))

			return
		}

		decoratedLogger.Debug("file added to deletion queue", slog.Int64("id", id))

		// Try to process the file immediately
		taskService.TryQueueDeletion()

		render.JSON(w, r, DeletionResponse{
			Response: resp.OK(),
			Id:       id,
		})
	}
}
