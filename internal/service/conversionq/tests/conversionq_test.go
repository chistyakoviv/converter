package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/chistyakoviv/converter/internal/config"
	dbMocks "github.com/chistyakoviv/converter/internal/db/mocks"
	"github.com/chistyakoviv/converter/internal/model"
	repositoryMocks "github.com/chistyakoviv/converter/internal/repository/mocks"
	"github.com/chistyakoviv/converter/internal/service/conversionq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddToConversionQueue(t *testing.T) {
	var (
		ctx                     = context.Background()
		errorId           int64 = -1
		successId         int64 = 0
		configPath              = "/go/src/github.com/chistyakoviv/converter/config/local.yaml"
		defaultsPath            = "/go/src/github.com/chistyakoviv/converter/config/defaults.yaml"
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
		// pngConversionInfo = &model.ConversionInfo{
		// 	Fullpath: "/files/images/gen.png",
		// 	Path:     "/files/images",
		// 	Filestem: "gen",
		// 	Ext:      "png",
		// }
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
			tc.mockConversionRepository(&tc)

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
