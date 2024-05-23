package store

import (
	"errors"
	"sync"

	"metrics/internal/receiver/internal/service"
)

var (
	ErrGaugeNotFound   = errors.New("gauge not found")
	ErrCounterNotFound = errors.New("counter not found")
)

type (
	gaugeStore struct {
		memory map[string]service.Gauge
		mu     sync.Mutex
	}

	counterStore struct {
		memory map[string]service.Counter
		mu     sync.Mutex
	}

	MemoryStore struct {
		gauge   gaugeStore
		counter counterStore
	}
)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		gauge: gaugeStore{
			memory: map[string]service.Gauge{},
			mu:     sync.Mutex{},
		},
		counter: counterStore{
			memory: map[string]service.Counter{},
			mu:     sync.Mutex{},
		},
	}
}

func (m *MemoryStore) AddGauge(gauge service.Gauge) {
	m.gauge.mu.Lock()
	m.gauge.memory[gauge.Name] = gauge
	m.gauge.mu.Unlock()
}

func (m *MemoryStore) AddCounter(counter service.Counter) {
	m.counter.mu.Lock()
	current := m.counter.memory[counter.Name]
	m.counter.mu.Unlock()

	counter.Value += current.Value

	m.counter.mu.Lock()
	m.counter.memory[counter.Name] = counter
	m.counter.mu.Unlock()
}

func (m *MemoryStore) GetGauge(name string) (service.Gauge, error) {
	m.gauge.mu.Lock()
	gauge, ok := m.gauge.memory[name]
	m.gauge.mu.Unlock()

	if !ok {
		return service.Gauge{}, ErrGaugeNotFound
	}

	return gauge, nil
}

func (m *MemoryStore) GetAllGauges() []service.Gauge {
	gauges := make([]service.Gauge, 0, len(m.gauge.memory))

	m.gauge.mu.Lock()

	for _, gauge := range m.gauge.memory {
		gauges = append(gauges, gauge)
	}

	m.gauge.mu.Unlock()

	return gauges
}

func (m *MemoryStore) GetCounter(name string) (service.Counter, error) {
	m.counter.mu.Lock()
	counters, ok := m.counter.memory[name]
	m.counter.mu.Unlock()

	if !ok {
		return service.Counter{}, ErrCounterNotFound
	}

	return counters, nil
}

func (m *MemoryStore) GetAllCounters() []service.Counter {
	counters := make([]service.Counter, 0, len(m.counter.memory))

	m.counter.mu.Lock()

	for _, counter := range m.counter.memory {
		counters = append(counters, counter)
	}

	m.counter.mu.Unlock()

	return counters
}
