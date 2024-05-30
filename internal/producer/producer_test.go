package producer_test

import (
	"testing"

	"metrics/internal/log"
	"metrics/internal/producer"
)

func prepare(t *testing.T) {
	t.Helper()

	log.Prepare()
}

func TestNewMetrics(t *testing.T) {
	prepare(t)

	t.Parallel()

	tests := []struct {
		name string
		want *producer.MetricsStore
	}{
		{
			name: "NewMetrics",
			want: new(producer.MetricsStore),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := producer.NewMetrics(); got == tt.want {
				t.Errorf("NewMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}
