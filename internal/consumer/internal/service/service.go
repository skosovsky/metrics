package service

import (
	"errors"
	"fmt"

	"metrics/config"
	"metrics/internal/log"
)

const (
	MetricCounter = "counter"
	MetricGauge   = "gauge"
)

type (
	Metric struct {
		ID         string   `json:"id"              validate:"required"`
		MetricType string   `json:"type"            validate:"required,oneof=gauge counter"`
		Delta      *int64   `json:"delta,omitempty"`
		Value      *float64 `json:"value,omitempty"`
	}
)

var (
	ErrMetricNotFound    = errors.New("metric not found")
	ErrUnknownMetricType = errors.New("unknown metric type")
	ErrUnknownDBType     = errors.New("unknown db type")
)

type Store interface {
	AddGauge(gauge Metric) error
	AddCounter(counter Metric, increment bool) error
	GetMetric(id string) (Metric, error)
	GetAllMetrics() []Metric
	Close()
}

type Consumer struct {
	store  Store
	config config.ConsumerConfig
}

func NewConsumerService(store Store, config config.ConsumerConfig) Consumer {
	return Consumer{
		store:  store,
		config: config,
	}
}

func (c Consumer) AddGauge(gaugeName string, gaugeValue float64) (Metric, error) {
	gauge := Metric{
		ID:         gaugeName,
		MetricType: MetricGauge,
		Value:      &gaugeValue,
		Delta:      nil,
	}

	if err := c.store.AddGauge(gauge); err != nil {
		return Metric{}, fmt.Errorf("failed to add gauge %s: %w", gaugeName, err)
	}

	log.Debug("gauge added",
		log.StringAttr("name", gauge.ID),
		log.Float64Attr("gauge", *gauge.Value))

	return gauge, nil
}

func (c Consumer) AddCounter(counterName string, counterValue int64) (Metric, error) {
	counter := Metric{
		ID:         counterName,
		MetricType: MetricCounter,
		Value:      nil,
		Delta:      &counterValue,
	}

	if err := c.store.AddCounter(counter, true); err != nil {
		return Metric{}, fmt.Errorf("failed to add gauge %s: %w", counterName, err)
	}

	log.Debug("counter added",
		log.StringAttr("name", counter.ID),
		log.Int64Attr("counter", *counter.Delta))

	return counter, nil
}

func (c Consumer) GetMetric(id string) (Metric, error) {
	metric, err := c.store.GetMetric(id)
	if err != nil {
		return Metric{}, ErrMetricNotFound
	}

	switch metric.MetricType {
	case MetricGauge:
		log.Debug("gauge returned",
			log.StringAttr("name", metric.ID),
			log.Float64Attr("gauge", *metric.Value))
	case MetricCounter:
		log.Debug("counter returned",
			log.StringAttr("name", metric.ID),
			log.Int64Attr("counter", *metric.Delta))
	default:
		return Metric{}, ErrUnknownMetricType
	}

	return metric, nil
}

func (c Consumer) GetAllMetrics() []Metric {
	metrics := c.store.GetAllMetrics()

	log.Debug("all metrics returned",
		log.StringAttr("metrics", fmt.Sprintf("%v", metrics)))

	return metrics
}
