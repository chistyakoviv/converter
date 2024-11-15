package conversion

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/chistyakoviv/converter/internal/db"
	"github.com/chistyakoviv/converter/internal/model"
	"github.com/chistyakoviv/converter/internal/repository"
	"github.com/chistyakoviv/converter/internal/service"
)

type serv struct {
	txManager            db.TxManager
	conversionRepository repository.ConversionQueueRepository
}

func NewService(
	txManager db.TxManager,
	conversionRepository repository.ConversionQueueRepository,
) service.ConversionService {
	return &serv{
		txManager:            txManager,
		conversionRepository: conversionRepository,
	}
}

func (s *serv) Add(ctx context.Context, info *model.ConversionInfo) (int64, error) {
	if !isSupported(info.Ext) {
		return -1, fmt.Errorf("file type '%s' not supported", info.Ext)
	}

	// Assign default format if no target formats are specified
	if info.ConvertTo == nil {
		defaultFormat, err := defaultFormatFor(info.Ext)
		if err != nil {
			return -1, fmt.Errorf("failed to get default format: %w", err)
		}
		info.ConvertTo = []string{defaultFormat}
	} else {
		var unsupportedFormats []string
		for _, ext := range info.ConvertTo {
			if !isConvertible(info.Ext, ext) {
				unsupportedFormats = append(unsupportedFormats, fmt.Sprintf("'%s'", ext))
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
	err := s.txManager.ReadCommitted(ctx, func(ctx context.Context) error {
		var errTx error
		_, errTx = s.conversionRepository.GetByFullpath(ctx, info.Fullpath)
		if !errors.Is(errTx, db.ErrNotFound) {
			return ErrPathAlreadyExist
		}
		id, errTx = s.conversionRepository.Create(ctx, info)

		return errTx
	})

	if err != nil {
		return -1, err
	}

	return id, nil
}

func (s *serv) Delete(ctx context.Context, fullpath string) error {
	return nil
}
