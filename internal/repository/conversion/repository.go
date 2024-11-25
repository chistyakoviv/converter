package conversion

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
	tablename = "conversion_queue"

	idColumn         = "id"
	fullpathColumn   = "fullpath"
	pathColumn       = "path"
	filestemColumn   = "filestem"
	extColumn        = "ext"
	convertToColumn  = "convert_to"
	isDoneColumn     = "is_done"
	isCanceledColumn = "is_canceled"
	errorCodeColumn  = "error_code"
	createdAtColumn  = "created_at"
	updatedAtColumn  = "updated_at"
)

type repo struct {
	db db.Client
	sq sq.StatementBuilderType
}

func NewRepository(db db.Client, sq sq.StatementBuilderType) repository.ConversionQueueRepository {
	return &repo{
		db: db,
		sq: sq,
	}
}

func (r *repo) Create(ctx context.Context, file *model.ConversionInfo) (int64, error) {
	ts := time.Now()
	builder := r.sq.Insert(tablename).
		Columns(
			fullpathColumn,
			pathColumn,
			filestemColumn,
			extColumn,
			convertToColumn,
			createdAtColumn,
			updatedAtColumn,
		).
		Values(
			file.Fullpath,
			file.Path,
			file.Filestem,
			file.Ext,
			file.ConvertTo,
			ts,
			ts,
		).
		Suffix("RETURNING id")

	sql, args, err := builder.ToSql()
	if err != nil {
		return -1, err
	}

	query := db.Query{
		Name:     "repository.conversion_queue.Create",
		QueryRaw: sql,
	}

	var id int64
	err = r.db.DB().QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", query.Name, err)
	}

	return id, nil
}

func (r *repo) FindByFullpath(ctx context.Context, fullpath string) (*model.Conversion, error) {
	builder := r.sq.Select("*").From(tablename).Where(sq.Eq{fullpathColumn: fullpath}).Limit(1)

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	query := db.Query{
		Name:     "repository.conversion_queue.FindByFullpath",
		QueryRaw: sql,
	}

	var file model.Conversion
	err = r.db.DB().QueryRow(ctx, query, args...).Scan(
		&file.Id,
		&file.Fullpath,
		&file.Path,
		&file.Filestem,
		&file.Ext,
		&file.ConvertTo,
		&file.IsDone,
		&file.IsCanceled,
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

func (r *repo) FindOldestQueued(ctx context.Context) (*model.Conversion, error) {
	builder := r.sq.
		Select("*").
		From(tablename).
		OrderBy(fmt.Sprintf("%s DESC", updatedAtColumn)).
		Where(
			sq.And{
				sq.Eq{isDoneColumn: false},
				sq.Eq{isCanceledColumn: false},
			},
		).
		Limit(1)

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	query := db.Query{
		Name:     "repository.conversion_queue.FindOldestQueued",
		QueryRaw: sql,
	}

	var file model.Conversion
	err = r.db.DB().QueryRow(ctx, query, args...).Scan(
		&file.Id,
		&file.Fullpath,
		&file.Path,
		&file.Filestem,
		&file.Ext,
		&file.ConvertTo,
		&file.IsDone,
		&file.IsCanceled,
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

func (r *repo) MarkAsCompleted(ctx context.Context, fullpath string) error {
	builder := r.sq.
		Update(tablename).
		Set(isDoneColumn, true).
		Set(updatedAtColumn, time.Now()).
		Where(sq.Eq{fullpathColumn: fullpath})

	sql, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	query := db.Query{
		Name:     "repository.conversion_queue.MarkAsCompleted",
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
		Set(isCanceledColumn, true).
		Set(updatedAtColumn, time.Now()).
		Set(errorCodeColumn, code).
		Where(sq.Eq{fullpathColumn: fullpath})

	sql, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	query := db.Query{
		Name:     "repository.conversion_queue.MarkAsCanceled",
		QueryRaw: sql,
	}

	_, err = r.db.DB().Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", query.Name, err)
	}
	return err
}
