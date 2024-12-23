package app

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	sq "github.com/Masterminds/squirrel"
	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/converter"
	"github.com/chistyakoviv/converter/internal/converter/ffmpeggo"
	"github.com/chistyakoviv/converter/internal/converter/govips"
	"github.com/chistyakoviv/converter/internal/db"
	"github.com/chistyakoviv/converter/internal/db/pg"
	"github.com/chistyakoviv/converter/internal/db/transaction"
	"github.com/chistyakoviv/converter/internal/deferredq"
	"github.com/chistyakoviv/converter/internal/di"
	mwLogger "github.com/chistyakoviv/converter/internal/http-server/middleware/logger"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
	"github.com/chistyakoviv/converter/internal/repository"
	conversionRepository "github.com/chistyakoviv/converter/internal/repository/conversion"
	deletionRepository "github.com/chistyakoviv/converter/internal/repository/deletion"
	"github.com/chistyakoviv/converter/internal/service"
	conversionQueueService "github.com/chistyakoviv/converter/internal/service/conversionq"
	converterService "github.com/chistyakoviv/converter/internal/service/converter"
	deletionQueueService "github.com/chistyakoviv/converter/internal/service/deletionq"
	"github.com/chistyakoviv/converter/internal/service/task"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

func bootstrap(ctx context.Context, c di.Container) {
	c.RegisterSingleton("config", func(c di.Container) *config.Config {
		cfg := config.MustLoad(nil)
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

		logger = logger.With(
			slog.String("service", "converter"),
		)

		return logger
	})

	c.RegisterSingleton("db", func(c di.Container) db.Client {
		cfg := resolveConfig(c)
		logger := resolveLogger(c)
		dq := resolveDeferredQ(c)

		client, err := pg.NewClient(ctx, cfg.Postgres.Dsn, logger)

		// Close db connections
		dq.Add(func() error {
			defer logger.Info("db connections closed")
			return client.Close()
		})

		if err != nil {
			logger.Error("failed to create db client", slogger.Err(err))
			os.Exit(1)
		}

		return client
	})

	c.RegisterSingleton("sq", func(c di.Container) sq.StatementBuilderType {
		return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	})

	c.RegisterSingleton("router", func(c di.Container) *chi.Mux {
		router := chi.NewRouter()
		logger := resolveLogger(c)

		router.Use(middleware.RequestID)
		// Replace middleware.Logger with custom logger middleware to keep logs consistent with the rest of the application
		// router.Use(middleware.Logger)
		router.Use(mwLogger.New(logger))
		// router.Use(middleware.Heartbeat("/ping"))
		router.Use(middleware.Recoverer)
		router.Use(middleware.URLFormat)
		router.Use(middleware.NoCache)

		return router
	})

	c.RegisterSingleton("httpServer", func(c di.Container) *http.Server {
		cfg := resolveConfig(c)
		router := resolveRouter(c)

		return &http.Server{
			Addr:         cfg.HTTPServer.Address,
			Handler:      router,
			ReadTimeout:  cfg.HTTPServer.ReadTimeout,
			WriteTimeout: cfg.HTTPServer.WriteTimeout,
			IdleTimeout:  cfg.HTTPServer.IdleTimeout,
		}
	})

	c.RegisterSingleton("dq", func(c di.Container) deferredq.DQueue {
		return deferredq.New(resolveLogger(c))
	})

	c.RegisterSingleton("validator", func(c di.Container) *validator.Validate {
		return validator.New()
	})

	c.RegisterSingleton("txManager", func(c di.Container) db.TxManager {
		return transaction.NewTransactionManager(resolveDbClient(c).DB())
	})

	c.RegisterSingleton("imageConverter", func(c di.Container) converter.ImageConverter {
		serv := govips.NewImageConverter(resolveConfig(c), resolveLogger(c))
		dq := resolveDeferredQ(c)
		logger := resolveLogger(c)

		dq.Add(func() error {
			defer logger.Info("image converter shutdown")
			serv.Shutdown()
			return nil
		})

		return serv
	})

	c.RegisterSingleton("videoConverter", func(c di.Container) converter.VideoConverter {
		serv := ffmpeggo.NewVideoConverter(resolveConfig(c), resolveLogger(c))
		dq := resolveDeferredQ(c)
		logger := resolveLogger(c)

		dq.Add(func() error {
			defer logger.Info("video converter shutdown")
			serv.Shutdown()
			return nil
		})

		return serv
	})

	// Repositories
	c.RegisterSingleton("conversionQueueRepository", func(c di.Container) repository.ConversionQueueRepository {
		return conversionRepository.NewRepository(resolveDbClient(c), resolveStatementBuilder(c))
	})

	c.RegisterSingleton("deletionQueueRepository", func(c di.Container) repository.DeletionQueueRepository {
		return deletionRepository.NewRepository(resolveDbClient(c), resolveStatementBuilder(c))
	})

	// Services
	c.RegisterSingleton("conversionQueueService", func(c di.Container) service.ConversionQueueService {
		return conversionQueueService.NewService(
			resolveConfig(c),
			resolveTxManager(c),
			resolveConversionQueueRepository(c),
		)
	})

	c.RegisterSingleton("deletionQueueService", func(c di.Container) service.DeletionQueueService {
		return deletionQueueService.NewService(
			resolveConfig(c),
			resolveLogger(c),
			resolveTxManager(c),
			resolveDeletionQueueRepository(c),
			resolveConversionQueueRepository(c),
		)
	})

	c.RegisterSingleton("taskService", func(c di.Container) service.TaskService {
		return task.NewService(
			resolveLogger(c),
			resolveConversionQueueService(c),
			resolveDeletionQueueService(c),
			resolveConverterService(c),
		)
	})

	c.RegisterSingleton("converterService", func(c di.Container) service.ConverterService {
		serv, err := converterService.NewService(resolveConfig(c),
			resolveLogger(c),
			resolveImageConverter(c),
			resolveVideoConverter(c),
		)

		if err != nil {
			log.Fatalf("Couldn't create converter service: %v", err)
		}

		return serv
	})
}
