package app

import (
	"context"
	"net/http"

	"github.com/chistyakoviv/converter/internal/di"
	"github.com/chistyakoviv/converter/internal/http-server/handlers/convert"
	"github.com/chistyakoviv/converter/internal/http-server/handlers/delete"
	"github.com/chistyakoviv/converter/internal/http-server/handlers/scan"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
)

func initRoutes(ctx context.Context, c di.Container) {
	router := resolveRouter(c)

	router.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		logger := resolveLogger(c)
		if _, err := w.Write([]byte("alive")); err != nil {
			// optional: log or handle the error
			logger.Error("failed to write response: %v", slogger.Err(err))
		}
	})

	router.Post("/convert", convert.New(
		ctx,
		resolveLogger(c),
		resolveValidator(c),
		resolveConversionQueueService(c),
		resolveTaskService(c),
	))

	router.Post("/delete", delete.New(
		ctx,
		resolveLogger(c),
		resolveValidator(c),
		resolveDeletionQueueService(c),
		resolveTaskService(c),
	))

	router.Post("/scan", scan.New(
		ctx,
		resolveLogger(c),
		resolveTaskService(c),
	))
}
