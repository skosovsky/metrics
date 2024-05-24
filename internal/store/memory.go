package store

import (
	"sync"

	"metrics/internal/model"
)

type MemoryStore struct {
	memoryGauge    map[string]model.Gauge
	muGauge        sync.RWMutex
	memoryCounters map[string][]model.Counter
	muCounters     sync.RWMutex
}

func NewMemoryStore() (*MemoryStore, error) {
	return &MemoryStore{
		memoryGauge:    map[string]model.Gauge{},
		muGauge:        sync.RWMutex{},
		memoryCounters: map[string][]model.Counter{},
		muCounters:     sync.RWMutex{},
	}, nil
}

func (m *MemoryStore) AddGauge(gauge model.Gauge) bool {
	m.muGauge.Lock()
	m.memoryGauge[gauge.Name] = gauge
	m.muGauge.Unlock()

	return true
}

func (m *MemoryStore) AddCounter(counter model.Counter) bool {
	m.muCounters.Lock()
	m.memoryCounters[counter.Name] = append(m.memoryCounters[counter.Name], counter)
	m.muCounters.Unlock()

	return true
}

func (m *MemoryStore) GetGauge(name string) (model.Gauge, bool) {
	m.muGauge.RLock()
	gauge, ok := m.memoryGauge[name]
	m.muGauge.RUnlock()

	if !ok {
		return model.Gauge{}, false //nolint:exhaustruct // empty
	}

	return gauge, true
}

func (m *MemoryStore) GetCounters(name string) ([]model.Counter, bool) {
	m.muCounters.RLock()
	counters, ok := m.memoryCounters[name]
	m.muCounters.RUnlock()

	if !ok {
		return nil, false
	}

	return counters, true
}
