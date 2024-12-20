package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/db"
	dbMocks "github.com/chistyakoviv/converter/internal/db/mocks"
	"github.com/chistyakoviv/converter/internal/model"
	repositoryMocks "github.com/chistyakoviv/converter/internal/repository/mocks"
	"github.com/chistyakoviv/converter/internal/service"
	"github.com/chistyakoviv/converter/internal/service/conversionq"
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
		errorId           int64 = -1
		successId         int64 = 0
		defaultCfg              = config.MustLoad(configPath, defaultsPath)
		jpgConversionInfo       = func() *model.ConversionInfo {
			// Generate models inside a function, because the conversion queue modifies the model passed to the Add method
			return &model.ConversionInfo{
				Fullpath: "/files/images/gen.jpg",
				Path:     "/files/images",
				Filestem: "gen",
				Ext:      "jpg",
			}
		}
		absentFileConversionInfo = func() *model.ConversionInfo {
			return &model.ConversionInfo{
				Fullpath: "/files/absent.jpg",
				Path:     "/files",
				Filestem: "absent",
				Ext:      "jpg",
			}
		}
		txtConversionInfo = func() *model.ConversionInfo {
			return &model.ConversionInfo{
				Fullpath: "/files/other/test.txt",
				Path:     "/files/other",
				Filestem: "test",
				Ext:      "txt",
			}
		}
		mp4ConversionInfo = func() *model.ConversionInfo {
			return &model.ConversionInfo{
				Fullpath: "/files/videos/gen.mp4",
				Path:     "/files/videos",
				Filestem: "gen",
				Ext:      "mp4",
			}
		}
	)

	type testcase struct {
		name                     string
		err                      string
		id                       int64
		configPath               string
		defaultsPath             string
		conversionInfo           *model.ConversionInfo
		convertTo                []model.ConvertTo
		mockConversionRepository func(tc *testcase) *repositoryMocks.MockConversionQueueRepository
		mockTxManager            func(tc *testcase) *dbMocks.MockTxManager
	}

	cases := []testcase{
		{
			name:           "Add non-existing file to conversion queue",
			id:             errorId,
			err:            fmt.Sprintf("%s: %s", absentFileConversionInfo().Fullpath, conversionq.ErrFileDoesNotExist),
			configPath:     configPath,
			defaultsPath:   defaultsPath,
			conversionInfo: absentFileConversionInfo(),
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				// All repository methods are invoked within a transaction as an anonymous function,
				//making it impossible to directly verify the calls
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
		{
			name:           "Add file with unsupported extension to conversion queue",
			id:             errorId,
			err:            fmt.Sprintf("%s: %s", txtConversionInfo().Ext, conversionq.ErrFileTypeNotSupported.Error()),
			configPath:     configPath,
			defaultsPath:   defaultsPath,
			conversionInfo: txtConversionInfo(),
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				// All repository methods are invoked within a transaction as an anonymous function,
				//making it impossible to directly verify the calls
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
		{
			name:           "Don't allow empty conversion targets",
			id:             errorId,
			err:            fmt.Sprintf("target formats not specified: %s", conversionq.ErrEmptyTargetFormatList.Error()),
			configPath:     configPath,
			defaultsPath:   "config/empty_defaults.yaml",
			conversionInfo: jpgConversionInfo(),
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				// All repository methods are invoked within a transaction as an anonymous function,
				//making it impossible to directly verify the calls
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
		{
			name:           "Don't allow wrong conversion targets",
			id:             errorId,
			err:            fmt.Sprintf("conversion from '%s' to 'ext', 'txt': %s", jpgConversionInfo().Ext, conversionq.ErrInvalidConversionFormat.Error()),
			configPath:     configPath,
			defaultsPath:   "config/wrong_defaults.yaml",
			conversionInfo: jpgConversionInfo(),
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				// All repository methods are invoked within a transaction as an anonymous function,
				//making it impossible to directly verify the calls
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
		{
			name:           "Check loading default conversion targets for images",
			id:             successId,
			configPath:     configPath,
			defaultsPath:   defaultsPath,
			conversionInfo: jpgConversionInfo(),
			convertTo:      defaultCfg.Defaults.Image.Formats,
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				// All repository methods are invoked within a transaction as an anonymous function,
				//making it impossible to directly verify the calls
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				mockTxManager.On("ReadCommitted", mock.AnythingOfType("context.backgroundCtx"), mock.Anything).Return(nil)
				return mockTxManager
			},
		},
		{
			name:           "Check loading default conversion targets for videos",
			id:             successId,
			configPath:     configPath,
			defaultsPath:   defaultsPath,
			conversionInfo: mp4ConversionInfo(),
			convertTo:      defaultCfg.Defaults.Video.Formats,
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				// All repository methods are invoked within a transaction as an anonymous function,
				//making it impossible to directly verify the calls
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				mockTxManager.On("ReadCommitted", mock.AnythingOfType("context.backgroundCtx"), mock.Anything).Return(nil)
				return mockTxManager
			},
		},
		{
			name:           "Successful addition of an image to conversion queue",
			id:             successId,
			configPath:     configPath,
			defaultsPath:   defaultsPath,
			conversionInfo: jpgConversionInfo(),
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				// All repository methods are invoked within a transaction as an anonymous function,
				//making it impossible to directly verify the calls
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

			mockConversionRepository := tc.mockConversionRepository(&tc)
			mockTxManager := tc.mockTxManager(&tc)

			serv := conversionq.NewService(
				config.MustLoad(tc.configPath, tc.defaultsPath),
				mockTxManager,
				mockConversionRepository,
			)

			id, err := serv.Add(ctx, tc.conversionInfo)

			if tc.err != "" {
				assert.Equal(t, err.Error(), tc.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.id, id)
			}

			if tc.convertTo != nil {
				assert.Equal(t, tc.convertTo, tc.conversionInfo.ConvertTo)
			}

			mockConversionRepository.AssertExpectations(t)
			mockTxManager.AssertExpectations(t)
		})
	}
}

func TestPopFromConversionQueue(t *testing.T) {
	var (
		conversion = &model.Conversion{
			Fullpath: "/files/images/gen.jpg",
			Path:     "/files/images",
			Filestem: "gen",
			Ext:      "jpg",
		}
	)

	type testcase struct {
		name                     string
		err                      error
		conversion               *model.Conversion
		mockConversionRepository func(tc *testcase) *repositoryMocks.MockConversionQueueRepository
		mockTxManager            func(tc *testcase) *dbMocks.MockTxManager
	}

	cases := []testcase{
		{
			name:       "Empty conversion queue",
			err:        db.ErrNotFound,
			conversion: conversion,
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				mockConversionRepository.On("FindOldestQueued", mock.AnythingOfType("context.backgroundCtx")).Return(nil, db.ErrNotFound)
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
		{
			name:       "Successful pop from conversion queue",
			conversion: conversion,
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				mockConversionRepository.On("FindOldestQueued", mock.AnythingOfType("context.backgroundCtx")).Return(conversion, nil)
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

			mockConversionRepository := tc.mockConversionRepository(&tc)
			mockTxManager := tc.mockTxManager(&tc)

			serv := conversionq.NewService(
				config.MustLoad(configPath, defaultsPath),
				mockTxManager,
				mockConversionRepository,
			)

			conversion, err := serv.Pop(ctx)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, conversion, tc.conversion)
			}

			mockConversionRepository.AssertExpectations(t)
			mockTxManager.AssertExpectations(t)
		})
	}
}

func TestGetFromConversionQueue(t *testing.T) {
	var (
		conversion = &model.Conversion{
			Fullpath: "/files/images/gen.jpg",
			Path:     "/files/images",
			Filestem: "gen",
			Ext:      "jpg",
		}
	)

	type testcase struct {
		name                     string
		err                      error
		path                     string
		conversion               *model.Conversion
		mockConversionRepository func(tc *testcase) *repositoryMocks.MockConversionQueueRepository
		mockTxManager            func(tc *testcase) *dbMocks.MockTxManager
	}

	cases := []testcase{
		{
			name:       "Item not found in conversion queue",
			err:        db.ErrNotFound,
			path:       "/path/to/file.ext",
			conversion: conversion,
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				mockConversionRepository.On("FindByFullpath", mock.AnythingOfType("context.backgroundCtx"), tc.path).Return(nil, db.ErrNotFound)
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
		{
			name:       "Successful get from conversion queue",
			path:       "/path/to/file.ext",
			conversion: conversion,
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				mockConversionRepository.On("FindByFullpath", mock.AnythingOfType("context.backgroundCtx"), tc.path).Return(conversion, nil)
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

			mockConversionRepository := tc.mockConversionRepository(&tc)
			mockTxManager := tc.mockTxManager(&tc)

			serv := conversionq.NewService(
				config.MustLoad(configPath, defaultsPath),
				mockTxManager,
				mockConversionRepository,
			)

			conversion, err := serv.Get(ctx, tc.path)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, conversion, tc.conversion)
			}

			mockConversionRepository.AssertExpectations(t)
			mockTxManager.AssertExpectations(t)
		})
	}
}

func TestMarkAsDoneForConversionQueue(t *testing.T) {
	type testcase struct {
		name                     string
		err                      error
		path                     string
		mockConversionRepository func(tc *testcase) *repositoryMocks.MockConversionQueueRepository
		mockTxManager            func(tc *testcase) *dbMocks.MockTxManager
	}

	cases := []testcase{
		{
			name: "Item not found in conversion queue",
			err:  db.ErrNotFound,
			path: "/path/to/file.ext",
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				mockConversionRepository.On("MarkAsDone", mock.AnythingOfType("context.backgroundCtx"), tc.path).Return(db.ErrNotFound)
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
		{
			name: "Successful mark as done for conversion queue",
			path: "/path/to/file.ext",
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				mockConversionRepository.On("MarkAsDone", mock.AnythingOfType("context.backgroundCtx"), tc.path).Return(nil)
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

			mockConversionRepository := tc.mockConversionRepository(&tc)
			mockTxManager := tc.mockTxManager(&tc)

			serv := conversionq.NewService(
				config.MustLoad(configPath, defaultsPath),
				mockTxManager,
				mockConversionRepository,
			)

			err := serv.MarkAsDone(ctx, tc.path)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockConversionRepository.AssertExpectations(t)
			mockTxManager.AssertExpectations(t)
		})
	}
}

func TestMarkAsCanceledForConversionQueue(t *testing.T) {
	type testcase struct {
		name                     string
		err                      error
		path                     string
		code                     uint32
		mockConversionRepository func(tc *testcase) *repositoryMocks.MockConversionQueueRepository
		mockTxManager            func(tc *testcase) *dbMocks.MockTxManager
	}

	cases := []testcase{
		{
			name: "Item not found in conversion queue",
			err:  db.ErrNotFound,
			path: "/path/to/file.ext",
			code: service.ErrFileDoesNotExist,
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				mockConversionRepository.On("MarkAsCanceled", mock.AnythingOfType("context.backgroundCtx"), tc.path, tc.code).Return(db.ErrNotFound)
				return mockConversionRepository
			},
			mockTxManager: func(tc *testcase) *dbMocks.MockTxManager {
				mockTxManager := dbMocks.NewMockTxManager(t)
				return mockTxManager
			},
		},
		{
			name: "Successful mark as canceled for conversion queue",
			path: "/path/to/file.ext",
			code: service.ErrFileDoesNotExist,
			mockConversionRepository: func(tc *testcase) *repositoryMocks.MockConversionQueueRepository {
				mockConversionRepository := repositoryMocks.NewMockConversionQueueRepository(t)
				mockConversionRepository.On("MarkAsCanceled", mock.AnythingOfType("context.backgroundCtx"), tc.path, tc.code).Return(nil)
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

			mockConversionRepository := tc.mockConversionRepository(&tc)
			mockTxManager := tc.mockTxManager(&tc)

			serv := conversionq.NewService(
				config.MustLoad(configPath, defaultsPath),
				mockTxManager,
				mockConversionRepository,
			)

			err := serv.MarkAsCanceled(ctx, tc.path, tc.code)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockConversionRepository.AssertExpectations(t)
			mockTxManager.AssertExpectations(t)
		})
	}
}
