package service

import (
	"errors"
	"fmt"

	"metrics/config"
	log "metrics/internal/logger"
	"metrics/internal/model"
)

var (
	ErrGaugeNotFound   = errors.New("gauge not found")
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

	log.Debug("gauge added",
		log.StringAttr("name", gauge.Name),
		log.Float64Attr("gauge", gauge.Value))

	return gauge
}

func (r Receiver) AddCounter(counterName string, counterValue int64) model.Counter {
	counter := model.Counter{
		Name:  counterName,
		Value: counterValue,
	}

	r.store.AddCounter(counter)

	log.Debug("counter added",
		log.StringAttr("name", counter.Name),
		log.Int64Attr("counter", counter.Value))

	return counter
}

func (r Receiver) GetGauge(gaugeName string) (model.Gauge, error) {
	gauge, err := r.store.GetGauge(gaugeName)
	if err != nil {
		return model.Gauge{}, ErrGaugeNotFound
	}

	log.Debug("gauge returned",
		log.StringAttr("name", gauge.Name),
		log.Float64Attr("gauge", gauge.Value))

	return gauge, nil
}

func (r Receiver) GetAllGauges() []model.Gauge {
	gauges := r.store.GetAllGauges()

	log.Debug("all gauges returned",
		log.StringAttr("gauges", fmt.Sprintf("%v", gauges)))

	return gauges
}

func (r Receiver) GetCounter(counterName string) (model.Counter, error) {
	counter, err := r.store.GetCounter(counterName)
	if err != nil {
		return model.Counter{}, ErrCounterNotFound
	}

	log.Debug("counter returned",
		log.StringAttr("name", counter.Name),
		log.Int64Attr("counter", counter.Value))

	return counter, nil
}

func (r Receiver) GetAllCounters() []model.Counter {
	counters := r.store.GetAllCounters()

	log.Debug("all counters returned",
		log.StringAttr("counters", fmt.Sprintf("%v", counters)))

	return counters
}
