package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/chistyakoviv/converter/internal/di"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Application interface {
	Run(ctx context.Context)
	Container() di.Container
}

type app struct {
	container di.Container
}

func NewApp(ctx context.Context) Application {
	container := di.NewContainer()
	a := &app{
		container: container,
	}
	a.init(ctx)
	return a
}

func (a *app) Container() di.Container {
	return a.container
}

func (a *app) init(ctx context.Context) {
	bootstrap(ctx, a.container)
}

func (a *app) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)

	cfg := resolveConfig(a.container)
	logger := resolveLogger(a.container)

	logger.Debug("Application is running in DEBUG mode")

	wg := &sync.WaitGroup{}

	wg.Add(1)

	go func() {
		// http router
		defer wg.Done()

		router := chi.NewRouter()

		router.Use(middleware.RequestID)
		router.Use(middleware.Logger)
		router.Use(middleware.Recoverer)
		router.Use(middleware.URLFormat)
		router.Use(middleware.NoCache)

		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hi"))
		})

		srv := &http.Server{
			Addr:         cfg.HTTPServer.Address,
			Handler:      router,
			ReadTimeout:  cfg.HTTPServer.ReadTimeout,
			WriteTimeout: cfg.HTTPServer.WriteTimeout,
			IdleTimeout:  cfg.HTTPServer.IdleTimeout,
		}

		logger.Info("starting server", slog.String("address", cfg.HTTPServer.Address), slog.String("env", cfg.Env))

		if err := srv.ListenAndServe(); err != nil {
			logger.Error("failed to start server")
		}
	}()

	a.gracefulShutdown(ctx, cancel, wg)
}

func (a *app) gracefulShutdown(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	logger := resolveLogger(a.container)

	select {
	case <-ctx.Done():
		logger.Info("terminating: context cancelled")
	case <-waitSignal():
		logger.Info("terminating: via signal")
	}

	cancel()
	if wg != nil {
		wg.Wait()
	}
}

func waitSignal() chan os.Signal {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	return stop
}
