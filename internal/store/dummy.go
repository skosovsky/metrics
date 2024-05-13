package store

import (
	"metrics/internal/model"
)

type DummyStore struct{}

func NewDummyStore() (*DummyStore, error) {
	return &DummyStore{}, nil
}

func (m *DummyStore) AddGauge(_ model.Gauge) bool {
	return true
}

func (m *DummyStore) AddCounter(_ model.Counter) bool {
	return true
}

func (m *DummyStore) GetGauge(_ string) (model.Gauge, bool) {
	return model.Gauge{}, true //nolint:exhaustruct // empty
}

func (m *DummyStore) GetAllGauges() []model.Gauge {
	return []model.Gauge{}
}

func (m *DummyStore) GetCounters(_ string) ([]model.Counter, bool) {
	return nil, true
}

func (m *DummyStore) GetAllCounters() [][]model.Counter {
	return [][]model.Counter{}
}
