package tests

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/chistyakoviv/converter/internal/db"
	"github.com/chistyakoviv/converter/internal/logger/dummy"
	"github.com/chistyakoviv/converter/internal/model"
	"github.com/chistyakoviv/converter/internal/service"
	serviceMocks "github.com/chistyakoviv/converter/internal/service/mocks"
	"github.com/chistyakoviv/converter/internal/service/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaskService(t *testing.T) {
	var (
		logger   = dummy.NewDummyLogger()
		fileInfo = &model.Conversion{
			Id:       1,
			Fullpath: "/path/to/file.ext",
			Path:     "/path/to",
			Filestem: "file",
			Ext:      "ext",
			ConvertTo: []model.ConvertTo{
				{
					Ext: "jpg",
				},
			},
			Status:    model.ConversionStatusPending,
			ErrorCode: 0,
			CreatedAt: time.Now(),
			UpdatedAt: sql.NullTime{},
		}
		deletionInfo = &model.Deletion{
			Id:        1,
			Fullpath:  "/path/to/file.ext",
			Status:    model.ConversionStatusPending,
			ErrorCode: 0,
			CreatedAt: time.Now(),
			UpdatedAt: sql.NullTime{},
		}
	)

	type testcase struct {
		name                  string
		conversionQeueueLen   int
		deletionQueueLen      int
		fileInfo              *model.Conversion
		deletionInfo          *model.Deletion
		mockConversionService func(tc *testcase) *serviceMocks.MockConversionQueueService
		mockDeletionService   func(tc *testcase) *serviceMocks.MockDeletionQueueService
		mockConverterService  func(tc *testcase) *serviceMocks.MockConverterService
	}

	cases := []testcase{
		{
			name: "Empty qeues",
			mockConversionService: func(tc *testcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				return mockConversionService
			},
			mockDeletionService: func(tc *testcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				return mockDeletionService
			},
			mockConverterService: func(tc *testcase) *serviceMocks.MockConverterService {
				mockConverterService := serviceMocks.NewMockConverterService(t)
				return mockConverterService
			},
		},
		{
			name:                "No tasks to process",
			conversionQeueueLen: 1,
			mockConversionService: func(tc *testcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Pop", mock.AnythingOfType("*context.cancelCtx")).Return(nil, db.ErrNotFound).Once()
				return mockConversionService
			},
			mockDeletionService: func(tc *testcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				return mockDeletionService
			},
			mockConverterService: func(tc *testcase) *serviceMocks.MockConverterService {
				mockConverterService := serviceMocks.NewMockConverterService(t)
				return mockConverterService
			},
		},
		{
			name:                "Unknown error when popping from conversion queue",
			conversionQeueueLen: 1,
			mockConversionService: func(tc *testcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Pop", mock.AnythingOfType("*context.cancelCtx")).Return(nil, errors.New("unknown error")).Once()
				return mockConversionService
			},
			mockDeletionService: func(tc *testcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				return mockDeletionService
			},
			mockConverterService: func(tc *testcase) *serviceMocks.MockConverterService {
				mockConverterService := serviceMocks.NewMockConverterService(t)
				return mockConverterService
			},
		},
		{
			name:                "Cancel task enqueued for deletion",
			conversionQeueueLen: 1,
			fileInfo:            fileInfo,
			deletionInfo:        deletionInfo,
			mockConversionService: func(tc *testcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Pop", mock.AnythingOfType("*context.cancelCtx")).Return(tc.fileInfo, nil).Once()
				mockConversionService.On("MarkAsCanceled", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo.Fullpath, service.ErrFileQueuedForDeletion).
					Return(errors.New("simulate error to prevent next iteration")).
					Once()
				return mockConversionService
			},
			mockDeletionService: func(tc *testcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionService.On("Get", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo.Fullpath).Return(tc.deletionInfo, nil).Once()
				return mockDeletionService
			},
			mockConverterService: func(tc *testcase) *serviceMocks.MockConverterService {
				mockConverterService := serviceMocks.NewMockConverterService(t)
				return mockConverterService
			},
		},
		{
			name:                "Abort task execution when deletion info retrieval fails",
			conversionQeueueLen: 1,
			fileInfo:            fileInfo,
			deletionInfo:        deletionInfo,
			mockConversionService: func(tc *testcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Pop", mock.AnythingOfType("*context.cancelCtx")).Return(tc.fileInfo, nil).Once()
				return mockConversionService
			},
			mockDeletionService: func(tc *testcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionService.On("Get", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo.Fullpath).Return(tc.deletionInfo, errors.New("unknown error")).Once()
				return mockDeletionService
			},
			mockConverterService: func(tc *testcase) *serviceMocks.MockConverterService {
				mockConverterService := serviceMocks.NewMockConverterService(t)
				return mockConverterService
			},
		},
		{
			name:                "Abort task execution when conversion fails",
			conversionQeueueLen: 1,
			fileInfo:            fileInfo,
			deletionInfo:        deletionInfo,
			mockConversionService: func(tc *testcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Pop", mock.AnythingOfType("*context.cancelCtx")).Return(tc.fileInfo, nil).Once()
				mockConversionService.On("MarkAsCanceled", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo.Fullpath, service.ErrUnableToConvertFile).
					Return(errors.New("simulate error to prevent next iteration")).
					Once()
				return mockConversionService
			},
			mockDeletionService: func(tc *testcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionService.On("Get", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo.Fullpath).Return(tc.deletionInfo, db.ErrNotFound).Once()
				return mockDeletionService
			},
			mockConverterService: func(tc *testcase) *serviceMocks.MockConverterService {
				mockConverterService := serviceMocks.NewMockConverterService(t)
				mockConverterService.On("Convert", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo).Return(service.NewConverterError("unknown error", service.ErrUnableToConvertFile)).Once()
				return mockConverterService
			},
		},
		{
			name:                "Successful conversion task execution",
			conversionQeueueLen: 1,
			fileInfo:            fileInfo,
			deletionInfo:        deletionInfo,
			mockConversionService: func(tc *testcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Pop", mock.AnythingOfType("*context.cancelCtx")).Return(tc.fileInfo, nil).Once()
				mockConversionService.On("MarkAsDone", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo.Fullpath).
					Return(errors.New("simulate error to prevent next iteration")).
					Once()
				return mockConversionService
			},
			mockDeletionService: func(tc *testcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionService.On("Get", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo.Fullpath).Return(tc.deletionInfo, db.ErrNotFound).Once()
				return mockDeletionService
			},
			mockConverterService: func(tc *testcase) *serviceMocks.MockConverterService {
				mockConverterService := serviceMocks.NewMockConverterService(t)
				mockConverterService.On("Convert", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo).Return(nil).Once()
				return mockConverterService
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())

			mockConversionService := tc.mockConversionService(&tc)
			mockDeletionService := tc.mockDeletionService(&tc)
			mockConverterService := tc.mockConverterService(&tc)

			taskService := task.NewService(
				logger,
				mockConversionService,
				mockDeletionService,
				mockConverterService,
			)

			for i := 0; i < tc.conversionQeueueLen; i++ {
				res := taskService.TryQueueConversion()
				if i > 0 {
					assert.False(t, res)
				} else {
					assert.True(t, res)
				}
			}

			for i := 0; i < tc.deletionQueueLen; i++ {
				res := taskService.TryQueueDeletion()
				if i > 0 {
					assert.False(t, res)
				} else {
					assert.True(t, res)
				}
			}

			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				defer wg.Done()
				taskService.ProcessQueues(ctx)
			}()

			time.Sleep(100 * time.Millisecond)

			cancel()

			wg.Wait()

			mockConversionService.AssertExpectations(t)
			mockDeletionService.AssertExpectations(t)
			mockConverterService.AssertExpectations(t)
		})
	}
}
