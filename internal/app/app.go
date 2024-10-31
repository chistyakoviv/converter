package app

import "github.com/chistyakoviv/converter/internal/di"

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

}
