package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"metrics/config"
	"metrics/internal/consumer/internal/service"
	"metrics/internal/log"
)

type FileStore struct {
	file    *os.File
	encoder *json.Encoder
	*MemoryStore
}

func NewFileStore(cfg config.Store) (*FileStore, error) {
	file, err := os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("open File %s error: %w", cfg.FileStoragePath, err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("get File %s stat error: %w", cfg.FileStoragePath, err)
	}

	memoryStore, err := NewMemoryStore(cfg)
	if err != nil {
		return nil, fmt.Errorf("create memory store error: %w", err)
	}

	fileStore := &FileStore{
		file:        file,
		encoder:     json.NewEncoder(file),
		MemoryStore: memoryStore,
	}

	if !cfg.ShouldRestore {
		if err = fileStore.Clear(); err != nil {
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
				_ = fileStore.MemoryStore.AddGauge(metric) // err nil
			case service.MetricCounter:
				_ = fileStore.MemoryStore.AddCounter(metric, false) // err nil
			default:
				return nil, fmt.Errorf("metric type: %s, %w", metric.MetricType, service.ErrUnknownMetricType)
			}
		}

		if err = scanner.Err(); err != nil {
			return nil, fmt.Errorf("scanner error: %w", err)
		}
	}

	return fileStore, nil
}

func (f *FileStore) AddGauge(gauge service.Metric) error {
	_ = f.MemoryStore.AddGauge(gauge) // err nil

	if err := f.saveAllMetrics(); err != nil {
		return fmt.Errorf("save all metrics error: %w", err)
	}

	return nil
}

func (f *FileStore) AddCounter(counter service.Metric, increment bool) error {
	_ = f.MemoryStore.AddCounter(counter, increment) // err nil

	if err := f.saveAllMetrics(); err != nil {
		return fmt.Errorf("save all metrics error: %w", err)
	}

	return nil
}

func (f *FileStore) saveAllMetrics() error {
	for _, metric := range f.memory {
		if metric.MetricType != service.MetricCounter && metric.MetricType != service.MetricGauge {
			log.Error("unknown metric type",
				log.ErrAttr(service.ErrUnknownMetricType))

			continue
		}

		err := f.encoder.Encode(metric)
		if err != nil {
			return fmt.Errorf("encode File %s error: %w", f.file.Name(), err)
		}
	}

	return nil
}

func (f *FileStore) Close() {
	err := f.file.Close()
	if err != nil {
		log.Error("close File %s error", f.file.Name(),
			log.ErrAttr(err))
	}
}

func (f *FileStore) Clear() error {
	err := f.file.Truncate(0)
	if err != nil {
		return fmt.Errorf("truncate File %s error: %w", f.file.Name(), err)
	}

	_, err = f.file.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("seek File %s error: %w", f.file.Name(), err)
	}

	return nil
}
