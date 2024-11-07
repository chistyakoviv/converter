package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/chistyakoviv/converter/internal/di"
	"github.com/chistyakoviv/converter/internal/lib/sl"
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

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(middleware.NoCache)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	// TODO: move to di container
	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	logger.Info("starting server", slog.String("address", cfg.HTTPServer.Address), slog.String("env", cfg.Env))

	go func() {
		// http router
		defer wg.Done()

		// ListenAndServe always returns a non-nil error. After [Server.Shutdown] or [Server.Close], the returned error is [ErrServerClosed].
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			logger.Info("http server is closing gracefully")
		} else {
			logger.Error("http server error", sl.Err(err))
		}
		logger.Info("http server stopped")
	}()

	a.gracefulShutdown(ctx, cancel, wg, srv)
}

// TODO: refactor graceful shutdown to properly terminate all components
func (a *app) gracefulShutdown(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, srv *http.Server) {
	logger := resolveLogger(a.container)

	select {
	case <-ctx.Done():
		logger.Info("terminating: context canceled")
	case <-waitSignal():
		logger.Info("terminating: via signal")
	}

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("failed to stop server", sl.Err(err))
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
