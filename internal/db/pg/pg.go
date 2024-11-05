package pg

import (
	"context"

	"github.com/chistyakoviv/converter/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pg struct {
	dbc *pgxpool.Pool
}

func NewDB(dbc *pgxpool.Pool) db.DB {
	return &pg{
		dbc: dbc,
	}
}

func (p *pg) Exec(ctx context.Context, q db.Query, args ...interface{}) (pgconn.CommandTag, error) {
	return p.dbc.Exec(ctx, q.QueryRaw, args...)
}

func (p *pg) Query(ctx context.Context, q db.Query, args ...interface{}) (pgx.Rows, error) {
	return p.dbc.Query(ctx, q.QueryRaw, args...)
}

func (p *pg) QueryRow(ctx context.Context, q db.Query, args ...interface{}) pgx.Row {
	return p.dbc.QueryRow(ctx, q.QueryRaw, args...)
}

func (p *pg) Close() {
	p.dbc.Close()
}
