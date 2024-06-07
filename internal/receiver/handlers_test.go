package receiver_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"metrics/config"
	"metrics/internal/log"
	"metrics/internal/receiver"
	"metrics/internal/receiver/internal/service"
	"metrics/internal/receiver/internal/store"
)

func prepare(t *testing.T) {
	t.Helper()

	log.Prepare()
}

func TestPostAddMetric(t *testing.T) {
	prepare(t)

	t.Parallel()

	type want struct {
		code        int
		response    string
		contentType string
	}

	testCases := []struct {
		name      string
		pathValue map[string]string
		want      want
	}{
		{
			name:      "Add empty metric, empty path value",
			pathValue: nil,
			want: want{
				code:        404,
				response:    "Not Found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:      "Add not valid kind metric",
			pathValue: map[string]string{"kind": "test", "name": "Test", "value": "0"},
			want: want{
				code:        400,
				response:    "Bad Request\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:      "Add not valid value metric",
			pathValue: map[string]string{"kind": "gauge", "name": "Test", "value": "wrong"},
			want: want{
				code:        400,
				response:    "Bad Request\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:      "Add not full metric",
			pathValue: map[string]string{"kind": "gauge"},
			want: want{
				code:        404,
				response:    "Not Found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:      "Add not full metric 2",
			pathValue: map[string]string{"kind": "gauge", "name": "Test"},
			want: want{
				code:        404,
				response:    "Not Found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:      "Add valid metric",
			pathValue: map[string]string{"kind": "gauge", "name": "Test", "value": "0"},
			want: want{
				code:        200,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	var cfg config.ReceiverConfig
	db := store.NewDummyStore()
	receiverService := service.NewReceiverService(db, cfg)
	handler := receiver.NewHandler(receiverService)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
			for name, value := range tt.pathValue {
				request.SetPathValue(name, value)
			}

			responseRecorder := httptest.NewRecorder()

			handler.AddMetric(responseRecorder, request)

			response := responseRecorder.Result()

			assert.Equal(t, tt.want.code, response.StatusCode)
			responseBody, err := io.ReadAll(response.Body)
			require.NoError(t, err)

			err = response.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.response, string(responseBody))
			assert.Equal(t, tt.want.contentType, response.Header.Get("Content-Type"))
		})
	}
}
