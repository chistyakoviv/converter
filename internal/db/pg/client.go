package pg

import (
	"context"
	"log/slog"

	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chistyakoviv/converter/internal/db"
)

type pgClient struct {
	logger    *slog.Logger
	masterDBC db.DB
}

func NewClient(ctx context.Context, dsn string, logger *slog.Logger) (db.Client, error) {
	const op = "db.pg.NewClient"

	dbc, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &pgClient{
		masterDBC: NewDB(
			"master db",
			dbc,
			logger,
		),
		logger: logger,
	}, nil
}

func (c *pgClient) DB() db.DB {
	return c.masterDBC
}

func (c *pgClient) Close() error {
	if c.masterDBC != nil {
		c.masterDBC.Close()
	}

	return nil
}
