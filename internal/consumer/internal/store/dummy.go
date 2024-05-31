package store

import "metrics/internal/consumer/internal/service"

type DummyStore struct{}

func NewDummyStore() *DummyStore {
	return &DummyStore{}
}

func (*DummyStore) AddGauge(_ service.Metric) error {
	return nil
}

func (*DummyStore) AddCounter(_ service.Metric, _ bool) error {
	return nil
}

func (*DummyStore) GetMetric(_ string) (service.Metric, error) {
	return service.Metric{}, nil //nolint:exhaustruct // empty
}

func (*DummyStore) GetAllMetrics() []service.Metric {
	return []service.Metric{}
}

func (*DummyStore) Close() {}
