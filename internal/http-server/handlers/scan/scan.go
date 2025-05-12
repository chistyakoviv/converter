package scan

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/chistyakoviv/converter/internal/constants"
	loggerDecorator "github.com/chistyakoviv/converter/internal/http-server/decorators/logger"
	resp "github.com/chistyakoviv/converter/internal/lib/http/response"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
	"github.com/chistyakoviv/converter/internal/service"
	"github.com/go-chi/render"
)

type ScanResponse struct {
	resp.Response
}

func New(
	ctx context.Context,
	logger *slog.Logger,
	taskService service.TaskService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoratedLogger := loggerDecorator.LoggerDecorator("handlers.scan.New", logger, r)

		if taskService.IsScanning() {
			decoratedLogger.Debug("scan is already running")

			render.Status(r, http.StatusConflict) // 409
			render.JSON(w, r, resp.Error("scan is already running"))

			return
		}

		// Do not wait for the scan to complete
		go func() {
			err := taskService.ProcessScanfs(ctx, constants.FilesRootDir)
			if err != nil {
				decoratedLogger.Error("failed to scan filesystem: ", slogger.Err(err))
			}

			decoratedLogger.Debug("scan completed")

			// Try to process the files immediately
			taskService.TryQueueConversion()
		}()

		render.JSON(w, r, ScanResponse{
			Response: resp.OK(),
		})
	}
}
