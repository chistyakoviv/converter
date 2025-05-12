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
	"github.com/chistyakoviv/converter/internal/service"
)

type serv struct {
	logger                 *slog.Logger
	conversionQueueService service.ConversionQueueService
	deletionQueueService   service.DeletionQueueService
	converterService       service.ConverterService
	conversionQueue        chan struct{}
	deletionQueue          chan struct{}
	done                   chan struct{}
	doneOnce               sync.Once
	mu                     sync.RWMutex
	isScanning             bool
}

/**
* We cannot add a task to the deletion queue while a conversion is in progress,
* because the queue is non-blocking, and if there is no active receiver, the task will be lost.
* To prevent this, use buffered channels to allow tasks to be queued even when there is no active receiver.
**/
func NewService(
	logger *slog.Logger,
	conversionQueueService service.ConversionQueueService,
	deletionQueueService service.DeletionQueueService,
	converterService service.ConverterService,
) service.TaskService {
	return &serv{
		logger:                 logger,
		conversionQueueService: conversionQueueService,
		deletionQueueService:   deletionQueueService,
		converterService:       converterService,
		conversionQueue:        make(chan struct{}, 1),
		deletionQueue:          make(chan struct{}, 1),
		done:                   make(chan struct{}),
	}
}

// Try to add a conversion task only if the queue is not full
func (s *serv) TryQueueConversion() bool {
	// Do not try to schedule handling if the queue is closed
	select {
	case <-s.done:
		return false
	default:
	}

	select {
	case s.conversionQueue <- struct{}{}:
		return true
	default:
		return false
	}
}

// Try to add a deletion task only if the queue is not full
func (s *serv) TryQueueDeletion() bool {
	// Do not try to schedule handling if the queue is closed
	select {
	case <-s.done:
		return false
	default:
	}

	select {
	case s.deletionQueue <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *serv) ProcessQueues(ctx context.Context) {
	for {
		select {
		case <-s.conversionQueue:
			_ = s.processConversion(ctx)
		case <-s.deletionQueue:
			_ = s.processDeletion(ctx)
		case <-ctx.Done():
			s.logger.Info("context done, exiting from task processing")
			s.Shutdown()
			return
		}
	}
}

func (s *serv) processConversion(ctx context.Context) error {
	op := "service.TaskService.ProcessConversion"

	logger := s.logger.With(slog.String("op", op))
	for {
		// It is safe to ask for a task outside a transaction
		// because there is no contention for resources,
		// as the operation is processed in a single thread (monitor goroutine).
		fileInfo, err := s.conversionQueueService.Pop(ctx)
		if errors.Is(err, db.ErrNotFound) {
			return nil
		}
		if err != nil {
			logger.Error("failed to get conversion task", slogger.Err(err))
			return err
		}

		_, err = s.deletionQueueService.Get(ctx, fileInfo.Fullpath)
		if err == nil {
			// Mark the task as canceled if the file is in the deletion queue.
			doneErr := s.conversionQueueService.MarkAsCanceled(ctx, fileInfo.Fullpath, service.ErrFileQueuedForDeletion)
			if doneErr != nil {
				logger.Error("failed to mark conversion task as done", slogger.Err(doneErr))
				return doneErr
			}
			continue
		}
		if !errors.Is(err, db.ErrNotFound) {
			logger.Error("failed to get deletion task while executing conversion task", slogger.Err(err))
			return err
		}

		err = s.converterService.Convert(ctx, fileInfo)
		if err != nil {
			logger.Error("failed to convert file from conversion queue", slogger.Err(err))
			cancelErr := s.conversionQueueService.MarkAsCanceled(ctx, fileInfo.Fullpath, service.GetConverterError(err).Code())
			if cancelErr != nil {
				logger.Error("failed to mark conversion task as canceled", slogger.Err(cancelErr))
				return cancelErr
			}
			continue
		}

		err = s.conversionQueueService.MarkAsDone(ctx, fileInfo.Fullpath)
		if err != nil {
			logger.Error("failed to mark conversion task as done", slogger.Err(err))
			return err
		}
	}
}

func (s *serv) processDeletion(ctx context.Context) error {
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

		fileInfo, err := s.conversionQueueService.Get(ctx, file.Fullpath)
		if errors.Is(err, db.ErrNotFound) {
			// Cancel the task if the file is not in the conversion queue.
			err = s.deletionQueueService.MarkAsCanceled(ctx, file.Fullpath, service.ErrFailedToRemoveFile)
			if err != nil {
				logger.Error("failed to mark deletion task as canceled", slogger.Err(err))
				return err
			}
			continue
		}
		if err != nil {
			logger.Error("failed to get conversion task while executing deletion task", slogger.Err(err))
			return err
		}
		if fileInfo.IsPending() {
			// Mark the task as done if the file is not converted, as there’s no need to delete unconverted files.
			doneErr := s.deletionQueueService.MarkAsDone(ctx, file.Fullpath)
			if doneErr != nil {
				logger.Error("failed to mark deletion task as done", slogger.Err(doneErr))
				return doneErr
			}
			continue
		}

		var removeErrs []error
		for _, entry := range fileInfo.ConvertTo {
			dest, err := fileInfo.AbsoluteDestinationPath(entry)
			if err != nil {
				return err
			}
			// The absence of a file is not considered an error.
			if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
				removeErrs = append(removeErrs, err)
			}
		}
		if len(removeErrs) > 0 {
			// Do not return an error, just mark as canceled
			logger.Error("Failed to remove files from deletion task", slogger.GroupErr(removeErrs))
			err = s.deletionQueueService.MarkAsCanceled(ctx, file.Fullpath, service.ErrFailedToRemoveFile)
			if err != nil {
				logger.Error("failed to mark deletion task as canceled", slogger.Err(err))
				return err
			}
			continue
		}

		err = s.deletionQueueService.MarkAsDone(ctx, fileInfo.Fullpath)
		if err != nil {
			logger.Error("failed to mark deletion task as done", slogger.Err(err))
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

func (s *serv) Shutdown() {
	s.doneOnce.Do(func() {
		close(s.done)
		close(s.conversionQueue)
		close(s.deletionQueue)
	})
}
