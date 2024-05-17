package service

import (
	"errors"

	"metrics/config"
	"metrics/internal/model"
	log "metrics/pkg/logger"
)

var (
	ErrGaugeNotAdded   = errors.New("gauge not added")
	ErrGaugeNotFound   = errors.New("gauge not found")
	ErrCounterNotAdded = errors.New("counter not added")
	ErrCounterNotFound = errors.New("counter not found")
)

type Store interface {
	AddGauge(model.Gauge)
	AddCounter(model.Counter)
	GetGauge(string) (model.Gauge, error)
	GetAllGauges() []model.Gauge
	GetCounter(string) (model.Counter, error)
	GetAllCounters() []model.Counter
}

type Receiver struct {
	store  Store
	config config.ReceiverConfig
}

func NewReceiverService(store Store, config config.ReceiverConfig) Receiver {
	return Receiver{
		store:  store,
		config: config,
	}
}

func (r Receiver) AddGauge(gaugeName string, gaugeValue float64) model.Gauge {
	gauge := model.Gauge{
		Name:  gaugeName,
		Value: gaugeValue,
	}

	r.store.AddGauge(gauge)

	log.Debug("gauge added", log.AnyAttr("gauge", gauge))

	return gauge
}

func (r Receiver) AddCounter(counterName string, counterValue int64) model.Counter {
	counter := model.Counter{
		Name:  counterName,
		Value: counterValue,
	}

	r.store.AddCounter(counter)

	log.Debug("counter added", log.AnyAttr("counter", counter))

	return counter
}

func (r Receiver) GetGauge(gaugeName string) (model.Gauge, error) {
	gauge, err := r.store.GetGauge(gaugeName)
	if err != nil {
		return model.Gauge{}, ErrGaugeNotFound
	}

	log.Debug("gauge returned", log.AnyAttr("gauge", gauge))

	return gauge, nil
}

func (r Receiver) GetAllGauges() []model.Gauge {
	gauges := r.store.GetAllGauges()

	log.Debug("all gauges returned", log.AnyAttr("gauge", gauges))

	return gauges
}

func (r Receiver) GetCounter(counterName string) (model.Counter, error) {
	counters, err := r.store.GetCounter(counterName)
	if err != nil {
		return model.Counter{}, ErrCounterNotFound
	}

	log.Debug("counters returned", log.AnyAttr("counters", counters))

	return counters, nil
}

func (r Receiver) GetAllCounters() []model.Counter {
	counters := r.store.GetAllCounters()

	log.Debug("all counters returned", log.AnyAttr("counters", counters))

	return counters
}
