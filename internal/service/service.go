package service

import (
	"context"

	"github.com/chistyakoviv/converter/internal/model"
)

type ConversionService interface {
	Add(ctx context.Context, info *model.ConversionInfo) (int64, error)
	Delete(ctx context.Context, fullpath string) error
}

type TaskService interface {
	TrySchedule() bool
	Tasks() <-chan interface{}
}
