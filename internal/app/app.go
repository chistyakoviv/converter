package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/chistyakoviv/converter/internal/di"
	"github.com/chistyakoviv/converter/internal/lib/sl"
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
	dq := resolveDeferredQ(a.container)

	logger.Debug("Application is running in DEBUG mode")

	initRoutes(resolveRouter(a.container))

	go func() {
		logger.Info("starting server", slog.String("address", cfg.HTTPServer.Address), slog.String("env", cfg.Env))

		srv := resolveHttpServer(a.container)
		dq.Add(func() error {
			return srv.Shutdown(ctx)
		})

		// ListenAndServe always returns a non-nil error. After [Server.Shutdown] or [Server.Close], the returned error is [ErrServerClosed].
		err := srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server error", sl.Err(err))
		}
		logger.Info("http server stopped")
	}()

	// Graceful Shutdown
	select {
	case <-ctx.Done():
		logger.Info("terminating: context canceled")
	// No need for a wait group until the application is blocked, waiting for an OS signal.
	case <-waitSignal():
		logger.Info("terminating: via signal")
	}

	// Call all deferred functions and wait them to be done
	dq.Release()
	dq.Wait()

	cancel()
	// if wg != nil {
	// 	wg.Wait()
	// }
}

func waitSignal() chan os.Signal {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	return stop
}
