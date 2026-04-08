package main

import (
	"context"

	"github.com/chistyakoviv/converter/app"
)

func main() {
	ctx := context.Background()
	// TODO: remove ctx from app and db client constructors and return error from http handler if context is canceled
	a := app.NewApp(ctx)
	a.Run(ctx)
}
