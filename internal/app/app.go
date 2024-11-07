package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/chistyakoviv/converter/internal/di"
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

	logger.Info("Application is up and running", slog.String("env", cfg.Env))
	logger.Debug("Application is running in DEBUG mode")

	wg := &sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()
		// Start http server here
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
