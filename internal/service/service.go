package service

import (
	"context"

	"github.com/chistyakoviv/converter/internal/model"
)

type ConversionQueueService interface {
	Add(ctx context.Context, info *model.ConversionInfo) (int64, error)
	Pop(ctx context.Context) (*model.Conversion, error)
	MarkAsCompleted(ctx context.Context, fullpath string) error
	MarkAsCanceled(ctx context.Context, fullpath string, code uint32) error
}

type DeletionQueueService interface {
	Add(ctx context.Context, info *model.DeletionInfo) (int64, error)
	Pop(ctx context.Context) (*model.Deletion, error)
	MarkAsCompleted(ctx context.Context, fullpath string) error
	MarkAsCanceled(ctx context.Context, fullpath string, code uint32) error
}

type TaskService interface {
	TryQueueConversion() bool
	TryQueueDeletion() bool
	ProcessConversion(ctx context.Context)
	ProcessDeletion(ctx context.Context)
}

type ConverterService interface {
	Convert(ctx context.Context, info *model.Conversion) error
}
