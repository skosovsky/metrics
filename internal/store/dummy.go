package store

import (
	"metrics/internal/model"
)

type DummyStore struct{}

func NewDummyStore() *DummyStore {
	return &DummyStore{}
}

func (m *DummyStore) AddGauge(_ model.Gauge) {
}

func (m *DummyStore) AddCounter(_ model.Counter) {
}

func (m *DummyStore) GetGauge(_ string) (model.Gauge, error) {
	return model.Gauge{}, nil //nolint:exhaustruct // empty
}

func (m *DummyStore) GetAllGauges() []model.Gauge {
	return []model.Gauge{}
}

func (m *DummyStore) GetCounter(_ string) (model.Counter, error) {
	return model.Counter{}, nil //nolint:exhaustruct // empty
}

func (m *DummyStore) GetAllCounters() []model.Counter {
	return []model.Counter{}
}
