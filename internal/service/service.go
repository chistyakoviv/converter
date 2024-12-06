package service

import (
	"context"

	"github.com/chistyakoviv/converter/internal/model"
)

type ConversionQueueService interface {
	Add(ctx context.Context, info *model.ConversionInfo) (int64, error)
	Pop(ctx context.Context) (*model.Conversion, error)
	MarkAsDone(ctx context.Context, fullpath string) error
	MarkAsCanceled(ctx context.Context, fullpath string, code uint32) error
}

type DeletionQueueService interface {
	Add(ctx context.Context, info *model.DeletionInfo) (int64, error)
	Pop(ctx context.Context) (*model.Deletion, error)
	MarkAsDone(ctx context.Context, fullpath string) error
	MarkAsCanceled(ctx context.Context, fullpath string, code uint32) error
}

type TaskService interface {
	TryQueueConversion() bool
	TryQueueDeletion() bool
	ProcessQueues(ctx context.Context)
	// ProcessConversion(ctx context.Context)
	// ProcessDeletion(ctx context.Context)
	ProcessScanfs(ctx context.Context, rootDir string) error
	IsScanning() bool
}

type ConverterService interface {
	Convert(ctx context.Context, info *model.Conversion) error
}
