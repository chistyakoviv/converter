package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Client incapsulates connections to
// different databases (master, slave)
type Client interface {
	DB() DB
	Close() error
}

// DB provides interface for working with db
type DB interface {
	QueryExecutor
	Close()
}

type Query struct {
	Name     string
	QueryRaw string
}

type QueryExecutor interface {
	Exec(ctx context.Context, q Query, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, q Query, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, q Query, args ...interface{}) pgx.Row
}
