package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	// _ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chistyakoviv/converter/internal/di"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
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
	taskService := resolveTaskService(a.container)

	logger.Debug("Application is running in DEBUG mode")

	initRoutes(ctx, a.container)

	// Exec the command to empty the memory buffer: echo 3 | sudo tee /proc/sys/vm/drop_caches
	// see https://medium.com/@bobzsj87/demist-the-memory-ghost-d6b7cf45dd2a
	// pprof
	// if cfg.Env == config.EnvLocal {
	// 	go func() {
	// 		logger.Info("pprof server started")
	// 		http.ListenAndServe("0.0.0.0:6060", nil)
	// 	}()
	// }

	// http server
	go func() {
		logger.Info("starting http server", slog.String("address", cfg.HTTPServer.Address), slog.String("env", cfg.Env))

		srv := resolveHttpServer(a.container)
		dq.Add(func() error {
			return srv.Shutdown(ctx)
		})

		// ListenAndServe always returns a non-nil error. After [Server.Shutdown] or [Server.Close], the returned error is [ErrServerClosed]
		err := srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server error", slogger.Err(err))
		}
		logger.Info("http server stopped")
	}()

	// Periodic task scheduling
	go func() {
		logger.Info("periodic task scheduling started", slog.String("timeout", cfg.Task.CheckTimeout.String()))

		ticker := time.NewTicker(cfg.Task.CheckTimeout)
		dq.Add(func() error {
			ticker.Stop()
			taskService.Shutdown()
			return nil
		})

		for range ticker.C {
			taskService.TryQueueConversion()
			taskService.TryQueueDeletion()
		}
	}()

	// Process queues
	go func() {
		logger.Info("tasks processing started")

		// Processing automatically stops when Shutdown is called
		// by the periodic task scheduling goroutine
		taskService.ProcessQueues(ctx)
	}()

	// Graceful Shutdown
	select {
	case <-ctx.Done():
		logger.Info("terminating: context canceled")
	// No need for a wait group until the application is blocked, waiting for an OS signal
	case <-waitSignal():
		logger.Info("terminating: via signal")
	}

	// Call all deferred functions and wait them to be done
	dq.Release()
	dq.Wait()

	cancel()
}

func waitSignal() chan os.Signal {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	return stop
}
