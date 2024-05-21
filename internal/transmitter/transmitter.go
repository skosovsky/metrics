package transmitter

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"time"

	"metrics/config"
	"metrics/internal/log"
	"metrics/internal/model"
)

const (
	clientTimeout = 1 * time.Second
	baseProtocol  = "http://"
)

type Metrics struct {
	Gauges    map[string]model.Gauge
	PollCount model.Counter
}

func NewMetrics() *Metrics {
	return &Metrics{
		Gauges: map[string]model.Gauge{},
		PollCount: model.Counter{
			Name:  "",
			Value: 0,
		},
	}
}

func (m *Metrics) Update() {
	var memMetrics runtime.MemStats
	runtime.ReadMemStats(&memMetrics)

	m.PollCount.Value++

	m.Gauges["Alloc"] = model.Gauge{Name: "Alloc", Value: float64(memMetrics.Alloc)}
	m.Gauges["BuckHashSys"] = model.Gauge{Name: "BuckHashSys", Value: float64(memMetrics.BuckHashSys)}
	m.Gauges["Frees"] = model.Gauge{Name: "Frees", Value: float64(memMetrics.Frees)}
	m.Gauges["GCCPUFraction"] = model.Gauge{Name: "GCCPUFraction", Value: memMetrics.GCCPUFraction}
	m.Gauges["GCSys"] = model.Gauge{Name: "GCSys", Value: float64(memMetrics.GCSys)}
	m.Gauges["HeapAlloc"] = model.Gauge{Name: "HeapAlloc", Value: float64(memMetrics.HeapAlloc)}
	m.Gauges["HeapIdle"] = model.Gauge{Name: "HeapIdle", Value: float64(memMetrics.HeapIdle)}
	m.Gauges["HeapInuse"] = model.Gauge{Name: "HeapInuse", Value: float64(memMetrics.HeapInuse)}
	m.Gauges["LastGC"] = model.Gauge{Name: "LastGC", Value: float64(memMetrics.LastGC)}
	m.Gauges["Lookups"] = model.Gauge{Name: "Lookups", Value: float64(memMetrics.Lookups)}
	m.Gauges["MCacheInuse"] = model.Gauge{Name: "MCacheInuse", Value: float64(memMetrics.MCacheInuse)}
	m.Gauges["MCacheSys"] = model.Gauge{Name: "MCacheSys", Value: float64(memMetrics.MCacheSys)}
	m.Gauges["MSpanInuse"] = model.Gauge{Name: "MSpanInuse", Value: float64(memMetrics.MSpanInuse)}
	m.Gauges["MSpanSys"] = model.Gauge{Name: "MSpanSys", Value: float64(memMetrics.MSpanSys)}
	m.Gauges["Mallocs"] = model.Gauge{Name: "Mallocs", Value: float64(memMetrics.Mallocs)}
	m.Gauges["NextGC"] = model.Gauge{Name: "NextGC", Value: float64(memMetrics.NextGC)}
	m.Gauges["NumForcedGC"] = model.Gauge{Name: "NumForcedGC", Value: float64(memMetrics.NumForcedGC)}
	m.Gauges["NumGC"] = model.Gauge{Name: "NumGC", Value: float64(memMetrics.NumGC)}
	m.Gauges["OtherSys"] = model.Gauge{Name: "OtherSys", Value: float64(memMetrics.OtherSys)}
	m.Gauges["PauseTotalNs"] = model.Gauge{Name: "PauseTotalNs", Value: float64(memMetrics.PauseTotalNs)}
	m.Gauges["StackInuse"] = model.Gauge{Name: "StackInuse", Value: float64(memMetrics.StackInuse)}
	m.Gauges["StackSys"] = model.Gauge{Name: "StackSys", Value: float64(memMetrics.StackSys)}
	m.Gauges["Sys"] = model.Gauge{Name: "Sys", Value: float64(memMetrics.Sys)}
	m.Gauges["TotalAlloc"] = model.Gauge{Name: "TotalAlloc", Value: float64(memMetrics.TotalAlloc)}
	m.Gauges["RandomValue"] = model.Gauge{Name: "RandomValue", Value: float64(rand.Int())} //nolint:gosec // i know
}

func (m *Metrics) Report(cfg config.Transmitter) error {
	urls, err := m.prepareUrls(cfg)
	if err != nil {
		return fmt.Errorf("prepare urls: %w", err)
	}

	m.sendRequest(urls)
	m.PollCount.Value = 0

	return nil
}

func (m *Metrics) prepareUrls(cfg config.Transmitter) ([]string, error) {
	var err error
	urls := make([]string, 0, len(m.Gauges)+1)

	for _, gauge := range m.Gauges {
		urlGauge, errGauge := url.JoinPath(baseProtocol+cfg.Address.String(), "update", "gauge", gauge.Name, strconv.FormatFloat(gauge.Value, 'f', -1, 64))
		if errGauge != nil {
			err = errors.Join(err, errGauge)
		}

		urls = append(urls, urlGauge)
	}

	urlCounter, errCounter := url.JoinPath(baseProtocol+cfg.Address.String(), "update", "counter", m.PollCount.Name, strconv.FormatInt(m.PollCount.Value, 10))
	if errCounter != nil {
		err = errors.Join(err, errCounter)
	}

	urls = append(urls, urlCounter)

	if err != nil {
		return nil, err
	}

	return urls, nil
}

func (*Metrics) sendRequest(urls []string) {
	var contentType = "text/plain"
	var client = &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       clientTimeout,
	}

	for _, urlMetric := range urls {
		response, err := client.Post(urlMetric, contentType, http.NoBody) //nolint:noctx //TODO: добавить контекст, прокинуть от запуска
		if err != nil {
			log.Error("Failed to send request",
				log.ErrAttr(err),
				log.StringAttr("url", urlMetric),
			)

			return
		}

		err = response.Body.Close()
		if err != nil {
			log.Error("Failed to close response body",
				log.ErrAttr(err))
		}
	}
}
