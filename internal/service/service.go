package service

import (
	"context"

	"github.com/chistyakoviv/converter/internal/model"
)

type ConversionQueueService interface {
	Add(ctx context.Context, info *model.ConversionInfo) (int64, error)
	Pop(ctx context.Context) (*model.Conversion, error)
	Delete(ctx context.Context, fullpath string) error
	MarkAsCompleted(ctx context.Context, fullpath string) error
	MarkAsCanceled(ctx context.Context, fullpath string, code uint32) error
}

type TaskService interface {
	TrySchedule() bool
	Tasks() <-chan interface{}
	Process(ctx context.Context) error
}

type ConverterService interface {
	Convert(ctx context.Context, info *model.Conversion) error
}
