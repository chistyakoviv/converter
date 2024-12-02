package main

import (
	"context"

	"github.com/chistyakoviv/converter/app"
)

func main() {
	ctx := context.Background()
	a := app.NewApp(ctx)
	a.Run(ctx)
}
