package app

import (
	"log"
	"log/slog"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/db"
	"github.com/chistyakoviv/converter/internal/deferredq"
	"github.com/chistyakoviv/converter/internal/di"
	"github.com/chistyakoviv/converter/internal/repository"
	"github.com/chistyakoviv/converter/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
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

func resolveStatementBuilder(c di.Container) sq.StatementBuilderType {
	sq, err := di.Resolve[sq.StatementBuilderType](c, "sq")

	if err != nil {
		log.Fatalf("Couldn't resolve statement builder definition: %v", err)
	}

	return sq
}

func resolveRouter(c di.Container) *chi.Mux {
	router, err := di.Resolve[*chi.Mux](c, "router")

	if err != nil {
		log.Fatalf("Couldn't resolve router definition: %v", err)
	}

	return router
}

func resolveHttpServer(c di.Container) *http.Server {
	srv, err := di.Resolve[*http.Server](c, "httpServer")

	if err != nil {
		log.Fatalf("Couldn't resolve http server definition: %v", err)
	}

	return srv
}

func resolveDeferredQ(c di.Container) deferredq.DQueue {
	dq, err := di.Resolve[deferredq.DQueue](c, "dq")

	if err != nil {
		log.Fatalf("Couldn't resolve deferred queue definition: %v", err)
	}

	return dq
}

func resolveValidator(c di.Container) *validator.Validate {
	validator, err := di.Resolve[*validator.Validate](c, "validator")

	if err != nil {
		log.Fatalf("Couldn't resolve validator definition: %v", err)
	}

	return validator
}

// Repositories
func resolveConversionQueueRepository(c di.Container) repository.ConversionQueueRepository {
	repo, err := di.Resolve[repository.ConversionQueueRepository](c, "conversionQueueRepository")

	if err != nil {
		log.Fatalf("Couldn't resolve conversion queue repository definition: %v", err)
	}

	return repo
}

// Services
func resolveConversionService(c di.Container) service.ConversionService {
	serv, err := di.Resolve[service.ConversionService](c, "conversionService")

	if err != nil {
		log.Fatalf("Couldn't resolve conversion service definition: %v", err)
	}

	return serv
}
