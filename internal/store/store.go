package store

import "metrics/internal/model"

type Store interface {
	AddGauge(model.Gauge) bool
	AddCounter(model.Counter) bool
	GetGauge(string) (model.Gauge, bool)
	GetCounters(string) ([]model.Counter, bool)
}
