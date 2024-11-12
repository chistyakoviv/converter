package app

import (
	"context"
	"net/http"

	"github.com/chistyakoviv/converter/internal/di"
	"github.com/chistyakoviv/converter/internal/http-server/decorators/logger"
	"github.com/chistyakoviv/converter/internal/http-server/decorators/validation"
	"github.com/chistyakoviv/converter/internal/http-server/deps"
	"github.com/chistyakoviv/converter/internal/http-server/handlers/convert"
	"github.com/chistyakoviv/converter/internal/pipe"
)

func initRoutes(ctx context.Context, c di.Container) {
	router := resolveRouter(c)

	router.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("alive"))
	})

	h := pipe.New[deps.ConversionDeps, http.HandlerFunc](&deps.ConversionDeps{
		Ctx:               ctx,
		Logger:            resolveLogger(c),
		Validator:         resolveValidator(c),
		ConversionService: resolveConversionService(c),
	}).
		Pipe(logger.LoggerDecorator("handlers.conversion.New")).
		Pipe(validation.ValidationDecorator).
		Pipe(convert.New)

	router.Post("/convert", h.Build())
}
