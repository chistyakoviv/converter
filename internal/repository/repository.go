package repository

import (
	"context"

	"github.com/chistyakoviv/converter/internal/model"
)

type ConversionQueueRepository interface {
	Create(ctx context.Context, file *model.ConversionInfo) (int64, error)
	FindByFullpath(ctx context.Context, fullpath string) (*model.Conversion, error)
	FindByFilestem(ctx context.Context, filestem string) (*model.Conversion, error)
	FindOldestQueued(ctx context.Context) (*model.Conversion, error)
	MarkAsCompleted(ctx context.Context, fullpath string) error
	MarkAsCanceled(ctx context.Context, fullpath string, code uint32) error
}

type DeletionQueueRepository interface {
	Create(ctx context.Context, file *model.DeletionInfo) (int64, error)
	FindByFullpath(ctx context.Context, fullpath string) (*model.Deletion, error)
	FindOldestQueued(ctx context.Context) (*model.Deletion, error)
	MarkAsCompleted(ctx context.Context, fullpath string) error
	MarkAsCanceled(ctx context.Context, fullpath string, code uint32) error
}
