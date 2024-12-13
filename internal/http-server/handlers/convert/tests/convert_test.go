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

	"github.com/chistyakoviv/converter/internal/http-server/handlers/convert"
	handlersMocks "github.com/chistyakoviv/converter/internal/http-server/handlers/mocks"
	"github.com/chistyakoviv/converter/internal/logger/dummy"
	serviceMocks "github.com/chistyakoviv/converter/internal/service/mocks"
)

func TestConvertHandler(t *testing.T) {
	t.Run("Incorrect request: empty data", func(t *testing.T) {
		t.Parallel()

		mockValidator := handlersMocks.NewMockValidator(t)

		mockConversionService := serviceMocks.NewMockConversionQueueService(t)
		mockConversionService.AssertNotCalled(t, "Add")
		mockTaskService := serviceMocks.NewMockTaskService(t)
		mockTaskService.AssertNotCalled(t, "TryQueueConversion")

		handler := convert.New(
			context.Background(),
			dummy.NewDummyLogger(),
			mockValidator,
			mockConversionService,
			mockTaskService,
		)
		input := ""
		req, err := http.NewRequest(http.MethodPost, "/convert", bytes.NewReader([]byte(input)))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		body := rr.Body.String()

		var resp convert.ConversionResponse

		require.NoError(t, json.Unmarshal([]byte(body), &resp))

		assert.Equal(t, "empty request", resp.Error)
		assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
	})

	t.Run("Incorrect request: invalid data", func(t *testing.T) {
		t.Parallel()

		mockConversionService := serviceMocks.NewMockConversionQueueService(t)
		mockConversionService.AssertNotCalled(t, "Add")
		mockTaskService := serviceMocks.NewMockTaskService(t)
		mockTaskService.AssertNotCalled(t, "TryQueueConversion")

		handler := convert.New(
			context.Background(),
			dummy.NewDummyLogger(),
			validator.New(),
			mockConversionService,
			mockTaskService,
		)

		input := `{"fullpath": "/path/to/file.ext"}`
		req, err := http.NewRequest(http.MethodPost, "/convert", bytes.NewReader([]byte(input)))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		body := rr.Body.String()

		var resp convert.ConversionResponse

		require.NoError(t, json.Unmarshal([]byte(body), &resp))

		assert.Equal(t, "field Path is a required field", resp.Error)
		assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
	})
}
