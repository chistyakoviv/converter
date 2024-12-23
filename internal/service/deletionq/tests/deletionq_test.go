package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/db"
	dbMocks "github.com/chistyakoviv/converter/internal/db/mocks"
	"github.com/chistyakoviv/converter/internal/logger/dummy"
	"github.com/chistyakoviv/converter/internal/model"
	repositoryMocks "github.com/chistyakoviv/converter/internal/repository/mocks"
	"github.com/chistyakoviv/converter/internal/service"
	"github.com/chistyakoviv/converter/internal/service/deletionq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	ctx          = context.Background()
	configPath   = "config/local.yaml"
	defaultsPath = "config/defaults.yaml"
)

func TestAddToConversionQueue(t *testing.T) {
	var (
		logger          = dummy.NewDummyLogger()
		errorId   int64 = -1
		successId int64 = 0
	)

	type testcase struct {
		name                     string
		err                      string
		id                       int64
		deletionInfo             *model.DeletionInfo
		mockDeletionRepository   func(tc *testcase) *repositoryMocks.MockDeletionQueueRepository
		mockConversionRepository func(tc *testcase) *repositoryMocks.MockConversionQueueRepository
		mockTxManager            func(tc *testcase) *dbMocks.MockTxManager
	}

	cases := []testcase{
		{
			name: "Failed to add to deletion queue",
			err:  "Any error",
			id:   errorId,
			deletionInfo: &model.DeletionInfo{
				Fullpath: "/files/images/gen.jpg",
			},
			mockDeletionRepository: func(tc *testcase) *repositoryMocks.MockDeletionQueueRepository {
				mockDeletionRepository := repositoryMocks.NewMockDeletionQueueRepository(t)
				return mockDeletionRepository
			},
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				mockTxManager.On("ReadCommitted", mock.AnythingOfType("context.backgroundCtx"), mock.Anything).Return(errors.New("Any error"))
				return mockTxManager
			},
		},
		{
			name: "Successful add to deletion queue",
			id:   successId,
			deletionInfo: &model.DeletionInfo{
				Fullpath: "/files/images/gen.jpg",
			},
			mockDeletionRepository: func(tc *testcase) *repositoryMocks.MockDeletionQueueRepository {
				mockDeletionRepository := repositoryMocks.NewMockDeletionQueueRepository(t)
				return mockDeletionRepository
			},
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				mockTxManager.On("ReadCommitted", mock.AnythingOfType("context.backgroundCtx"), mock.Anything).Return(nil)
				return mockTxManager
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockDeletionRepository := tc.mockDeletionRepository(&tc)
			mockConversionRepository := tc.mockConversionRepository(&tc)
			mockTxManager := tc.mockTxManager(&tc)

			serv := deletionq.NewService(
				config.MustLoad(configPath, defaultsPath),
				logger,
				mockTxManager,
				mockDeletionRepository,
				mockConversionRepository,
			)

			id, err := serv.Add(ctx, tc.deletionInfo)

			if tc.err != "" {
				assert.Equal(t, err.Error(), tc.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.id, id)
			}

			mockDeletionRepository.AssertExpectations(t)
			mockConversionRepository.AssertExpectations(t)
			mockTxManager.AssertExpectations(t)
		})
	}
}

func TestPopFromDeletionQueue(t *testing.T) {
	var (
		logger   = dummy.NewDummyLogger()
		deletion = &model.Deletion{
			Fullpath: "/files/images/gen.jpg",
		}
	)

	type testcase struct {
		name                     string
		err                      error
		deletion                 *model.Deletion
		mockDeletionRepository   func(tc *testcase) *repositoryMocks.MockDeletionQueueRepository
		mockConversionRepository func(tc *testcase) *repositoryMocks.MockConversionQueueRepository
		mockTxManager            func(tc *testcase) *dbMocks.MockTxManager
	}

	cases := []testcase{
		{
			name:     "Empty deletion queue",
			err:      db.ErrNotFound,
			deletion: deletion,
			mockDeletionRepository: func(tc *testcase) *repositoryMocks.MockDeletionQueueRepository {
				mockDeletionRepository := repositoryMocks.NewMockDeletionQueueRepository(t)
				mockDeletionRepository.On("FindOldestQueued", mock.AnythingOfType("context.backgroundCtx")).Return(nil, db.ErrNotFound)
				return mockDeletionRepository
			},
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
		{
			name:     "Successful pop from deletion queue",
			deletion: deletion,
			mockDeletionRepository: func(tc *testcase) *repositoryMocks.MockDeletionQueueRepository {
				mockDeletionRepository := repositoryMocks.NewMockDeletionQueueRepository(t)
				mockDeletionRepository.On("FindOldestQueued", mock.AnythingOfType("context.backgroundCtx")).Return(deletion, nil)
				return mockDeletionRepository
			},
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockDeletionRepository := tc.mockDeletionRepository(&tc)
			mockConversionRepository := tc.mockConversionRepository(&tc)
			mockTxManager := tc.mockTxManager(&tc)

			serv := deletionq.NewService(
				config.MustLoad(configPath, defaultsPath),
				logger,
				mockTxManager,
				mockDeletionRepository,
				mockConversionRepository,
			)

			conversion, err := serv.Pop(ctx)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, conversion, tc.deletion)
			}

			mockDeletionRepository.AssertExpectations(t)
			mockConversionRepository.AssertExpectations(t)
			mockTxManager.AssertExpectations(t)
		})
	}
}

func TestGetFromDeletionQueue(t *testing.T) {
	var (
		logger   = dummy.NewDummyLogger()
		deletion = &model.Deletion{
			Fullpath: "/files/images/gen.jpg",
		}
	)

	type testcase struct {
		name                     string
		err                      error
		path                     string
		deletion                 *model.Deletion
		mockDeletionRepository   func(tc *testcase) *repositoryMocks.MockDeletionQueueRepository
		mockConversionRepository func(tc *testcase) *repositoryMocks.MockConversionQueueRepository
		mockTxManager            func(tc *testcase) *dbMocks.MockTxManager
	}

	cases := []testcase{
		{
			name:     "Item not found in deletion queue",
			err:      db.ErrNotFound,
			path:     "/path/to/file.ext",
			deletion: deletion,
			mockDeletionRepository: func(tc *testcase) *repositoryMocks.MockDeletionQueueRepository {
				mockDeletionRepository := repositoryMocks.NewMockDeletionQueueRepository(t)
				mockDeletionRepository.On("FindByFullpath", mock.AnythingOfType("context.backgroundCtx"), tc.path).Return(nil, db.ErrNotFound)
				return mockDeletionRepository
			},
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
		{
			name:     "Successful get from deletion queue",
			path:     "/path/to/file.ext",
			deletion: deletion,
			mockDeletionRepository: func(tc *testcase) *repositoryMocks.MockDeletionQueueRepository {
				mockDeletionRepository := repositoryMocks.NewMockDeletionQueueRepository(t)
				mockDeletionRepository.On("FindByFullpath", mock.AnythingOfType("context.backgroundCtx"), tc.path).Return(deletion, nil)
				return mockDeletionRepository
			},
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockDeletionRepository := tc.mockDeletionRepository(&tc)
			mockConversionRepository := tc.mockConversionRepository(&tc)
			mockTxManager := tc.mockTxManager(&tc)

			serv := deletionq.NewService(
				config.MustLoad(configPath, defaultsPath),
				logger,
				mockTxManager,
				mockDeletionRepository,
				mockConversionRepository,
			)

			deletion, err := serv.Get(ctx, tc.path)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, deletion, tc.deletion)
			}

			mockDeletionRepository.AssertExpectations(t)
			mockConversionRepository.AssertExpectations(t)
			mockTxManager.AssertExpectations(t)
		})
	}
}

func TestMarkAsDoneForDeletionQueue(t *testing.T) {
	var (
		logger = dummy.NewDummyLogger()
	)

	type testcase struct {
		name                     string
		err                      error
		path                     string
		mockDeletionRepository   func(tc *testcase) *repositoryMocks.MockDeletionQueueRepository
		mockConversionRepository func(tc *testcase) *repositoryMocks.MockConversionQueueRepository
		mockTxManager            func(tc *testcase) *dbMocks.MockTxManager
	}

	cases := []testcase{
		{
			name: "Item not found in deletion queue",
			err:  db.ErrNotFound,
			path: "/path/to/file.ext",
			mockDeletionRepository: func(tc *testcase) *repositoryMocks.MockDeletionQueueRepository {
				mockDeletionRepository := repositoryMocks.NewMockDeletionQueueRepository(t)
				mockDeletionRepository.On("MarkAsDone", mock.AnythingOfType("context.backgroundCtx"), tc.path).Return(db.ErrNotFound)
				return mockDeletionRepository
			},
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
		{
			name: "Successful mark as done for deletion queue",
			path: "/path/to/file.ext",
			mockDeletionRepository: func(tc *testcase) *repositoryMocks.MockDeletionQueueRepository {
				mockDeletionRepository := repositoryMocks.NewMockDeletionQueueRepository(t)
				mockDeletionRepository.On("MarkAsDone", mock.AnythingOfType("context.backgroundCtx"), tc.path).Return(nil)
				return mockDeletionRepository
			},
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockDeletionRepository := tc.mockDeletionRepository(&tc)
			mockConversionRepository := tc.mockConversionRepository(&tc)
			mockTxManager := tc.mockTxManager(&tc)

			serv := deletionq.NewService(
				config.MustLoad(configPath, defaultsPath),
				logger,
				mockTxManager,
				mockDeletionRepository,
				mockConversionRepository,
			)

			err := serv.MarkAsDone(ctx, tc.path)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockDeletionRepository.AssertExpectations(t)
			mockConversionRepository.AssertExpectations(t)
			mockTxManager.AssertExpectations(t)
		})
	}
}

func TestMarkAsCanceledForDeletionQueue(t *testing.T) {
	var (
		logger = dummy.NewDummyLogger()
	)

	type testcase struct {
		name                     string
		err                      error
		path                     string
		code                     uint32
		mockDeletionRepository   func(tc *testcase) *repositoryMocks.MockDeletionQueueRepository
		mockConversionRepository func(tc *testcase) *repositoryMocks.MockConversionQueueRepository
		mockTxManager            func(tc *testcase) *dbMocks.MockTxManager
	}

	cases := []testcase{
		{
			name: "Item not found in deletion queue",
			err:  db.ErrNotFound,
			path: "/path/to/file.ext",
			code: service.ErrFileDoesNotExist,
			mockDeletionRepository: func(tc *testcase) *repositoryMocks.MockDeletionQueueRepository {
				mockDeletionRepository := repositoryMocks.NewMockDeletionQueueRepository(t)
				mockDeletionRepository.On("MarkAsCanceled", mock.AnythingOfType("context.backgroundCtx"), tc.path, tc.code).Return(db.ErrNotFound)
				return mockDeletionRepository
			},
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
		{
			name: "Successful mark as canceled for deletion queue",
			path: "/path/to/file.ext",
			code: service.ErrFileDoesNotExist,
			mockDeletionRepository: func(tc *testcase) *repositoryMocks.MockDeletionQueueRepository {
				mockDeletionRepository := repositoryMocks.NewMockDeletionQueueRepository(t)
				mockDeletionRepository.On("MarkAsCanceled", mock.AnythingOfType("context.backgroundCtx"), tc.path, tc.code).Return(nil)
				return mockDeletionRepository
			},
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockDeletionRepository := tc.mockDeletionRepository(&tc)
			mockConversionRepository := tc.mockConversionRepository(&tc)
			mockTxManager := tc.mockTxManager(&tc)

			serv := deletionq.NewService(
				config.MustLoad(configPath, defaultsPath),
				logger,
				mockTxManager,
				mockDeletionRepository,
				mockConversionRepository,
			)

			err := serv.MarkAsCanceled(ctx, tc.path, tc.code)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockDeletionRepository.AssertExpectations(t)
			mockConversionRepository.AssertExpectations(t)
			mockTxManager.AssertExpectations(t)
		})
	}
}
