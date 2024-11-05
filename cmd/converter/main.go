package main

import (
	"context"

	"github.com/chistyakoviv/converter/internal/app"
)

func main() {
	ctx := context.Background()
	app := app.NewApp(ctx)
	app.Run()
}
