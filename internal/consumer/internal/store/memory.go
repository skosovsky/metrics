package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"metrics/config"
	"metrics/internal/consumer/internal/service"
	"metrics/internal/log"
)

type (
	MemoryStore struct {
		memory map[string]service.Metric
		mu     sync.Mutex
	}
)

func NewMemoryStore(cfg config.Store) (*MemoryStore, error) {
	file, err := os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("open File %s error: %w", cfg.FileStoragePath, err)
	}

	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Error("close file error", cfg.FileStoragePath,
				log.ErrAttr(err))
		}
	}(file)

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("get File %s stat error: %w", cfg.FileStoragePath, err)
	}

	memoryStore := MemoryStore{
		memory: map[string]service.Metric{},
		mu:     sync.Mutex{},
	}

	if !cfg.ShouldRestore {
		if err = clearFile(file); err != nil {
			return nil, fmt.Errorf("clear File %s error: %w", cfg.FileStoragePath, err)
		}
	}

	if fileInfo.Size() != 0 {
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			data := scanner.Bytes()

			var metric service.Metric
			if err = json.Unmarshal(data, &metric); err != nil {
				return nil, fmt.Errorf("unmarshal json error: %w", err)
			}

			switch metric.MetricType {
			case service.MetricGauge:
				_ = memoryStore.AddGauge(metric) // err nil
			case service.MetricCounter:
				_ = memoryStore.AddCounter(metric, false) // err nil
			default:
				return nil, fmt.Errorf("metric type: %s, %w", metric.MetricType, service.ErrUnknownMetricType)
			}
		}

		if err = scanner.Err(); err != nil {
			return nil, fmt.Errorf("scanner error: %w", err)
		}
	}

	return &memoryStore, nil
}

func (m *MemoryStore) AddGauge(gauge service.Metric) error {
	m.mu.Lock()

	m.memory[gauge.ID] = gauge

	m.mu.Unlock()

	return nil
}

func (m *MemoryStore) AddCounter(counter service.Metric, increment bool) error {
	m.mu.Lock()

	current := m.memory[counter.ID]

	if current.Delta != nil && increment {
		*counter.Delta += *current.Delta
	}

	m.memory[counter.ID] = counter

	m.mu.Unlock()

	return nil
}

func (m *MemoryStore) GetMetric(id string) (service.Metric, error) {
	m.mu.Lock()

	metric, ok := m.memory[id]

	m.mu.Unlock()

	if !ok {
		return service.Metric{}, service.ErrMetricNotFound
	}

	return metric, nil
}

func (m *MemoryStore) GetAllMetrics() []service.Metric {
	m.mu.Lock()

	metrics := make([]service.Metric, 0, len(m.memory))

	for _, metric := range m.memory {
		metrics = append(metrics, metric)
	}

	m.mu.Unlock()

	return metrics
}

func (*MemoryStore) Close() {}

func clearFile(file *os.File) error {
	err := file.Truncate(0)
	if err != nil {
		return fmt.Errorf("truncate File %s error: %w", file.Name(), err)
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("seek File %s error: %w", file.Name(), err)
	}

	return nil
}
