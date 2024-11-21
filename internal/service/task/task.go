package task

import (
	"context"
	"errors"
	"log/slog"

	"github.com/chistyakoviv/converter/internal/db"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
	"github.com/chistyakoviv/converter/internal/service"
	"github.com/chistyakoviv/converter/internal/service/converter"
)

type serv struct {
	logger                 *slog.Logger
	conversionQueueService service.ConversionQueueService
	converterService       service.ConverterService
	queue                  chan interface{}
}

func NewService(
	logger *slog.Logger,
	conversionQueueService service.ConversionQueueService,
	converterService service.ConverterService,
) service.TaskService {
	return &serv{
		logger:                 logger,
		conversionQueueService: conversionQueueService,
		converterService:       converterService,
		queue:                  make(chan interface{}),
	}
}

// Try to add a task only if the queue is not full
func (s *serv) TrySchedule() bool {
	select {
	case s.queue <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *serv) Tasks() <-chan interface{} {
	return s.queue
}

func (s *serv) Process(ctx context.Context) error {
	for {
		fileInfo, err := s.conversionQueueService.Pop(ctx)
		if errors.Is(err, db.ErrNotFound) {
			return nil
		}

		if err != nil {
			s.logger.Error("failed to get conversion task", slogger.Err(err))
			return err
		}

		err = s.converterService.Convert(ctx, fileInfo)
		if err != nil {
			s.logger.Error("failed to convert file", slogger.Err(err))
			s.conversionQueueService.MarkAsCanceled(ctx, fileInfo.Fullpath, converter.GetConversionError(err).Code())
			continue
		}

		err = s.conversionQueueService.MarkAsCompleted(ctx, fileInfo.Fullpath)
		if err != nil {
			s.logger.Error("failed to mark as completed", slogger.Err(err))
			return err
		}
	}
}
