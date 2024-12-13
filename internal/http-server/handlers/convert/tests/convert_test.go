package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chistyakoviv/converter/internal/http-server/handlers"
	"github.com/chistyakoviv/converter/internal/http-server/handlers/convert"
	handlersMocks "github.com/chistyakoviv/converter/internal/http-server/handlers/mocks"
	"github.com/chistyakoviv/converter/internal/logger/dummy"
	"github.com/chistyakoviv/converter/internal/service"
	serviceMocks "github.com/chistyakoviv/converter/internal/service/mocks"
)

func TestConvertHandler(t *testing.T) {
	var (
		ctx        = context.Background()
		logger     = dummy.NewDummyLogger()
		validation = validator.New()
	)

	type testcase struct {
		name                  string
		input                 string
		respError             string
		statusCode            int
		mockValidator         func(tc *testcase) handlers.Validator
		mockConversionService func(tc *testcase) service.ConversionQueueService
		mockTaskService       func(tc *testcase) service.TaskService
	}

	cases := []testcase{
		{
			name:       "Incorrect request: empty data",
			input:      "",
			respError:  "empty request",
			statusCode: http.StatusBadRequest,
			mockValidator: func(tc *testcase) handlers.Validator {
				mockValidator := handlersMocks.NewMockValidator(t)
				mockValidator.AssertNotCalled(t, "Struct")
				return mockValidator
			},
			mockConversionService: func(tc *testcase) service.ConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.AssertNotCalled(t, "Add")
				return mockConversionService
			},
			mockTaskService: func(tc *testcase) service.TaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				mockTaskService.AssertNotCalled(t, "TryQueueConversion")
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
			mockConversionService: func(tc *testcase) service.ConversionQueueService {
				mockConversionService := serviceMocks.NewMockConversionQueueService(t)
				mockConversionService.AssertNotCalled(t, "Add")
				return mockConversionService
			},
			mockTaskService: func(tc *testcase) service.TaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				mockTaskService.AssertNotCalled(t, "TryQueueConversion")
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
		})
	}
}
