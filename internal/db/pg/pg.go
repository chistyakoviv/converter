package pg

import (
	"context"
	"log/slog"

	"github.com/chistyakoviv/converter/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pg struct {
	name   string
	dbc    *pgxpool.Pool
	logger *slog.Logger
}

func NewDB(name string, dbc *pgxpool.Pool, logger *slog.Logger) db.DB {
	return &pg{
		name:   name,
		dbc:    dbc,
		logger: logger,
	}
}

func (p *pg) Exec(ctx context.Context, q db.Query, args ...interface{}) (pgconn.CommandTag, error) {
	p.logger.Debug("query debug", slog.Attr{Key: "query name", Value: slog.StringValue(q.Name)}, slog.Attr{Key: "raw sql", Value: slog.StringValue(q.QueryRaw)})
	return p.dbc.Exec(ctx, q.QueryRaw, args...)
}

func (p *pg) Query(ctx context.Context, q db.Query, args ...interface{}) (pgx.Rows, error) {
	p.logger.Debug("query debug", slog.Attr{Key: "query name", Value: slog.StringValue(q.Name)}, slog.Attr{Key: "raw sql", Value: slog.StringValue(q.QueryRaw)})
	return p.dbc.Query(ctx, q.QueryRaw, args...)
}

func (p *pg) QueryRow(ctx context.Context, q db.Query, args ...interface{}) pgx.Row {
	p.logger.Debug("query debug", slog.Attr{Key: "query name", Value: slog.StringValue(q.Name)}, slog.Attr{Key: "raw sql", Value: slog.StringValue(q.QueryRaw)})
	return p.dbc.QueryRow(ctx, q.QueryRaw, args...)
}

func (p *pg) Close() {
	p.dbc.Close()
}
