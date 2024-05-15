package service

import (
	"errors"

	"metrics/config"
	"metrics/internal/model"
	"metrics/internal/store"
	log "metrics/pkg/logger"
)

var (
	ErrGaugeNotAdded    = errors.New("gauge not added")
	ErrGaugeNotFound    = errors.New("gauge not found")
	ErrCounterNotAdded  = errors.New("counter not added")
	ErrCountersNotFound = errors.New("counters not found")
)

type MetricsGetter struct {
	store  store.Store
	config config.ReceiverConfig
}

func NewMetricsGetterService(store store.Store, config config.ReceiverConfig) MetricsGetter {
	return MetricsGetter{
		store:  store,
		config: config,
	}
}

func (m MetricsGetter) AddGauge(gaugeName string, gaugeValue float64) (model.Gauge, error) {
	gauge := model.Gauge{
		Name:  gaugeName,
		Value: gaugeValue,
	}

	if ok := m.store.AddGauge(gauge); !ok {
		return model.Gauge{}, ErrGaugeNotAdded
	}

	log.Info("gauge added", log.AnyAttr("gauge", gauge))

	return gauge, nil
}

func (m MetricsGetter) AddCounter(counterName string, counterValue int64) (model.Counter, error) {
	counter := model.Counter{
		Name:  counterName,
		Value: counterValue,
	}

	if ok := m.store.AddCounter(counter); !ok {
		return model.Counter{}, ErrCounterNotAdded
	}

	log.Info("counter added", log.AnyAttr("counter", counter))

	return counter, nil
}

func (m MetricsGetter) GetGauge(gaugeName string) (model.Gauge, error) {
	gauge, ok := m.store.GetGauge(gaugeName)
	if !ok {
		return model.Gauge{}, ErrGaugeNotFound
	}

	log.Info("gauge returned", log.AnyAttr("gauge", gauge))

	return gauge, nil
}

func (m MetricsGetter) GetAllGauges() []model.Gauge {
	gauges := m.store.GetAllGauges()

	log.Info("all gauges returned", log.AnyAttr("gauge", gauges))

	return gauges
}

func (m MetricsGetter) GetCounters(counterName string) ([]model.Counter, error) {
	counters, ok := m.store.GetCounters(counterName)
	if !ok {
		return nil, ErrCountersNotFound
	}

	log.Info("counters returned", log.AnyAttr("counters", counters))

	return counters, nil
}

func (m MetricsGetter) GetAllCounters() [][]model.Counter {
	counters := m.store.GetAllCounters()

	log.Info("all counters returned", log.AnyAttr("counters", counters))

	return counters
}
