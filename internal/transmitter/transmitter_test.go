package transmitter_test

import (
	"reflect"
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
			want: &transmitter.Metrics{}, //nolint:exhaustruct // empty
		},
	}
	for _, tt := range tests {
		tt := tt //nolint:copyloopvar // it's for stupid Yandex Practicum static test
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := transmitter.NewMetrics(); reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}
