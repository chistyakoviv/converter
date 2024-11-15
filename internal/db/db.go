package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TxHandler func(ctx context.Context) error

// Client incapsulates connections to
// different databases (master, slave)
type Client interface {
	DB() DB
	Close() error
}

type TxManager interface {
	ReadCommitted(ctx context.Context, f TxHandler) error
}

// DB provides interface for working with db
type DB interface {
	QueryExecutor
	Transactor
	Close()
}

type Query struct {
	Name     string
	QueryRaw string
}

type Transactor interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type QueryExecutor interface {
	Exec(ctx context.Context, q Query, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, q Query, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, q Query, args ...interface{}) pgx.Row
}
