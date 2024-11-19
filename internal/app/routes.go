package app

import (
	"context"
	"net/http"

	"github.com/chistyakoviv/converter/internal/di"
	"github.com/chistyakoviv/converter/internal/http-server/handlers/convert"
)

func initRoutes(ctx context.Context, c di.Container) {
	router := resolveRouter(c)

	router.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("alive"))
	})

	router.Post("/convert", convert.New(
		ctx,
		resolveLogger(c),
		resolveValidator(c),
		resolveConversionQueueService(c),
		resolveTaskService(c),
	))
}
