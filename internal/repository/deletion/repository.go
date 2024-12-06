package deletion

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/chistyakoviv/converter/internal/db"
	"github.com/chistyakoviv/converter/internal/model"
	"github.com/chistyakoviv/converter/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

const (
	tablename = "deletion_queue"

	idColumn        = "id"
	fullpathColumn  = "fullpath"
	statusColumn    = "status"
	errorCodeColumn = "error_code"
	createdAtColumn = "created_at"
	updatedAtColumn = "updated_at"
)

type repo struct {
	db db.Client
	sq sq.StatementBuilderType
}

func NewRepository(db db.Client, sq sq.StatementBuilderType) repository.DeletionQueueRepository {
	return &repo{
		db: db,
		sq: sq,
	}
}

func (r *repo) Create(ctx context.Context, file *model.DeletionInfo) (int64, error) {
	ts := time.Now()
	builder := r.sq.Insert(tablename).
		Columns(
			fullpathColumn,
			createdAtColumn,
			updatedAtColumn,
		).
		Values(
			file.Fullpath,
			ts,
			ts,
		).
		Suffix("RETURNING id")

	sql, args, err := builder.ToSql()
	if err != nil {
		return -1, err
	}

	query := db.Query{
		Name:     "repository.deletion_queue.Create",
		QueryRaw: sql,
	}

	var id int64
	err = r.db.DB().QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", query.Name, err)
	}

	return id, nil
}

func (r *repo) FindByFullpath(ctx context.Context, fullpath string) (*model.Deletion, error) {
	builder := r.sq.
		Select("*").
		From(tablename).
		Where(sq.Eq{fullpathColumn: fullpath}).
		Limit(1)

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	query := db.Query{
		Name:     "repository.deletion_queue.GetByFullpath",
		QueryRaw: sql,
	}

	var file model.Deletion
	err = r.db.DB().QueryRow(ctx, query, args...).Scan(
		&file.Id,
		&file.Fullpath,
		&file.Status,
		&file.ErrorCode,
		&file.CreatedAt,
		&file.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", query.Name, db.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", query.Name, err)
	}

	return &file, nil
}

func (r *repo) FindOldestQueued(ctx context.Context) (*model.Deletion, error) {
	builder := r.sq.
		Select("*").
		From(tablename).
		OrderBy(fmt.Sprintf("%s ASC", updatedAtColumn)).
		Where(
			sq.Eq{statusColumn: model.DeletionStatusPending},
		).
		Limit(1)

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	query := db.Query{
		Name:     "repository.deletion_queue.FindOldestQueued",
		QueryRaw: sql,
	}

	var file model.Deletion
	err = r.db.DB().QueryRow(ctx, query, args...).Scan(
		&file.Id,
		&file.Fullpath,
		&file.Status,
		&file.ErrorCode,
		&file.CreatedAt,
		&file.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", query.Name, db.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", query.Name, err)
	}

	return &file, nil
}

func (r *repo) MarkAsDone(ctx context.Context, fullpath string) error {
	builder := r.sq.
		Update(tablename).
		Set(statusColumn, model.DeletionStatusDone).
		Set(updatedAtColumn, time.Now()).
		Where(sq.Eq{fullpathColumn: fullpath})

	sql, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	query := db.Query{
		Name:     "repository.deletion_queue.MarkAsDone",
		QueryRaw: sql,
	}

	_, err = r.db.DB().Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", query.Name, err)
	}
	return err
}

func (r *repo) MarkAsCanceled(ctx context.Context, fullpath string, code uint32) error {
	builder := r.sq.
		Update(tablename).
		Set(statusColumn, model.DeletionStatusCanceled).
		Set(updatedAtColumn, time.Now()).
		Set(errorCodeColumn, code).
		Where(sq.Eq{fullpathColumn: fullpath})

	sql, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	query := db.Query{
		Name:     "repository.deletion_queue.MarkAsCanceled",
		QueryRaw: sql,
	}

	_, err = r.db.DB().Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", query.Name, err)
	}
	return err
}
