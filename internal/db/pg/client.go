package pg

import (
	"context"

	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chistyakoviv/converter/internal/db"
)

type pgClient struct {
	masterDBC db.DB
}

func NewClient(ctx context.Context, dsn string) (db.Client, error) {
	const op = "db.pg.NewClient"

	dbc, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &pgClient{
		masterDBC: &pg{dbc: dbc},
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
