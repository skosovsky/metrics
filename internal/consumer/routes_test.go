package consumer_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"metrics/config"
	"metrics/internal/consumer"
	"metrics/internal/consumer/internal/service"
	"metrics/internal/consumer/internal/store"
)

func TestRouting(t *testing.T) {
	prepare(t)

	t.Parallel()

	type want struct {
		status int
	}

	testCases := []struct {
		name        string
		method      string
		request     string
		requestBody string
		want        want
	}{
		{
			name:        "Post update with requestBody, empty body",
			method:      http.MethodPost,
			request:     "/update",
			requestBody: "",
			want: want{
				status: 404,
			},
		},
		{
			name:        "Post update with requestBody, full path empty body",
			method:      http.MethodPost,
			request:     "/update/",
			requestBody: "",
			want: want{
				status: 400,
			},
		},
		{
			name:        "Post update with requestBody, full path, json",
			method:      http.MethodPost,
			request:     "/update/",
			requestBody: `{"id": "Test","type": "gauge","value": 0}`,
			want: want{
				status: 200,
			},
		},
		{
			name:        "Post value with requestBody, empty body",
			method:      http.MethodPost,
			request:     "/value",
			requestBody: "",
			want: want{
				status: 404,
			},
		},
		{
			name:        "Post update with requestBody, full path empty body",
			method:      http.MethodPost,
			request:     "/value/",
			requestBody: "",
			want: want{
				status: 400,
			},
		},
		{
			name:        "Get value with empty body",
			method:      http.MethodGet,
			request:     "/value/",
			requestBody: "",
			want: want{
				status: 404,
			},
		},
		{
			name:        "Put value with empty body",
			method:      http.MethodPut,
			request:     "/value/",
			requestBody: "",
			want: want{
				status: 405,
			},
		},
		{
			name:        "Delete value with empty body",
			method:      http.MethodDelete,
			request:     "/value/",
			requestBody: "",
			want: want{
				status: 405,
			},
		},
	}

	var cfg config.ConsumerConfig
	db := store.NewDummyStore()
	consumerService := service.NewConsumerService(db, cfg)
	handler := consumer.NewHandler(consumerService)

	server := httptest.NewServer(handler.InitRoutes())

	t.Cleanup(server.Close)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			url := server.URL + tt.request
			request, err := http.NewRequest(tt.method, url, strings.NewReader(tt.requestBody))
			require.NoError(t, err)

			response, err := http.DefaultClient.Do(request)
			require.NoError(t, err)

			assert.Equal(t, tt.want.status, response.StatusCode)

			err = response.Body.Close()
			require.NoError(t, err)
		})
	}
}
