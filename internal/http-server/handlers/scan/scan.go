package scan

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/chistyakoviv/converter/internal/constants"
	loggerDecorator "github.com/chistyakoviv/converter/internal/http-server/decorators/logger"
	resp "github.com/chistyakoviv/converter/internal/lib/http/response"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
	"github.com/chistyakoviv/converter/internal/service"
	"github.com/chistyakoviv/converter/internal/service/task"
	"github.com/go-chi/render"
)

type DeletionResponse struct {
	resp.Response
}

func New(
	ctx context.Context,
	logger *slog.Logger,
	taskService service.TaskService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoratedLogger := loggerDecorator.LoggerDecorator("handlers.scan.New", logger, r)

		err := taskService.ProcessScanfs(ctx, constants.FilesRootDir)

		if errors.Is(err, task.ErrScanAlreadyRunning) {
			decoratedLogger.Debug("scan is already running")

			render.Status(r, http.StatusConflict) // 409
			render.JSON(w, r, resp.Error("scan is already running"))

			return
		}
		if err != nil {
			decoratedLogger.Error("failed to scan files", slogger.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to scan files"))

			return
		}

		decoratedLogger.Debug("scan completed")

		// Try to process the files immediately
		taskService.TryQueueConversion()

		render.JSON(w, r, resp.OK())
	}
}
