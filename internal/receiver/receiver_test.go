package receiver_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"metrics/config"
	"metrics/internal/receiver"
	"metrics/internal/service"
	"metrics/internal/store"
)

func TestRouting(t *testing.T) { // TODO: понять как же мы тут тестируем роутинг
	t.Parallel()

	type want struct {
		path   string
		status int
		body   string
		// answer string // TODO: add later
	}

	testCases := []struct {
		name string
		want want
	}{
		{
			name: "with correct data",
			want: want{
				path:   "/update/gauge/NumGC/5.11",
				status: 200,
				body:   "",
				// answer: "https://ya.ru", // TODO: add later
			},
		},
	}

	db, _ := store.NewDummyStore()
	cfg, _ := config.NewReceiverConfig()
	metricsGetter := service.NewMetricsGetterService(db, cfg)
	ctx := context.WithValue(context.Background(), receiver.KeyServiceCtx{}, metricsGetter)

	server := httptest.NewUnstartedServer(receiver.Handler())
	server.Config.ConnContext = func(_ context.Context, _ net.Conn) context.Context { return ctx }
	server.Start()

	t.Cleanup(server.Close)

	for _, tt := range testCases {
		tt := tt //nolint:copyloopvar // it's for stupid Yandex Practicum static test
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			url := server.URL + tt.want.path
			request, err := http.NewRequest(http.MethodPost, url, http.NoBody)
			require.NoError(t, err)

			response, err := http.DefaultClient.Do(request)
			require.NoError(t, err)

			assert.Equal(t, tt.want.status, response.StatusCode)

			responseBody, err := io.ReadAll(response.Body)
			require.NoError(t, err)

			err = response.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.body, string(responseBody))
		})
	}
}
