package service

import (
	"errors"
	"fmt"

	"metrics/config"
	"metrics/internal/log"
)

type (
	Counter struct {
		Name  string
		Value int64
	}

	Gauge struct {
		Name  string
		Value float64
	}
)

var (
	ErrGaugeNotFound   = errors.New("gauge not found")
	ErrCounterNotFound = errors.New("counter not found")
)

type Store interface {
	AddGauge(Gauge)
	AddCounter(Counter)
	GetGauge(string) (Gauge, error)
	GetAllGauges() []Gauge
	GetCounter(string) (Counter, error)
	GetAllCounters() []Counter
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

func (r Receiver) AddGauge(gaugeName string, gaugeValue float64) Gauge {
	gauge := Gauge{
		Name:  gaugeName,
		Value: gaugeValue,
	}

	r.store.AddGauge(gauge)

	log.Debug("gauge added",
		log.StringAttr("name", gauge.Name),
		log.Float64Attr("gauge", gauge.Value))

	return gauge
}

func (r Receiver) AddCounter(counterName string, counterValue int64) Counter {
	counter := Counter{
		Name:  counterName,
		Value: counterValue,
	}

	r.store.AddCounter(counter)

	log.Debug("counter added",
		log.StringAttr("name", counter.Name),
		log.Int64Attr("counter", counter.Value))

	return counter
}

func (r Receiver) GetGauge(gaugeName string) (Gauge, error) {
	gauge, err := r.store.GetGauge(gaugeName)
	if err != nil {
		return Gauge{}, ErrGaugeNotFound
	}

	log.Debug("gauge returned",
		log.StringAttr("name", gauge.Name),
		log.Float64Attr("gauge", gauge.Value))

	return gauge, nil
}

func (r Receiver) GetAllGauges() []Gauge {
	gauges := r.store.GetAllGauges()

	log.Debug("all gauges returned",
		log.StringAttr("gauges", fmt.Sprintf("%v", gauges)))

	return gauges
}

func (r Receiver) GetCounter(counterName string) (Counter, error) {
	counter, err := r.store.GetCounter(counterName)
	if err != nil {
		return Counter{}, ErrCounterNotFound
	}

	log.Debug("counter returned",
		log.StringAttr("name", counter.Name),
		log.Int64Attr("counter", counter.Value))

	return counter, nil
}

func (r Receiver) GetAllCounters() []Counter {
	counters := r.store.GetAllCounters()

	log.Debug("all counters returned",
		log.StringAttr("counters", fmt.Sprintf("%v", counters)))

	return counters
}
