package conversion

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/chistyakoviv/converter/internal/db"
	"github.com/chistyakoviv/converter/internal/model"
	"github.com/chistyakoviv/converter/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

const (
	tablename = "conversion_queue"

	idColumn             = "id"
	fullpathColumn       = "fullpath"
	pathColumn           = "path"
	filestemColumn       = "filestem"
	extColumn            = "ext"
	convertToColumn      = "convert_to"
	isDoneColumn         = "is_done"
	isCanceledColumn     = "is_canceled"
	replaceOrigExtColumn = "replace_orig_ext"
	errorCodeColumn      = "error_code"
	createdAtColumn      = "created_at"
	updatedAtColumn      = "updated_at"
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

func (r *repo) Create(ctx context.Context, file *model.Conversion) (int64, error) {
	builder := r.sq.Insert(tablename).
		Columns(idColumn, fullpathColumn, pathColumn, filestemColumn, extColumn, convertToColumn, isDoneColumn, isCanceledColumn, replaceOrigExtColumn, errorCodeColumn, createdAtColumn, updatedAtColumn).
		Values(file.Id, file.Fullpath, file.Path, file.Filestem, file.Ext, file.ConvertTo, file.IsDone, file.IsCanceled, file.ReplaceOrigExt, file.ErrorCode, file.CreatedAt, file.UpdatedAt).
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

func (r *repo) GetByFullpath(ctx context.Context, fullpath string) (*model.Conversion, error) {
	builder := r.sq.Select("*").From(tablename).Where(sq.Eq{fullpathColumn: fullpath}).Limit(1)

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	query := db.Query{
		Name:     "repository.conversion_queue.GetByFullpath",
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
		&file.ReplaceOrigExt,
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
