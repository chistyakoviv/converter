package app

import (
	"log/slog"

	"github.com/chistyakoviv/converter/internal/di"
)

type Application interface {
	Run()
	Container() di.Container
}

type app struct {
	container di.Container
}

func NewApp() Application {
	container := di.NewContainer()
	a := &app{
		container: container,
	}
	a.init()
	return a
}

func (a *app) Container() di.Container {
	return a.container
}

func (a *app) init() {
	Bootstrap(a.container)
}

func (a *app) Run() {
	cfg := resolveConfig(a.container)
	logger := resolveLogger(a.container)

	logger.Info("Application is up and running", slog.String("env", cfg.Env))
	logger.Debug("Application is running in DEBUG mode")
}
