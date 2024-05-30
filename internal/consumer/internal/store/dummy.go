package store

import "metrics/internal/consumer/internal/service"

type DummyStore struct{}

func NewDummyStore() *DummyStore {
	return &DummyStore{}
}

func (m *DummyStore) AddGauge(_ service.Gauge) {
}

func (m *DummyStore) AddCounter(_ service.Counter) {
}

func (m *DummyStore) GetGauge(_ string) (service.Gauge, error) {
	return service.Gauge{}, nil //nolint:exhaustruct // empty
}

func (m *DummyStore) GetAllGauges() []service.Gauge {
	return []service.Gauge{}
}

func (m *DummyStore) GetCounter(_ string) (service.Counter, error) {
	return service.Counter{}, nil //nolint:exhaustruct // empty
}

func (m *DummyStore) GetAllCounters() []service.Counter {
	return []service.Counter{}
}
