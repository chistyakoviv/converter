package app

import (
	"fmt"
	"log"

	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/di"
)

func Bootstrap(c di.Container) {
	cfg := config.MustLoad()

	c.RegisterSingleton("config", func(c di.Container) *config.Config {
		return cfg
	})

	conf, err := di.Resolve[*config.Config](c, "config")
	if err != nil {
		log.Fatalf("Couldn't resolve config definition: %v", err)
	}
	fmt.Printf("config: %+v\n", conf)
}
