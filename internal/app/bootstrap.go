package app

import (
	"log"
	"log/slog"
	"os"

	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/db"
	"github.com/chistyakoviv/converter/internal/db/pg"
	"github.com/chistyakoviv/converter/internal/di"
)

// Retrieves the application configuration from the dependency injection container,
// centralizing error handling to avoid repetitive error checks across the codebase.
// Logs a fatal error and terminates the program if the configuration cannot be resolved.
func resolveConfig(c di.Container) *config.Config {
	cfg, err := di.Resolve[*config.Config](c, "config")

	if err != nil {
		log.Fatalf("Couldn't resolve config definition: %v", err)
	}

	return cfg
}

func resolveLogger(c di.Container) *slog.Logger {
	logger, err := di.Resolve[*slog.Logger](c, "logger")

	if err != nil {
		log.Fatalf("Couldn't resolve logger definition: %v", err)
	}

	return logger
}

func resolveDbClient(c di.Container) db.Client {
	client, err := di.Resolve[db.Client](c, "db")

	if err != nil {
		log.Fatalf("Couldn't resolve db client definition: %v", err)
	}

	return client
}

func bootstrap(c di.Container) {
	c.RegisterSingleton("config", func(c di.Container) *config.Config {
		cfg := config.MustLoad()
		return cfg
	})

	c.RegisterSingleton("logger", func(c di.Container) *slog.Logger {
		cfg := resolveConfig(c)

		var logger *slog.Logger

		switch cfg.Env {
		case config.EnvProd:
			logger = slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
			)
		case config.EnvDev:
			logger = slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			)
		default:
			logger = slog.New(
				slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			)
		}

		return logger
	})

	c.RegisterSingleton("db", func(c di.Container) db.Client {
		cfg := resolveConfig(c)
		client, err := pg.NewClient(c.Context(), cfg.Postgres.Dsn)

		if err != nil {
			log.Fatalf("failed to create db client: %v", err)
		}

		return client
	})

	resolveDbClient(c)
}
