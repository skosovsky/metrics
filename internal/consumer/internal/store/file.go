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
	File    *os.File
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

	if !cfg.ShouldRestore {
		if err = ClearFile(file); err != nil {
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

	return &FileStore{
		File:        file,
		encoder:     json.NewEncoder(file),
		MemoryStore: memoryStore,
	}, nil
}

func (f *FileStore) AddGauge(gauge service.Metric) error {
	_ = f.MemoryStore.AddGauge(gauge) // err nil

	err := f.deleteMetric(gauge.ID)
	if err != nil {
		return fmt.Errorf("add gauge: %w", err)
	}

	err = f.encoder.Encode(gauge)
	if err != nil {
		return fmt.Errorf("encode File %s error: %w", f.File.Name(), err)
	}

	return nil
}

func (f *FileStore) AddCounter(counter service.Metric, increment bool) error {
	_ = f.MemoryStore.AddCounter(counter, increment) // err nil

	err := f.deleteMetric(counter.ID)
	if err != nil {
		return fmt.Errorf("add counter: %w", err)
	}

	err = f.encoder.Encode(counter)
	if err != nil {
		return fmt.Errorf("encode File %s error: %w", f.File.Name(), err)
	}

	return nil
}

func (f *FileStore) deleteMetric(id string) error {
	var err error

	_, err = f.File.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("seek File %s error: %w", f.File.Name(), err)
	}

	var lines [][]byte
	var found bool
	scanner := bufio.NewScanner(f.File)

	for scanner.Scan() {
		data := scanner.Bytes()

		var metric service.Metric
		if err = json.Unmarshal(data, &metric); err != nil {
			return fmt.Errorf("unmarshal json error: %w", err)
		}

		if metric.ID == id {
			found = true

			continue
		}

		lines = append(lines, data, []byte("\n"))
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	if !found {
		return nil
	}

	if err = ClearFile(f.File); err != nil {
		return fmt.Errorf("clear File error: %w", err)
	}

	writer := bufio.NewWriter(f.File)

	for _, line := range lines {
		_, err = writer.Write(line)
		if err != nil {
			return fmt.Errorf("write File %s error: %w", f.File.Name(), err)
		}
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("flush File %s error: %w", f.File.Name(), err)
	}

	return nil
}

func (f *FileStore) Close() {
	err := f.File.Close()
	if err != nil {
		log.Error("close File %s error", f.File.Name(),
			log.ErrAttr(err))
	}
}

func ClearFile(file *os.File) error {
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
