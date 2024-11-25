package conversionq

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/db"
	"github.com/chistyakoviv/converter/internal/file"
	"github.com/chistyakoviv/converter/internal/model"
	"github.com/chistyakoviv/converter/internal/repository"
	"github.com/chistyakoviv/converter/internal/service"
)

type serv struct {
	cfg                  *config.Config
	txManager            db.TxManager
	conversionRepository repository.ConversionQueueRepository
}

func NewService(
	cfg *config.Config,
	txManager db.TxManager,
	conversionRepository repository.ConversionQueueRepository,
) service.ConversionQueueService {
	return &serv{
		cfg:                  cfg,
		txManager:            txManager,
		conversionRepository: conversionRepository,
	}
}

func (s *serv) Add(ctx context.Context, info *model.ConversionInfo) (int64, error) {
	src, err := info.AbsoluteSourcePath()
	if err != nil {
		return -1, err
	}
	if !file.Exists(src) {
		return -1, fmt.Errorf("file '%s' does not exist", info.Fullpath)
	}
	if !isSupported(info.Ext) {
		return -1, fmt.Errorf("file type '%s' not supported", info.Ext)
	}

	// Assign default format if no target formats are specified
	if info.ConvertTo == nil {
		var err error
		var ok bool

		if ok, err = file.IsImage(info.Fullpath); ok {
			info.ConvertTo = s.cfg.Image.DefaultFormats
		}
		if err != nil {
			return -1, fmt.Errorf("failed to determine file type: %w", err)
		}
		if ok, err = file.IsVideo(info.Fullpath); ok {
			info.ConvertTo = s.cfg.Video.DefaultFormats
		}
		if err != nil {
			return -1, fmt.Errorf("failed to determine file type: %w", err)
		}
	} else {
		var unsupportedFormats []string
		for _, entry := range info.ConvertTo {
			if !isConvertible(info.Ext, entry.Ext) {
				unsupportedFormats = append(unsupportedFormats, fmt.Sprintf("'%s'", entry.Ext))
			}
		}
		if len(unsupportedFormats) > 0 {
			return -1, fmt.Errorf("file type '%s' is not convertible to %s", info.Ext, strings.Join(unsupportedFormats, ", "))
		}
	}

	var id int64

	// Since it's not possible to preemptively check if a query violates constraints,
	// use a transaction to first verify that `fullpath` does not already exist.
	// If `fullpath` exists, return an appropriate error. Otherwise, proceed to
	// insert a new row within the same transaction.
	err = s.txManager.ReadCommitted(ctx, func(ctx context.Context) error {
		var errTx error
		_, errTx = s.conversionRepository.FindByFullpath(ctx, info.Fullpath)
		if !errors.Is(errTx, db.ErrNotFound) {
			if errTx == nil {
				return ErrPathAlreadyExist
			}
			return errTx
		}
		id, errTx = s.conversionRepository.Create(ctx, info)

		return errTx
	})

	if err != nil {
		return -1, err
	}

	return id, nil
}

func (s *serv) Pop(ctx context.Context) (*model.Conversion, error) {
	return s.conversionRepository.FindOldestQueued(ctx)
}

func (s *serv) MarkAsCompleted(ctx context.Context, fullpath string) error {
	return s.conversionRepository.MarkAsCompleted(ctx, fullpath)
}

func (s *serv) MarkAsCanceled(ctx context.Context, fullpath string, code uint32) error {
	return s.conversionRepository.MarkAsCanceled(ctx, fullpath, code)
}
