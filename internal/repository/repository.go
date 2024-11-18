package repository

import (
	"context"

	"github.com/chistyakoviv/converter/internal/model"
)

type ConversionQueueRepository interface {
	Create(ctx context.Context, file *model.ConversionInfo) (int64, error)
	GetByFullpath(ctx context.Context, fullpath string) (*model.Conversion, error)
	FindOldestQueued(ctx context.Context) (*model.Conversion, error)
	MarkAsCompleted(ctx context.Context, fullpath string) error
	MarkAsCanceled(ctx context.Context, fullpath string, code uint32) error
}
