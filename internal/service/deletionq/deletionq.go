package deletionq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/db"
	"github.com/chistyakoviv/converter/internal/model"
	"github.com/chistyakoviv/converter/internal/repository"
	"github.com/chistyakoviv/converter/internal/service"
)

type serv struct {
	cfg                  *config.Config
	logger               *slog.Logger
	txManager            db.TxManager
	deletionRepository   repository.DeletionQueueRepository
	conversionRepository repository.ConversionQueueRepository
}

func NewService(
	cfg *config.Config,
	logger *slog.Logger,
	txManager db.TxManager,
	deletionRepository repository.DeletionQueueRepository,
	conversionRepository repository.ConversionQueueRepository,
) service.DeletionQueueService {
	return &serv{
		cfg:                  cfg,
		logger:               logger,
		txManager:            txManager,
		deletionRepository:   deletionRepository,
		conversionRepository: conversionRepository,
	}
}

func (s *serv) Add(ctx context.Context, info *model.DeletionInfo) (int64, error) {
	// Skip checking if the file exists, as the source file might already be deleted when attempting removal.
	var id int64

	err := s.txManager.ReadCommitted(ctx, func(ctx context.Context) error {
		var errTx error
		_, errTx = s.conversionRepository.FindByFullpath(ctx, info.Fullpath)
		if errors.Is(errTx, db.ErrNotFound) {
			return fmt.Errorf("deletion failed for '%s': %w", info.Fullpath, ErrFileDoesNotExist)
		}
		if errTx != nil {
			return errTx
		}
		_, errTx = s.deletionRepository.FindByFullpath(ctx, info.Fullpath)
		// Return an error if the file is found (== nil) in the deletion queue
		if errTx == nil {
			return fmt.Errorf("deletion failed for '%s': %w", info.Fullpath, ErrPathAlreadyExist)
		}
		// Return an error if it is not the NotFound error
		if !errors.Is(errTx, db.ErrNotFound) {
			return errTx
		}
		id, errTx = s.deletionRepository.Create(ctx, info)

		return errTx
	})

	if err != nil {
		return -1, err
	}

	return id, nil
}

func (s *serv) Pop(ctx context.Context) (*model.Deletion, error) {
	return s.deletionRepository.FindOldestQueued(ctx)
}

func (s *serv) MarkAsDone(ctx context.Context, fullpath string) error {
	return s.deletionRepository.MarkAsDone(ctx, fullpath)
}

func (s *serv) MarkAsCanceled(ctx context.Context, fullpath string, code uint32) error {
	return s.deletionRepository.MarkAsCanceled(ctx, fullpath, code)
}
