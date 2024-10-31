package main

import (
	"github.com/chistyakoviv/converter/internal/app"
)

func main() {
	app := app.NewApp()
	app.Run()
}

func test[V any](v V) V {
	return v
}
