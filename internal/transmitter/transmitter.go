package transmitter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"time"

	"metrics/config"
	"metrics/internal/log"
)

const (
	clientTimeout = 10 * time.Second
	baseProtocol  = "http://"
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

	MetricsStore struct {
		Gauges    map[string]Gauge
		PollCount Counter
	}
)

func NewMetrics() *MetricsStore {
	return &MetricsStore{
		Gauges: map[string]Gauge{},
		PollCount: Counter{
			Name:  "",
			Value: 0,
		},
	}
}

func (m *MetricsStore) Update() {
	var memMetrics runtime.MemStats
	runtime.ReadMemStats(&memMetrics)

	m.PollCount.Name = "PollCount"
	m.PollCount.Value++

	m.Gauges["Alloc"] = Gauge{Name: "Alloc", Value: float64(memMetrics.Alloc)}
	m.Gauges["BuckHashSys"] = Gauge{Name: "BuckHashSys", Value: float64(memMetrics.BuckHashSys)}
	m.Gauges["Frees"] = Gauge{Name: "Frees", Value: float64(memMetrics.Frees)}
	m.Gauges["GCCPUFraction"] = Gauge{Name: "GCCPUFraction", Value: memMetrics.GCCPUFraction}
	m.Gauges["GCSys"] = Gauge{Name: "GCSys", Value: float64(memMetrics.GCSys)}
	m.Gauges["HeapAlloc"] = Gauge{Name: "HeapAlloc", Value: float64(memMetrics.HeapAlloc)}
	m.Gauges["HeapIdle"] = Gauge{Name: "HeapIdle", Value: float64(memMetrics.HeapIdle)}
	m.Gauges["HeapInuse"] = Gauge{Name: "HeapInuse", Value: float64(memMetrics.HeapInuse)}
	m.Gauges["HeapObjects"] = Gauge{Name: "HeapObjects", Value: float64(memMetrics.HeapObjects)}
	m.Gauges["HeapReleased"] = Gauge{Name: "HeapReleased", Value: float64(memMetrics.HeapReleased)}
	m.Gauges["HeapSys"] = Gauge{Name: "HeapSys", Value: float64(memMetrics.HeapSys)}
	m.Gauges["LastGC"] = Gauge{Name: "LastGC", Value: float64(memMetrics.LastGC)}
	m.Gauges["Lookups"] = Gauge{Name: "Lookups", Value: float64(memMetrics.Lookups)}
	m.Gauges["MCacheInuse"] = Gauge{Name: "MCacheInuse", Value: float64(memMetrics.MCacheInuse)}
	m.Gauges["MCacheSys"] = Gauge{Name: "MCacheSys", Value: float64(memMetrics.MCacheSys)}
	m.Gauges["MSpanInuse"] = Gauge{Name: "MSpanInuse", Value: float64(memMetrics.MSpanInuse)}
	m.Gauges["MSpanSys"] = Gauge{Name: "MSpanSys", Value: float64(memMetrics.MSpanSys)}
	m.Gauges["Mallocs"] = Gauge{Name: "Mallocs", Value: float64(memMetrics.Mallocs)}
	m.Gauges["NextGC"] = Gauge{Name: "NextGC", Value: float64(memMetrics.NextGC)}
	m.Gauges["NumForcedGC"] = Gauge{Name: "NumForcedGC", Value: float64(memMetrics.NumForcedGC)}
	m.Gauges["NumGC"] = Gauge{Name: "NumGC", Value: float64(memMetrics.NumGC)}
	m.Gauges["OtherSys"] = Gauge{Name: "OtherSys", Value: float64(memMetrics.OtherSys)}
	m.Gauges["PauseTotalNs"] = Gauge{Name: "PauseTotalNs", Value: float64(memMetrics.PauseTotalNs)}
	m.Gauges["StackInuse"] = Gauge{Name: "StackInuse", Value: float64(memMetrics.StackInuse)}
	m.Gauges["StackSys"] = Gauge{Name: "StackSys", Value: float64(memMetrics.StackSys)}
	m.Gauges["Sys"] = Gauge{Name: "Sys", Value: float64(memMetrics.Sys)}
	m.Gauges["TotalAlloc"] = Gauge{Name: "TotalAlloc", Value: float64(memMetrics.TotalAlloc)}
	m.Gauges["RandomValue"] = Gauge{Name: "RandomValue", Value: float64(rand.Int())} //nolint:gosec // i know
}

func (m *MetricsStore) Report(cfg config.Transmitter) error {
	err := m.reportURL(cfg)
	if err != nil {
		return fmt.Errorf("reporting url metrics: %w", err)
	}

	err = m.reportJSON(cfg)
	if err != nil {
		return fmt.Errorf("reporting json metrics: %w", err)
	}

	m.PollCount.Value = 0

	return nil
}

func (m *MetricsStore) reportURL(cfg config.Transmitter) error {
	urls, err := m.prepareURLs(cfg)
	if err != nil {
		return fmt.Errorf("prepare urls: %w", err)
	}

	m.sendRequest(urls)

	return nil
}

func (m *MetricsStore) reportJSON(cfg config.Transmitter) error {
	jsons, err := m.prepareJSONs()
	if err != nil {
		return fmt.Errorf("prepare urls: %w", err)
	}

	m.sendRequestJSON(cfg, jsons)

	return nil
}

func (m *MetricsStore) prepareURLs(cfg config.Transmitter) ([]string, error) {
	var err error
	urls := make([]string, 0, len(m.Gauges)+1)

	for _, metric := range m.Gauges {
		urlGauge, errGauge := url.JoinPath(baseProtocol+cfg.Address.String(), "update", "gauge", metric.Name, strconv.FormatFloat(metric.Value, 'f', -1, 64))
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

func (m *MetricsStore) prepareJSONs() ([][]byte, error) {
	type Metric struct {
		ID    string   `json:"id"`
		MType string   `json:"type"`
		Delta *int64   `json:"delta,omitempty"`
		Value *float64 `json:"value,omitempty"`
	}

	var err error
	jsons := make([][]byte, 0, len(m.Gauges)+1)

	for _, metric := range m.Gauges {
		gauge := Metric{
			ID:    metric.Name,
			MType: "gauge",
			Delta: nil,
			Value: &metric.Value,
		}
		jsonGauge, errGauge := json.Marshal(gauge)
		if errGauge != nil {
			err = errors.Join(err, errGauge)
		}

		jsons = append(jsons, jsonGauge)
	}

	counter := Metric{
		ID:    m.PollCount.Name,
		MType: "counter",
		Delta: &m.PollCount.Value,
		Value: nil,
	}

	jsonCounter, errCounter := json.Marshal(counter)
	if errCounter != nil {
		err = errors.Join(err, errCounter)
	}

	jsons = append(jsons, jsonCounter)

	if err != nil {
		return nil, err
	}

	return jsons, nil
}

func (*MetricsStore) sendRequest(urls []string) {
	const contentType = "text/plain"
	var client = &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       clientTimeout,
	}

	for _, urlMetric := range urls {
		request, err := http.NewRequest(http.MethodPost, urlMetric, http.NoBody) //nolint:noctx //TODO: добавить контекст
		if err != nil {
			log.Error("Failed to create request",
				log.ErrAttr(err),
				log.StringAttr("url", urlMetric))
		}

		request.Header.Set("Content-Type", contentType)

		response, err := client.Do(request)
		if err != nil {
			log.Error("Failed to send request",
				log.ErrAttr(err),
				log.StringAttr("url", urlMetric),
			)

			continue
		}

		if response.StatusCode == http.StatusServiceUnavailable || response.StatusCode == http.StatusNotFound {
			log.Error("server returned unexpected status code after sending url",
				log.StringAttr("status", response.Status))
		}

		_, err = io.Copy(io.Discard, response.Body)
		if err != nil {
			log.Error("Failed to send request with body to discard",
				log.ErrAttr(err))
		}

		err = response.Body.Close()
		if err != nil {
			log.Error("Failed to close response body",
				log.ErrAttr(err))
		}
	}
}

func (m *MetricsStore) sendRequestJSON(cfg config.Transmitter, jsons [][]byte) {
	const contentType = "application/json"
	var client = &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       clientTimeout,
	}

	for _, jsonMetric := range jsons {
		request, err := http.NewRequest(http.MethodPost, baseProtocol+cfg.Address.String()+"/update/", bytes.NewReader(jsonMetric)) //nolint:noctx //TODO: добавить контекст
		if err != nil {
			log.Error("Failed to create request",
				log.ErrAttr(err),
				log.StringAttr("url", string(jsonMetric)))
		}

		request.Header.Set("Content-Type", contentType)

		response, err := client.Do(request)
		if err != nil {
			log.Error("Failed to send request with body",
				log.ErrAttr(err),
				log.StringAttr("json", string(jsonMetric)),
			)

			continue
		}

		if response.StatusCode == http.StatusServiceUnavailable || response.StatusCode == http.StatusNotFound {
			log.Error("server returned unexpected status code after sending json",
				log.StringAttr("status", response.Status))
		}

		_, err = io.Copy(io.Discard, response.Body)
		if err != nil {
			log.Error("Failed to send request with body to discard",
				log.ErrAttr(err))
		}

		err = response.Body.Close()
		if err != nil {
			log.Error("Failed to close response body",
				log.ErrAttr(err))
		}
	}
}
