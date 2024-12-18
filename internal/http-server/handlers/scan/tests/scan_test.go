package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chistyakoviv/converter/internal/constants"
	"github.com/chistyakoviv/converter/internal/http-server/handlers/scan"
	"github.com/chistyakoviv/converter/internal/logger/dummy"
	"github.com/chistyakoviv/converter/internal/service/mocks"
	serviceMocks "github.com/chistyakoviv/converter/internal/service/mocks"
)

func TestScanHandler(t *testing.T) {
	var (
		ctx    = context.Background()
		logger = dummy.NewDummyLogger()
	)

	type testcase struct {
		name            string
		input           string
		respError       string
		statusCode      int
		mockTaskService func(tc *testcase) *mocks.MockTaskService
	}

	cases := []testcase{
		{
			name:       "Failed request: scan is already running",
			respError:  "scan is already running",
			statusCode: http.StatusConflict,
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				mockTaskService.On("IsScanning").Return(true).Once()
				return mockTaskService
			},
		},
		{
			name:       "Successful request",
			respError:  "",
			statusCode: http.StatusOK,
			mockTaskService: func(tc *testcase) *mocks.MockTaskService {
				mockTaskService := serviceMocks.NewMockTaskService(t)
				mockTaskService.On("IsScanning").Return(false).Once()
				mockTaskService.On("ProcessScanfs", ctx, constants.FilesRootDir).Return(nil).Maybe()
				mockTaskService.On("TryQueueConversion").Return(true).Maybe()
				return mockTaskService
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockTaskService := tc.mockTaskService(&tc)

			handler := scan.New(
				ctx,
				logger,
				mockTaskService,
			)
			req, err := http.NewRequest(http.MethodPost, "/scan", bytes.NewReader([]byte(tc.input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			body := rr.Body.String()

			var resp scan.ScanResponse

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			assert.Equal(t, tc.respError, resp.Error)
			assert.Equal(t, tc.statusCode, rr.Result().StatusCode)
			mockTaskService.AssertExpectations(t)
		})
	}
}
