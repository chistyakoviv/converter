package tests

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	converterMocks "github.com/chistyakoviv/converter/internal/converter/mocks"
	"github.com/chistyakoviv/converter/internal/db"
	"github.com/chistyakoviv/converter/internal/logger/dummy"
	"github.com/chistyakoviv/converter/internal/model"
	"github.com/chistyakoviv/converter/internal/service"
	serviceMocks "github.com/chistyakoviv/converter/internal/service/mocks"
	"github.com/chistyakoviv/converter/internal/service/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaskServiceProcessImageQueues(t *testing.T) {
	testTaskServiceProcessQueues(t, "image")
}

func TestTaskServiceProcessVideoQueues(t *testing.T) {
	testTaskServiceProcessQueues(t, "video")
}

type queueProcessingTestcase struct {
	name                  string
	conversionQeueueLen   int
	deletionQueueLen      int
	fileInfo              *model.Conversion
	deletionInfo          *model.Deletion
	mediaType             string
	started               chan struct{}
	startedOnce           *sync.Once
	mockConversionService func(tc *queueProcessingTestcase) *serviceMocks.MockConversionQueueService
	mockDeletionService   func(tc *queueProcessingTestcase) *serviceMocks.MockDeletionQueueService
	mockConverterService  func(tc *queueProcessingTestcase) *converterMocks.MockConverter
}

func testTaskServiceProcessQueues(t *testing.T, mediaType string) {
	var (
		logger                = dummy.NewDummyLogger()
		conversionPendingInfo = &model.Conversion{
			Id:        1,
			Fullpath:  "/path/to/file.ext",
			Path:      "/path/to",
			Filestem:  "file",
			Ext:       "ext",
			MediaType: mediaTypeOf(mediaType),
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
		conversionDoneInfo = &model.Conversion{
			Id:        1,
			Fullpath:  "/path/to/file.ext",
			Path:      "/path/to",
			Filestem:  "file",
			Ext:       "ext",
			MediaType: mediaTypeOf(mediaType),
			ConvertTo: []model.ConvertTo{
				{
					Ext: "jpg",
				},
			},
			Status:    model.ConversionStatusDone,
			ErrorCode: 0,
			CreatedAt: time.Now(),
			UpdatedAt: sql.NullTime{},
		}
		deletionInfo = &model.Deletion{
			Id:        1,
			Fullpath:  "/path/to/file.ext",
			Status:    model.DeletionStatusPending,
			MediaType: mediaTypeOf(mediaType),
			ErrorCode: 0,
			CreatedAt: time.Now(),
			UpdatedAt: sql.NullTime{},
		}
	)

	cases := []queueProcessingTestcase{
		{
			name: "Empty qeues",
			mockConversionService: func(tc *queueProcessingTestcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				return mockConversionService
			},
			mockDeletionService: func(tc *queueProcessingTestcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				return mockDeletionService
			},
			mockConverterService: func(tc *queueProcessingTestcase) *converterMocks.MockConverter {
				mockConverterService := converterMocks.NewMockConverter(t)
				return mockConverterService
			},
		},
		{
			name:                "No conversion tasks to process",
			conversionQeueueLen: 1,
			mockConversionService: func(tc *queueProcessingTestcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionPop(tc, mockConversionService, nil, db.ErrNotFound)
				return mockConversionService
			},
			mockDeletionService: func(tc *queueProcessingTestcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				return mockDeletionService
			},
			mockConverterService: func(tc *queueProcessingTestcase) *converterMocks.MockConverter {
				mockConverterService := converterMocks.NewMockConverter(t)
				return mockConverterService
			},
		},
		{
			name:                "Unknown error when popping from conversion queue",
			conversionQeueueLen: 1,
			mockConversionService: func(tc *queueProcessingTestcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionPop(tc, mockConversionService, nil, errors.New("unknown error"))
				return mockConversionService
			},
			mockDeletionService: func(tc *queueProcessingTestcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				return mockDeletionService
			},
			mockConverterService: func(tc *queueProcessingTestcase) *converterMocks.MockConverter {
				mockConverterService := converterMocks.NewMockConverter(t)
				return mockConverterService
			},
		},
		{
			name:                "Cancel task enqueued for deletion",
			conversionQeueueLen: 1,
			fileInfo:            conversionPendingInfo,
			deletionInfo:        deletionInfo,
			mockConversionService: func(tc *queueProcessingTestcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionPop(tc, mockConversionService, tc.fileInfo, nil)
				mockConversionService.On("MarkAsCanceled", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo.Fullpath, service.ErrFileQueuedForDeletion).
					Return(nil).
					Once()
				mockConversionPop(tc, mockConversionService, nil, db.ErrNotFound)
				return mockConversionService
			},
			mockDeletionService: func(tc *queueProcessingTestcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionService.On("Get", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo.Fullpath).Return(tc.deletionInfo, nil).Once()
				return mockDeletionService
			},
			mockConverterService: func(tc *queueProcessingTestcase) *converterMocks.MockConverter {
				mockConverterService := converterMocks.NewMockConverter(t)
				return mockConverterService
			},
		},
		{
			name:                "Abort task execution when deletion info retrieval fails",
			conversionQeueueLen: 1,
			fileInfo:            conversionPendingInfo,
			deletionInfo:        deletionInfo,
			mockConversionService: func(tc *queueProcessingTestcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionPop(tc, mockConversionService, tc.fileInfo, nil)
				return mockConversionService
			},
			mockDeletionService: func(tc *queueProcessingTestcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionService.On("Get", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo.Fullpath).Return(tc.deletionInfo, errors.New("unknown error")).Once()
				return mockDeletionService
			},
			mockConverterService: func(tc *queueProcessingTestcase) *converterMocks.MockConverter {
				mockConverterService := converterMocks.NewMockConverter(t)
				return mockConverterService
			},
		},
		{
			name:                "Abort task execution when conversion fails",
			conversionQeueueLen: 1,
			fileInfo:            conversionPendingInfo,
			deletionInfo:        deletionInfo,
			mockConversionService: func(tc *queueProcessingTestcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionPop(tc, mockConversionService, tc.fileInfo, nil)
				mockConversionService.On("MarkAsCanceled", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo.Fullpath, service.ErrUnableToConvertFile).
					Return(nil).
					Once()
				mockConversionPop(tc, mockConversionService, nil, db.ErrNotFound)
				return mockConversionService
			},
			mockDeletionService: func(tc *queueProcessingTestcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionService.On("Get", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo.Fullpath).Return(tc.deletionInfo, db.ErrNotFound).Once()
				return mockDeletionService
			},
			mockConverterService: func(tc *queueProcessingTestcase) *converterMocks.MockConverter {
				mockConverterService := converterMocks.NewMockConverter(t)
				mockConverterService.On("Convert", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo).Return(service.NewConverterError("unknown error", service.ErrUnableToConvertFile)).Once()
				return mockConverterService
			},
		},
		{
			name:                "Successful conversion task execution",
			conversionQeueueLen: 1,
			fileInfo:            conversionPendingInfo,
			deletionInfo:        deletionInfo,
			mockConversionService: func(tc *queueProcessingTestcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionPop(tc, mockConversionService, tc.fileInfo, nil)
				mockConversionService.On("MarkAsDone", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo.Fullpath).
					Return(nil).
					Once()
				mockConversionPop(tc, mockConversionService, nil, db.ErrNotFound)
				return mockConversionService
			},
			mockDeletionService: func(tc *queueProcessingTestcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionService.On("Get", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo.Fullpath).Return(tc.deletionInfo, db.ErrNotFound).Once()
				return mockDeletionService
			},
			mockConverterService: func(tc *queueProcessingTestcase) *converterMocks.MockConverter {
				mockConverterService := converterMocks.NewMockConverter(t)
				mockConverterService.On("Convert", mock.AnythingOfType("*context.cancelCtx"), tc.fileInfo).Return(nil).Once()
				return mockConverterService
			},
		},
		{
			name:             "No deletion tasks to process",
			deletionQueueLen: 1,
			mockConversionService: func(tc *queueProcessingTestcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				return mockConversionService
			},
			mockDeletionService: func(tc *queueProcessingTestcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionPop(tc, mockDeletionService, nil, db.ErrNotFound)
				return mockDeletionService
			},
			mockConverterService: func(tc *queueProcessingTestcase) *converterMocks.MockConverter {
				mockConverterService := converterMocks.NewMockConverter(t)
				return mockConverterService
			},
		},
		{
			name:             "Unknown error when popping from deletion queue",
			deletionQueueLen: 1,
			mockConversionService: func(tc *queueProcessingTestcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				return mockConversionService
			},
			mockDeletionService: func(tc *queueProcessingTestcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionPop(tc, mockDeletionService, nil, errors.New("unknown error"))
				return mockDeletionService
			},
			mockConverterService: func(tc *queueProcessingTestcase) *converterMocks.MockConverter {
				mockConverterService := converterMocks.NewMockConverter(t)
				return mockConverterService
			},
		},
		{
			name:             "Cancel a task that is not present in the conversion queue",
			deletionQueueLen: 1,
			fileInfo:         conversionPendingInfo,
			deletionInfo:     deletionInfo,
			mockConversionService: func(tc *queueProcessingTestcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Get", mock.AnythingOfType("*context.cancelCtx"), tc.deletionInfo.Fullpath).Return(nil, db.ErrNotFound).Once()
				return mockConversionService
			},
			mockDeletionService: func(tc *queueProcessingTestcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionPop(tc, mockDeletionService, tc.deletionInfo, nil)
				mockDeletionService.On("MarkAsCanceled", mock.AnythingOfType("*context.cancelCtx"), tc.deletionInfo.Fullpath, service.ErrFailedToRemoveFile).
					Return(nil).
					Once()
				mockDeletionPop(tc, mockDeletionService, nil, db.ErrNotFound)
				return mockDeletionService
			},
			mockConverterService: func(tc *queueProcessingTestcase) *converterMocks.MockConverter {
				mockConverterService := converterMocks.NewMockConverter(t)
				return mockConverterService
			},
		},
		{
			name:             "Abort task execution when conversion info retrieval fails",
			deletionQueueLen: 1,
			fileInfo:         conversionPendingInfo,
			deletionInfo:     deletionInfo,
			mockConversionService: func(tc *queueProcessingTestcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Get", mock.AnythingOfType("*context.cancelCtx"), tc.deletionInfo.Fullpath).Return(nil, errors.New("unknown error")).Once()
				return mockConversionService
			},
			mockDeletionService: func(tc *queueProcessingTestcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionPop(tc, mockDeletionService, tc.deletionInfo, nil)
				return mockDeletionService
			},
			mockConverterService: func(tc *queueProcessingTestcase) *converterMocks.MockConverter {
				mockConverterService := converterMocks.NewMockConverter(t)
				return mockConverterService
			},
		},
		{
			name:             "Attempt to mark as done a deletion task for a file currently pending conversion",
			deletionQueueLen: 1,
			fileInfo:         conversionPendingInfo,
			deletionInfo:     deletionInfo,
			mockConversionService: func(tc *queueProcessingTestcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Get", mock.AnythingOfType("*context.cancelCtx"), tc.deletionInfo.Fullpath).Return(tc.fileInfo, nil).Once()
				return mockConversionService
			},
			mockDeletionService: func(tc *queueProcessingTestcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionPop(tc, mockDeletionService, tc.deletionInfo, nil)
				mockDeletionService.On("MarkAsDone", mock.AnythingOfType("*context.cancelCtx"), tc.deletionInfo.Fullpath).
					Return(nil).
					Once()
				mockDeletionPop(tc, mockDeletionService, nil, db.ErrNotFound)
				return mockDeletionService
			},
			mockConverterService: func(tc *queueProcessingTestcase) *converterMocks.MockConverter {
				mockConverterService := converterMocks.NewMockConverter(t)
				return mockConverterService
			},
		},
		{
			name:             "Successful deletion task execution",
			deletionQueueLen: 1,
			fileInfo:         conversionDoneInfo,
			deletionInfo:     deletionInfo,
			mockConversionService: func(tc *queueProcessingTestcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Get", mock.AnythingOfType("*context.cancelCtx"), tc.deletionInfo.Fullpath).Return(tc.fileInfo, nil).Once()
				return mockConversionService
			},
			mockDeletionService: func(tc *queueProcessingTestcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionPop(tc, mockDeletionService, tc.deletionInfo, nil)
				mockDeletionService.On("MarkAsDone", mock.AnythingOfType("*context.cancelCtx"), tc.deletionInfo.Fullpath).
					Return(nil).
					Once()
				mockDeletionPop(tc, mockDeletionService, nil, db.ErrNotFound)
				return mockDeletionService
			},
			mockConverterService: func(tc *queueProcessingTestcase) *converterMocks.MockConverter {
				mockConverterService := converterMocks.NewMockConverter(t)
				return mockConverterService
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())

			tc.mediaType = mediaType
			tc.started = make(chan struct{})
			tc.startedOnce = &sync.Once{}

			mockConversionService := tc.mockConversionService(&tc)
			mockDeletionService := tc.mockDeletionService(&tc)
			mockConverterService := tc.mockConverterService(&tc)

			taskService := task.NewService(
				logger,
				mockConversionService,
				mockDeletionService,
				mockConverterService,
			)

			var wg sync.WaitGroup
			var done = make(chan struct{})

			wg.Add(1)
			go func() {
				defer wg.Done()
				close(done)
				processQueues(taskService, mediaType, ctx)
			}()

			// Wait for the goroutine to start
			<-done

			for i := 0; i < tc.conversionQeueueLen; i++ {
				res := enqueueConversion(taskService, mediaType)
				if i > 0 {
					assert.False(t, res)
				} else {
					assert.True(t, res)
				}
			}

			for i := 0; i < tc.deletionQueueLen; i++ {
				res := enqueueDeletion(taskService, mediaType)
				if i > 0 {
					assert.False(t, res)
				} else {
					assert.True(t, res)
				}
			}

			if tc.conversionQeueueLen+tc.deletionQueueLen > 0 {
				// Wait until the task is consumed and its processing has started.
				// Cancelling earlier would race with the worker goroutine's select,
				// which may pick the cancelled context instead of the queued task.
				<-tc.started
			}

			cancel()

			wg.Wait()

			mockConversionService.AssertExpectations(t)
			mockDeletionService.AssertExpectations(t)
			mockConverterService.AssertExpectations(t)
		})
	}
}

func TestTaskServiceProcessQueuesConcurrently(t *testing.T) {
	var (
		logger      = dummy.NewDummyLogger()
		ctx, cancel = context.WithCancel(context.Background())
		videoInfo   = &model.Conversion{
			Id:        1,
			Fullpath:  "/path/to/video.mp4",
			Path:      "/path/to",
			Filestem:  "video",
			Ext:       "mp4",
			MediaType: model.MediaTypeVideo,
			ConvertTo: []model.ConvertTo{
				{
					Ext: "webm",
				},
			},
			Status: model.ConversionStatusPending,
		}
		imageInfo = &model.Conversion{
			Id:        2,
			Fullpath:  "/path/to/image.jpg",
			Path:      "/path/to",
			Filestem:  "image",
			Ext:       "jpg",
			MediaType: model.MediaTypeImage,
			ConvertTo: []model.ConvertTo{
				{
					Ext: "webp",
				},
			},
			Status: model.ConversionStatusPending,
		}
		videoStarted = make(chan struct{})
		imageDone    = make(chan struct{})
		releaseVideo = make(chan struct{})
	)

	mockConversionService := serviceMocks.NewMockConversionQueueService(t)
	mockConversionService.On("PopVideos", mock.AnythingOfType("*context.cancelCtx")).Return(videoInfo, nil).Once()
	mockConversionService.On("PopImages", mock.AnythingOfType("*context.cancelCtx")).Return(imageInfo, nil).Once()
	mockConversionService.On("MarkAsDone", mock.AnythingOfType("*context.cancelCtx"), videoInfo.Fullpath).Return(nil).Once()
	mockConversionService.On("MarkAsDone", mock.AnythingOfType("*context.cancelCtx"), imageInfo.Fullpath).Return(nil).Once()
	mockConversionService.On("PopVideos", mock.AnythingOfType("*context.cancelCtx")).Return(nil, db.ErrNotFound).Once()
	mockConversionService.On("PopImages", mock.AnythingOfType("*context.cancelCtx")).Return(nil, db.ErrNotFound).Once()

	mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
	mockDeletionService.On("Get", mock.AnythingOfType("*context.cancelCtx"), videoInfo.Fullpath).Return(nil, db.ErrNotFound).Once()
	mockDeletionService.On("Get", mock.AnythingOfType("*context.cancelCtx"), imageInfo.Fullpath).Return(nil, db.ErrNotFound).Once()

	mockConverterService := converterMocks.NewMockConverter(t)
	mockConverterService.On("Convert", mock.AnythingOfType("*context.cancelCtx"), videoInfo).
		Run(func(args mock.Arguments) {
			close(videoStarted)
			<-releaseVideo
		}).
		Return(nil).
		Once()
	mockConverterService.On("Convert", mock.AnythingOfType("*context.cancelCtx"), imageInfo).
		Run(func(args mock.Arguments) {
			close(imageDone)
		}).
		Return(nil).
		Once()

	taskService := task.NewService(
		logger,
		mockConversionService,
		mockDeletionService,
		mockConverterService,
	)

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		taskService.ProcessImageQueues(ctx)
	}()
	go func() {
		defer wg.Done()
		taskService.ProcessVideoQueues(ctx)
	}()

	// Trigger both media types
	assert.True(t, taskService.TryQueueVideoConversion())
	assert.True(t, taskService.TryQueueImageConversion())

	// Video conversion is in progress and blocks
	<-videoStarted

	// Image conversion must complete while the video conversion is still in progress
	<-imageDone

	// Release the video conversion
	close(releaseVideo)

	cancel()

	wg.Wait()

	mockConversionService.AssertExpectations(t)
	mockDeletionService.AssertExpectations(t)
	mockConverterService.AssertExpectations(t)
}

func TestTaskServiceProcessScanfs(t *testing.T) {
	var (
		successId int64 = 1
		logger          = dummy.NewDummyLogger()
	)

	type testcase struct {
		name                  string
		mockConversionService func(tc *testcase) *serviceMocks.MockConversionQueueService
		mockDeletionService   func(tc *testcase) *serviceMocks.MockDeletionQueueService
		mockConverterService  func(tc *testcase) *converterMocks.MockConverter
	}

	cases := []testcase{
		{
			name: "Successful scanfs task execution",
			mockConversionService: func(tc *testcase) *serviceMocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.
					On(
						"Add",
						mock.AnythingOfType("*context.cancelCtx"),
						&model.ConversionInfo{
							Fullpath:  "/files/images/gen.jpg",
							Path:      "/files/images",
							Filestem:  "gen",
							Ext:       "jpg",
							MediaType: model.MediaTypeImage,
						},
					).
					Return(successId, nil).
					Once()
				mockConversionService.
					On(
						"Add",
						mock.AnythingOfType("*context.cancelCtx"),
						&model.ConversionInfo{
							Fullpath:  "/files/images/gen.png",
							Path:      "/files/images",
							Filestem:  "gen",
							Ext:       "png",
							MediaType: model.MediaTypeImage,
						},
					).
					Return(successId, nil).
					Once()
				mockConversionService.
					On(
						"Add",
						mock.AnythingOfType("*context.cancelCtx"),
						&model.ConversionInfo{
							Fullpath:  "/files/videos/gen.mp4",
							Path:      "/files/videos",
							Filestem:  "gen",
							Ext:       "mp4",
							MediaType: model.MediaTypeVideo,
						},
					).
					Return(successId, nil).
					Once()
				return mockConversionService
			},
			mockDeletionService: func(tc *testcase) *serviceMocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				return mockDeletionService
			},
			mockConverterService: func(tc *testcase) *converterMocks.MockConverter {
				mockConverterService := converterMocks.NewMockConverter(t)
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

			err := taskService.ProcessScanfs(ctx, "files")

			cancel()

			assert.NoError(t, err)

			mockConversionService.AssertExpectations(t)
			mockDeletionService.AssertExpectations(t)
			mockConverterService.AssertExpectations(t)
		})
	}
}

func mediaTypeOf(mediaType string) int {
	if mediaType == "video" {
		return model.MediaTypeVideo
	}
	return model.MediaTypeImage
}

func enqueueConversion(taskService service.TaskService, mediaType string) bool {
	if mediaType == "video" {
		return taskService.TryQueueVideoConversion()
	}
	return taskService.TryQueueImageConversion()
}

func enqueueDeletion(taskService service.TaskService, mediaType string) bool {
	if mediaType == "video" {
		return taskService.TryQueueVideoDeletion()
	}
	return taskService.TryQueueImageDeletion()
}

func processQueues(taskService service.TaskService, mediaType string, ctx context.Context) {
	if mediaType == "video" {
		taskService.ProcessVideoQueues(ctx)
		return
	}
	taskService.ProcessImageQueues(ctx)
}

func mockConversionPop(
	tc *queueProcessingTestcase,
	mockService *serviceMocks.MockConversionQueueService,
	result *model.Conversion,
	err error,
) {
	method := "PopImages"
	if tc.mediaType == "video" {
		method = "PopVideos"
	}
	mockService.On(method, mock.AnythingOfType("*context.cancelCtx")).
		Run(func(args mock.Arguments) {
			tc.startedOnce.Do(func() { close(tc.started) })
		}).
		Return(result, err).
		Once()
}

func mockDeletionPop(
	tc *queueProcessingTestcase,
	mockService *serviceMocks.MockDeletionQueueService,
	result *model.Deletion,
	err error,
) {
	method := "PopImages"
	if tc.mediaType == "video" {
		method = "PopVideos"
	}
	mockService.On(method, mock.AnythingOfType("*context.cancelCtx")).
		Run(func(args mock.Arguments) {
			tc.startedOnce.Do(func() { close(tc.started) })
		}).
		Return(result, err).
		Once()
}
