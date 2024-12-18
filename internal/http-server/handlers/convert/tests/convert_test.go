package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chistyakoviv/converter/internal/http-server/handlers"
	"github.com/chistyakoviv/converter/internal/http-server/handlers/convert"
	handlersMocks "github.com/chistyakoviv/converter/internal/http-server/handlers/mocks"
	"github.com/chistyakoviv/converter/internal/http-server/request"
	"github.com/chistyakoviv/converter/internal/logger/dummy"
	"github.com/chistyakoviv/converter/internal/model"
	"github.com/chistyakoviv/converter/internal/service/conversionq"
	"github.com/chistyakoviv/converter/internal/service/mocks"
	serviceMocks "github.com/chistyakoviv/converter/internal/service/mocks"
)

func TestConvertHandler(t *testing.T) {
	var (
		errorId    int64 = -1
		successId  int64 = 1
		ctx              = context.Background()
		logger           = dummy.NewDummyLogger()
		validation       = validator.New()
		convertTo        = []model.ConvertTo{{Ext: "123", Optional: map[string]interface{}{"replace_orig_ext": true}, ConvConf: map[string]interface{}{"quality": float64(100)}}}
	)

	type testcase struct {
		name                  string
		input                 string
		respError             string
		statusCode            int
		conversionInfo        *model.ConversionInfo
		conversionReq         *request.ConversionRequest
		mockValidator         func(tc *testcase) handlers.Validator
		mockConversionService func(tc *testcase) *mocks.MockConversionQueueService
		mockTaskService       func(tc *testcase) *mocks.MockTaskService
	}

	cases := []testcase{
		{
			name:       "Incorrect request: empty data",
			input:      "",
			respError:  "empty request",
			statusCode: http.StatusBadRequest,
			mockValidator: func(tc *testcase) handlers.Validator {
				mockValidator := handlersMocks.NewMockValidator(t)
				return mockValidator
			},
			mockConversionService: func(tc *testcase) *mocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				return mockConversionService
			},
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				return mockTaskService
			},
		},
		{
			name:       "Incorrect request: invalid data",
			input:      `{"fullpath": "/path/to/file.ext"}`,
			respError:  "field Path is a required field",
			statusCode: http.StatusBadRequest,
			mockValidator: func(tc *testcase) handlers.Validator {
				return validation
			},
			mockConversionService: func(tc *testcase) *mocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				return mockConversionService
			},
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				return mockTaskService
			},
		},
		{
			name:           "Incorrect request: path duplicate",
			input:          `{"path": "/path/to/file.ext"}`,
			respError:      "file with the specified path already exists in the conversion queue",
			statusCode:     http.StatusConflict,
			conversionInfo: &model.ConversionInfo{Fullpath: "/path/to/file.ext", Path: "/path/to", Filestem: "file", Ext: "ext"},
			conversionReq:  &request.ConversionRequest{Path: "/path/to/file.ext"},
			mockValidator: func(tc *testcase) handlers.Validator {
				mockValidator := handlersMocks.NewMockValidator(t)
				mockValidator.On("Struct", tc.conversionReq).Return(nil).Once()
				return mockValidator
			},
			mockConversionService: func(tc *testcase) *mocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Add", ctx, tc.conversionInfo).Return(errorId, conversionq.ErrPathAlreadyExist).Once()
				return mockConversionService
			},
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				return mockTaskService
			},
		},
		{
			name:           "Incorrect request: non-existent file",
			input:          `{"path": "/path/to/file.ext"}`,
			respError:      "file does not exist",
			statusCode:     http.StatusNotFound,
			conversionInfo: &model.ConversionInfo{Fullpath: "/path/to/file.ext", Path: "/path/to", Filestem: "file", Ext: "ext"},
			conversionReq:  &request.ConversionRequest{Path: "/path/to/file.ext"},
			mockValidator: func(tc *testcase) handlers.Validator {
				mockValidator := handlersMocks.NewMockValidator(t)
				mockValidator.On("Struct", tc.conversionReq).Return(nil).Once()
				return mockValidator
			},
			mockConversionService: func(tc *testcase) *mocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Add", ctx, tc.conversionInfo).Return(errorId, conversionq.ErrFileDoesNotExist).Once()
				return mockConversionService
			},
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				return mockTaskService
			},
		},
		{
			name:           "Incorrect request: unsupported extension",
			input:          `{"path": "/path/to/file.ext"}`,
			respError:      "file type not supported",
			statusCode:     http.StatusBadRequest,
			conversionInfo: &model.ConversionInfo{Fullpath: "/path/to/file.ext", Path: "/path/to", Filestem: "file", Ext: "ext"},
			conversionReq:  &request.ConversionRequest{Path: "/path/to/file.ext"},
			mockValidator: func(tc *testcase) handlers.Validator {
				mockValidator := handlersMocks.NewMockValidator(t)
				mockValidator.On("Struct", tc.conversionReq).Return(nil).Once()
				return mockValidator
			},
			mockConversionService: func(tc *testcase) *mocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Add", ctx, tc.conversionInfo).Return(errorId, conversionq.ErrFileTypeNotSupported).Once()
				return mockConversionService
			},
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				return mockTaskService
			},
		},
		{
			name:           "Incorrect request: unknown file type",
			input:          `{"path": "/path/to/file.ext"}`,
			respError:      "failed to determine file type",
			statusCode:     http.StatusUnprocessableEntity,
			conversionInfo: &model.ConversionInfo{Fullpath: "/path/to/file.ext", Path: "/path/to", Filestem: "file", Ext: "ext"},
			conversionReq:  &request.ConversionRequest{Path: "/path/to/file.ext"},
			mockValidator: func(tc *testcase) handlers.Validator {
				mockValidator := handlersMocks.NewMockValidator(t)
				mockValidator.On("Struct", tc.conversionReq).Return(nil).Once()
				return mockValidator
			},
			mockConversionService: func(tc *testcase) *mocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Add", ctx, tc.conversionInfo).Return(errorId, conversionq.ErrFailedDetermineFileType).Once()
				return mockConversionService
			},
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				return mockTaskService
			},
		},
		{
			name:           "Incorrect request: wrong conversion format",
			input:          `{"path": "/path/to/file.ext", "convert_to": [{"ext": "123", "optional": {"replace_orig_ext": true}, "conv_conf": {"quality": 100}}]}`,
			respError:      "cannot convert to the specified format",
			statusCode:     http.StatusBadRequest,
			conversionInfo: &model.ConversionInfo{Fullpath: "/path/to/file.ext", Path: "/path/to", Filestem: "file", Ext: "ext", ConvertTo: convertTo},
			conversionReq:  &request.ConversionRequest{Path: "/path/to/file.ext", ConvertTo: convertTo},
			mockValidator: func(tc *testcase) handlers.Validator {
				mockValidator := handlersMocks.NewMockValidator(t)
				mockValidator.On("Struct", tc.conversionReq).Return(nil).Once()
				return mockValidator
			},
			mockConversionService: func(tc *testcase) *mocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Add", ctx, tc.conversionInfo).Return(errorId, conversionq.ErrInvalidConversionFormat).Once()
				return mockConversionService
			},
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				return mockTaskService
			},
		},
		{
			name:           "Incorrect request: no target formats specified",
			input:          `{"path": "/path/to/file.ext"}`,
			respError:      "target format list is empty",
			statusCode:     http.StatusBadRequest,
			conversionInfo: &model.ConversionInfo{Fullpath: "/path/to/file.ext", Path: "/path/to", Filestem: "file", Ext: "ext"},
			conversionReq:  &request.ConversionRequest{Path: "/path/to/file.ext"},
			mockValidator: func(tc *testcase) handlers.Validator {
				mockValidator := handlersMocks.NewMockValidator(t)
				mockValidator.On("Struct", tc.conversionReq).Return(nil).Once()
				return mockValidator
			},
			mockConversionService: func(tc *testcase) *mocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Add", ctx, tc.conversionInfo).Return(errorId, conversionq.ErrEmptyTargetFormatList).Once()
				return mockConversionService
			},
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				return mockTaskService
			},
		},
		{
			name:           "Incorrect request: unknown error",
			input:          `{"path": "/path/to/file.ext"}`,
			respError:      "failed to add file to conversion queue",
			statusCode:     http.StatusInternalServerError,
			conversionInfo: &model.ConversionInfo{Fullpath: "/path/to/file.ext", Path: "/path/to", Filestem: "file", Ext: "ext"},
			conversionReq:  &request.ConversionRequest{Path: "/path/to/file.ext"},
			mockValidator: func(tc *testcase) handlers.Validator {
				mockValidator := handlersMocks.NewMockValidator(t)
				mockValidator.On("Struct", tc.conversionReq).Return(nil).Once()
				return mockValidator
			},
			mockConversionService: func(tc *testcase) *mocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Add", ctx, tc.conversionInfo).Return(errorId, errors.New("Unknown error")).Once()
				return mockConversionService
			},
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				return mockTaskService
			},
		},
		{
			name:           "Successful request",
			input:          `{"path": "/path/to/file.ext"}`,
			respError:      "",
			statusCode:     http.StatusOK,
			conversionInfo: &model.ConversionInfo{Fullpath: "/path/to/file.ext", Path: "/path/to", Filestem: "file", Ext: "ext"},
			conversionReq:  &request.ConversionRequest{Path: "/path/to/file.ext"},
			mockValidator: func(tc *testcase) handlers.Validator {
				mockValidator := handlersMocks.NewMockValidator(t)
				mockValidator.On("Struct", tc.conversionReq).Return(nil).Once()
				return mockValidator
			},
			mockConversionService: func(tc *testcase) *mocks.MockConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.On("Add", ctx, tc.conversionInfo).Return(successId, nil).Once()
				return mockConversionService
			},
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				mockTaskService.On("TryQueueConversion").Return(true).Once()
				return mockTaskService
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockValidator := tc.mockValidator(&tc)
			mockConversionService := tc.mockConversionService(&tc)
			mockTaskService := tc.mockTaskService(&tc)

			handler := convert.New(
				ctx,
				logger,
				mockValidator,
				mockConversionService,
				mockTaskService,
			)
			req, err := http.NewRequest(http.MethodPost, "/convert", bytes.NewReader([]byte(tc.input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			body := rr.Body.String()

			var resp convert.ConversionResponse

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			assert.Equal(t, tc.respError, resp.Error)
			assert.Equal(t, tc.statusCode, rr.Result().StatusCode)
			mockConversionService.AssertExpectations(t)
			mockTaskService.AssertExpectations(t)
		})
	}
}
