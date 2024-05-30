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

func (c Consumer) AddGauge(gaugeName string, gaugeValue float64) Gauge {
	gauge := Gauge{
		Name:  gaugeName,
		Value: gaugeValue,
	}

	c.store.AddGauge(gauge)

	log.Debug("gauge added",
		log.StringAttr("name", gauge.Name),
		log.Float64Attr("gauge", gauge.Value))

	return gauge
}

func (c Consumer) AddCounter(counterName string, counterValue int64) Counter {
	counter := Counter{
		Name:  counterName,
		Value: counterValue,
	}

	c.store.AddCounter(counter)

	log.Debug("counter added",
		log.StringAttr("name", counter.Name),
		log.Int64Attr("counter", counter.Value))

	return counter
}

func (c Consumer) GetGauge(gaugeName string) (Gauge, error) {
	gauge, err := c.store.GetGauge(gaugeName)
	if err != nil {
		return Gauge{}, ErrGaugeNotFound
	}

	log.Debug("gauge returned",
		log.StringAttr("name", gauge.Name),
		log.Float64Attr("gauge", gauge.Value))

	return gauge, nil
}

func (c Consumer) GetAllGauges() []Gauge {
	gauges := c.store.GetAllGauges()

	log.Debug("all gauges returned",
		log.StringAttr("gauges", fmt.Sprintf("%v", gauges)))

	return gauges
}

func (c Consumer) GetCounter(counterName string) (Counter, error) {
	counter, err := c.store.GetCounter(counterName)
	if err != nil {
		return Counter{}, ErrCounterNotFound
	}

	log.Debug("counter returned",
		log.StringAttr("name", counter.Name),
		log.Int64Attr("counter", counter.Value))

	return counter, nil
}

func (c Consumer) GetAllCounters() []Counter {
	counters := c.store.GetAllCounters()

	log.Debug("all counters returned",
		log.StringAttr("counters", fmt.Sprintf("%v", counters)))

	return counters
}
