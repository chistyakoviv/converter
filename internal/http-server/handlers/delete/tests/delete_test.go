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
	"github.com/chistyakoviv/converter/internal/http-server/handlers/delete"
	handlersMocks "github.com/chistyakoviv/converter/internal/http-server/handlers/mocks"
	"github.com/chistyakoviv/converter/internal/http-server/request"
	"github.com/chistyakoviv/converter/internal/logger/dummy"
	"github.com/chistyakoviv/converter/internal/model"
	"github.com/chistyakoviv/converter/internal/service/deletionq"
	"github.com/chistyakoviv/converter/internal/service/mocks"
	serviceMocks "github.com/chistyakoviv/converter/internal/service/mocks"
)

func TestDeleteHandler(t *testing.T) {
	var (
		errorId    int64 = -1
		successId  int64 = 1
		ctx              = context.Background()
		logger           = dummy.NewDummyLogger()
		validation       = validator.New()
	)

	type testcase struct {
		name                string
		input               string
		respError           string
		statusCode          int
		deletionInfo        *model.DeletionInfo
		deletionReq         *request.DeletionRequest
		mockValidator       func(tc *testcase) handlers.Validator
		mockDeletionService func(tc *testcase) *mocks.MockDeletionQueueService
		mockTaskService     func(tc *testcase) *mocks.MockTaskService
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
			mockDeletionService: func(tc *testcase) *mocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				return mockDeletionService
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
			mockDeletionService: func(tc *testcase) *mocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				return mockDeletionService
			},
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				return mockTaskService
			},
		},
		{
			name:         "Incorrect request: path duplicate",
			input:        `{"path": "/path/to/file.ext"}`,
			respError:    "file with the specified path already exists in the deletion queue",
			statusCode:   http.StatusConflict,
			deletionInfo: &model.DeletionInfo{Fullpath: "/path/to/file.ext"},
			deletionReq:  &request.DeletionRequest{Path: "/path/to/file.ext"},
			mockValidator: func(tc *testcase) handlers.Validator {
				mockValidator := handlersMocks.NewMockValidator(t)
				mockValidator.On("Struct", tc.deletionReq).Return(nil).Once()
				return mockValidator
			},
			mockDeletionService: func(tc *testcase) *mocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionService.On("Add", ctx, tc.deletionInfo).Return(errorId, deletionq.ErrPathAlreadyExist).Once()
				return mockDeletionService
			},
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				return mockTaskService
			},
		},
		{
			name:         "Incorrect request: non-existent file",
			input:        `{"path": "/path/to/file.ext"}`,
			respError:    "file does not exist",
			statusCode:   http.StatusNotFound,
			deletionInfo: &model.DeletionInfo{Fullpath: "/path/to/file.ext"},
			deletionReq:  &request.DeletionRequest{Path: "/path/to/file.ext"},
			mockValidator: func(tc *testcase) handlers.Validator {
				mockValidator := handlersMocks.NewMockValidator(t)
				mockValidator.On("Struct", tc.deletionReq).Return(nil).Once()
				return mockValidator
			},
			mockDeletionService: func(tc *testcase) *mocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionService.On("Add", ctx, tc.deletionInfo).Return(errorId, deletionq.ErrFileDoesNotExist).Once()
				return mockDeletionService
			},
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				return mockTaskService
			},
		},
		{
			name:         "Incorrect request: unknown error",
			input:        `{"path": "/path/to/file.ext"}`,
			respError:    "failed to add file to deletion queue",
			statusCode:   http.StatusInternalServerError,
			deletionInfo: &model.DeletionInfo{Fullpath: "/path/to/file.ext"},
			deletionReq:  &request.DeletionRequest{Path: "/path/to/file.ext"},
			mockValidator: func(tc *testcase) handlers.Validator {
				mockValidator := handlersMocks.NewMockValidator(t)
				mockValidator.On("Struct", tc.deletionReq).Return(nil).Once()
				return mockValidator
			},
			mockDeletionService: func(tc *testcase) *mocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionService.On("Add", ctx, tc.deletionInfo).Return(errorId, errors.New("unknown error")).Once()
				return mockDeletionService
			},
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				return mockTaskService
			},
		},
		{
			name:         "Successful request",
			input:        `{"path": "/path/to/file.ext"}`,
			respError:    "",
			statusCode:   http.StatusOK,
			deletionInfo: &model.DeletionInfo{Fullpath: "/path/to/file.ext"},
			deletionReq:  &request.DeletionRequest{Path: "/path/to/file.ext"},
			mockValidator: func(tc *testcase) handlers.Validator {
				mockValidator := handlersMocks.NewMockValidator(t)
				mockValidator.On("Struct", tc.deletionReq).Return(nil).Once()
				return mockValidator
			},
			mockDeletionService: func(tc *testcase) *mocks.MockDeletionQueueService {
				mockDeletionService := serviceMocks.NewMockDeletionQueueService(t)
				mockDeletionService.On("Add", ctx, tc.deletionInfo).Return(successId, nil).Once()
				return mockDeletionService
			},
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				mockTaskService.On("TryQueueDeletion").Return(true).Once()
				return mockTaskService
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockValidator := tc.mockValidator(&tc)
			mockDeletionService := tc.mockDeletionService(&tc)
			mockTaskService := tc.mockTaskService(&tc)

			handler := delete.New(
				ctx,
				logger,
				mockValidator,
				mockDeletionService,
				mockTaskService,
			)
			req, err := http.NewRequest(http.MethodPost, "/delete", bytes.NewReader([]byte(tc.input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			body := rr.Body.String()

			var resp delete.DeletionResponse

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			assert.Equal(t, tc.respError, resp.Error)
			assert.Equal(t, tc.statusCode, rr.Result().StatusCode)
			mockDeletionService.AssertExpectations(t)
			mockTaskService.AssertExpectations(t)
		})
	}
}
