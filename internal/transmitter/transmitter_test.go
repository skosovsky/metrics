package transmitter_test

import (
	"testing"

	"metrics/internal/transmitter"
)

func TestNewMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want *transmitter.Metrics
	}{
		{
			name: "NewMetrics",
			want: new(transmitter.Metrics),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := transmitter.NewMetrics(); got == tt.want {
				t.Errorf("NewMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}
