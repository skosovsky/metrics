package receiver_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"metrics/config"
	"metrics/internal/receiver"
	"metrics/internal/receiver/internal/service"
	"metrics/internal/receiver/internal/store"
)

func TestRouting(t *testing.T) {
	t.Parallel()

	type want struct {
		path   string
		status int
		body   string
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
			},
		},
		{
			name: "with incorrect data",
			want: want{
				path:   "/update/gauge/NumGC/ab",
				status: 400,
				body:   "Bad Request\n",
			},
		},
		{
			name: "with incorrect path",
			want: want{
				path:   "/delete/gauge/NumGC/ab",
				status: 404,
				body:   "404 page not found\n",
			},
		},
	}

	db := store.NewDummyStore()
	cfg, _ := config.NewReceiverConfig()
	receiverService := service.NewReceiverService(db, cfg)
	handler := receiver.NewHandler(receiverService)

	server := httptest.NewServer(handler.InitRoutes())

	t.Cleanup(server.Close)

	for _, tt := range testCases {
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
