package pg

import (
	"context"
	"log/slog"

	"github.com/chistyakoviv/converter/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type key string

const (
	TxKey key = "tx"
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
	p.logger.Debug("query debug", slog.Attr{Key: "name", Value: slog.StringValue(q.Name)}, slog.Attr{Key: "sql", Value: slog.StringValue(q.QueryRaw)})

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Exec(ctx, q.QueryRaw, args...)
	}

	return p.dbc.Exec(ctx, q.QueryRaw, args...)
}

func (p *pg) Query(ctx context.Context, q db.Query, args ...interface{}) (pgx.Rows, error) {
	p.logger.Debug("query debug", slog.Attr{Key: "name", Value: slog.StringValue(q.Name)}, slog.Attr{Key: "sql", Value: slog.StringValue(q.QueryRaw)})

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Query(ctx, q.QueryRaw, args...)
	}

	return p.dbc.Query(ctx, q.QueryRaw, args...)
}

func (p *pg) QueryRow(ctx context.Context, q db.Query, args ...interface{}) pgx.Row {
	p.logger.Debug("query debug", slog.Attr{Key: "name", Value: slog.StringValue(q.Name)}, slog.Attr{Key: "sql", Value: slog.StringValue(q.QueryRaw)})

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, q.QueryRaw, args...)
	}

	return p.dbc.QueryRow(ctx, q.QueryRaw, args...)
}

/**
* Begin acquires a connection from the Pool and starts a transaction.
* Unlike database/sql, the context only affects the begin command.
* i.e. there is no auto-rollback on context cancellation.
* Begin initiates a transaction block without explicitly setting
* a transaction mode for the block (see BeginTx with TxOptions
* if transaction mode is required). *pgxpool.Tx is returned,
* which implements the pgx.Tx interface. Commit or Rollback
* must be called on the returned transaction to finalize the transaction block.
 */
func (p *pg) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.dbc.Begin(ctx)
}

func (p *pg) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return p.dbc.BeginTx(ctx, txOptions)
}

func (p *pg) Close() {
	p.dbc.Close()
}

func MakeContextTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, TxKey, tx)
}
