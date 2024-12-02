package task

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/chistyakoviv/converter/internal/db"
	"github.com/chistyakoviv/converter/internal/file"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
	"github.com/chistyakoviv/converter/internal/model"
	"github.com/chistyakoviv/converter/internal/repository"
	"github.com/chistyakoviv/converter/internal/service"
	"github.com/chistyakoviv/converter/internal/service/converter"
	"github.com/chistyakoviv/converter/internal/service/deletionq"
)

type serv struct {
	logger                 *slog.Logger
	conversionQueueService service.ConversionQueueService
	deletionQueueService   service.DeletionQueueService
	converterService       service.ConverterService
	conversionRepository   repository.ConversionQueueRepository
	conversionQueue        chan interface{}
	deletionQueue          chan interface{}
	mu                     sync.RWMutex
	isScanning             bool
}

func NewService(
	logger *slog.Logger,
	conversionQueueService service.ConversionQueueService,
	deletionQueueService service.DeletionQueueService,
	converterService service.ConverterService,
	conversionRepository repository.ConversionQueueRepository,
) service.TaskService {
	return &serv{
		logger:                 logger,
		conversionQueueService: conversionQueueService,
		deletionQueueService:   deletionQueueService,
		converterService:       converterService,
		conversionRepository:   conversionRepository,
		conversionQueue:        make(chan interface{}),
		deletionQueue:          make(chan interface{}),
	}
}

// Try to add a conversion task only if the queue is not full
func (s *serv) TryQueueConversion() bool {
	select {
	case s.conversionQueue <- struct{}{}:
		return true
	default:
		return false
	}
}

// Try to add a deletion task only if the queue is not full
func (s *serv) TryQueueDeletion() bool {
	select {
	case s.deletionQueue <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *serv) ProcessConversion(ctx context.Context) {
	for range s.conversionQueue {
		s.processConversion(ctx)
	}
}

func (s *serv) processConversion(ctx context.Context) error {
	op := "service.TaskService.ProcessConversion"

	logger := s.logger.With(slog.String("op", op))
	for {
		// It is safe to ask for a task outside a transaction
		// because there is no contention for resources,
		// as the operation is processed in a single thread.
		fileInfo, err := s.conversionQueueService.Pop(ctx)
		if errors.Is(err, db.ErrNotFound) {
			return nil
		}

		if err != nil {
			logger.Error("failed to get conversion task", slogger.Err(err))
			return err
		}

		err = s.converterService.Convert(ctx, fileInfo)
		if err != nil {
			logger.Error("failed to convert file", slogger.Err(err))
			err = s.conversionQueueService.MarkAsCanceled(ctx, fileInfo.Fullpath, converter.GetConversionError(err).Code())
			if err != nil {
				logger.Error("failed to mark as canceled", slogger.Err(err))
				return err
			}
			continue
		}

		err = s.conversionQueueService.MarkAsDone(ctx, fileInfo.Fullpath)
		if err != nil {
			logger.Error("failed to mark as completed", slogger.Err(err))
			return err
		}
	}
}

func (s *serv) ProcessDeletion(ctx context.Context) {
	for range s.deletionQueue {
		s.processDeletions(ctx)
	}
}

func (s *serv) processDeletions(ctx context.Context) error {
	op := "service.TaskService.ProcessDeletion"

	logger := s.logger.With(slog.String("op", op))
	for {
		file, err := s.deletionQueueService.Pop(ctx)
		if errors.Is(err, db.ErrNotFound) {
			return nil
		}

		if err != nil {
			logger.Error("failed to get deletion task", slogger.Err(err))
			return err
		}

		fileInfo, err := s.conversionRepository.FindByFullpath(ctx, file.Fullpath)
		if err != nil {
			return err
		}
		var removeErrs []error
		for _, entry := range fileInfo.ConvertTo {
			dest, err := fileInfo.AbsoluteDestinationPath(entry.Ext)
			if err != nil {
				return err
			}
			if err := os.Remove(dest); err != nil {
				removeErrs = append(removeErrs, err)
			}
		}

		if len(removeErrs) > 0 {
			// Do not return an error, just mark as canceled
			logger.Error("Failed to remove files", slogger.GroupErr(removeErrs))
			err = s.deletionQueueService.MarkAsCanceled(ctx, fileInfo.Fullpath, deletionq.ErrFailedToRemoveFile)
			if err != nil {
				logger.Error("failed to mark as canceled", slogger.Err(err))
				return err
			}
			continue
		}

		err = s.deletionQueueService.MarkAsDone(ctx, fileInfo.Fullpath)
		if err != nil {
			logger.Error("failed to mark as completed", slogger.Err(err))
			return err
		}
	}
}

func (s *serv) IsScanning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isScanning
}

func (s *serv) ProcessScanfs(ctx context.Context, rootDir string) error {
	s.mu.Lock()
	if s.isScanning {
		s.mu.Unlock()
		return fmt.Errorf("another scan is already in progress: %w", ErrScanAlreadyRunning)
	}

	s.isScanning = true
	s.mu.Unlock()

	// Walk through the directory
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			s.logger.Error("Error accessing path", slogger.Err(err))
			return nil
		}

		// Paths must start with "/"
		path = file.EnsureLeadingSlash(path)

		// Perform enqueuing if the file is not a directory
		if !d.IsDir() {
			s.logger.Debug("Try to enqueue file", slog.String("path", path))

			imageOk, filetypeErr := file.IsImage(path)
			if filetypeErr != nil {
				s.logger.Error("failed to determine image type", slogger.Err(filetypeErr))
				return nil
			}
			videoOk, filetypeErr := file.IsVideo(path)
			if filetypeErr != nil {
				s.logger.Error("failed to determine video type", slogger.Err(filetypeErr))
				return nil
			}

			if imageOk || videoOk {
				src, err := file.Trimwd(path)
				if err != nil {
					s.logger.Error("failed to trim working directory", slogger.Err(err))
					return nil
				}
				finfo := file.ExtractInfo(src)
				_, err = s.conversionQueueService.Add(ctx, model.ToConversionInfoFromFileInfo(finfo))
				if err != nil {
					s.logger.Error("failed to enqueue conversion while scanning filesystem", slogger.Err(err))
				}
			}
		}

		return nil
	})

	s.mu.Lock()
	s.isScanning = false
	s.mu.Unlock()

	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	return nil
}
