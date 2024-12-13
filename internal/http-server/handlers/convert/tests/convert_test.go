package tests

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
		mockTaskService := serviceMocks.NewMockTaskService(t)

		handler := convert.New(
			context.Background(),
			dummy.NewDummyLogger(),
			mockValidator,
			mockConversionService,
			mockTaskService,
		)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/convert", bytes.NewBuffer(nil))
		handler(w, req)

		require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})
}
