package producer

import (
	"bytes"
	"compress/gzip"
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
	clientTimeout      = 10 * time.Second
	baseProtocol       = "http://"
	methodCompressGzip = "gzip"
)

const (
	MetricCounter = "counter"
	MetricGauge   = "gauge"
)

var ErrUnknownMetricType = errors.New("unknown metric type")

type (
	Metric struct {
		ID         string   `json:"id"              validate:"required"`
		MetricType string   `json:"type"            validate:"required,oneof=gauge counter"`
		Delta      *int64   `json:"delta,omitempty"`
		Value      *float64 `json:"value,omitempty"`
	}

	MetricsStore struct {
		memory map[string]Metric
	}
)

func NewMetrics() *MetricsStore {
	return &MetricsStore{
		memory: map[string]Metric{},
	}
}

func (m *MetricsStore) Update() { //nolint:funlen // true
	var memMetrics runtime.MemStats
	runtime.ReadMemStats(&memMetrics)

	currentDelta := m.memory["PollCount"].Delta
	if currentDelta == nil {
		currentDelta = new(int64)
	}

	*currentDelta++

	m.memory["PollCount"] = Metric{ID: "PollCount", MetricType: MetricCounter, Value: nil, Delta: currentDelta}

	memMetricsAlloc := float64(memMetrics.Alloc)
	m.memory["Alloc"] = Metric{ID: "Alloc", MetricType: MetricGauge, Value: &memMetricsAlloc, Delta: nil}

	memMetricsBuckHashSys := float64(memMetrics.BuckHashSys)
	m.memory["BuckHashSys"] = Metric{ID: "BuckHashSys", MetricType: MetricGauge, Value: &memMetricsBuckHashSys, Delta: nil}

	memMetricsFrees := float64(memMetrics.Frees)
	m.memory["Frees"] = Metric{ID: "Frees", MetricType: MetricGauge, Value: &memMetricsFrees, Delta: nil}

	memMetricsGCCPUFraction := memMetrics.GCCPUFraction
	m.memory["GCCPUFraction"] = Metric{ID: "GCCPUFraction", MetricType: MetricGauge, Value: &memMetricsGCCPUFraction, Delta: nil}

	memMetricsGCSys := float64(memMetrics.GCSys)
	m.memory["GCSys"] = Metric{ID: "GCSys", MetricType: MetricGauge, Value: &memMetricsGCSys, Delta: nil}

	memMetricsHeapAlloc := float64(memMetrics.HeapAlloc)
	m.memory["HeapAlloc"] = Metric{ID: "HeapAlloc", MetricType: MetricGauge, Value: &memMetricsHeapAlloc, Delta: nil}

	memMetricsHeapIdle := float64(memMetrics.HeapIdle)
	m.memory["HeapIdle"] = Metric{ID: "HeapIdle", MetricType: MetricGauge, Value: &memMetricsHeapIdle, Delta: nil}

	memMetricsHeapInuse := float64(memMetrics.HeapInuse)
	m.memory["HeapInuse"] = Metric{ID: "HeapInuse", MetricType: MetricGauge, Value: &memMetricsHeapInuse, Delta: nil}

	memMetricsHeapObjects := float64(memMetrics.HeapObjects)
	m.memory["HeapObjects"] = Metric{ID: "HeapObjects", MetricType: MetricGauge, Value: &memMetricsHeapObjects, Delta: nil}

	memMetricsHeapReleased := float64(memMetrics.HeapReleased)
	m.memory["HeapReleased"] = Metric{ID: "HeapReleased", MetricType: MetricGauge, Value: &memMetricsHeapReleased, Delta: nil}

	memMetricsHeapSys := float64(memMetrics.HeapSys)
	m.memory["HeapSys"] = Metric{ID: "HeapSys", MetricType: MetricGauge, Value: &memMetricsHeapSys, Delta: nil}

	memMetricsLastGC := float64(memMetrics.LastGC)
	m.memory["LastGC"] = Metric{ID: "LastGC", MetricType: MetricGauge, Value: &memMetricsLastGC, Delta: nil}

	memMetricsLookups := float64(memMetrics.Lookups)
	m.memory["Lookups"] = Metric{ID: "Lookups", MetricType: MetricGauge, Value: &memMetricsLookups, Delta: nil}

	memMetricsMCacheInuse := float64(memMetrics.MCacheInuse)
	m.memory["MCacheInuse"] = Metric{ID: "MCacheInuse", MetricType: MetricGauge, Value: &memMetricsMCacheInuse, Delta: nil}

	memMetricsMCacheSys := float64(memMetrics.MCacheSys)
	m.memory["MCacheSys"] = Metric{ID: "MCacheSys", MetricType: MetricGauge, Value: &memMetricsMCacheSys, Delta: nil}

	memMetricsMSpanInuse := float64(memMetrics.MSpanInuse)
	m.memory["MSpanInuse"] = Metric{ID: "MSpanInuse", MetricType: MetricGauge, Value: &memMetricsMSpanInuse, Delta: nil}

	memMetricsMSpanSys := float64(memMetrics.MSpanSys)
	m.memory["MSpanSys"] = Metric{ID: "MSpanSys", MetricType: MetricGauge, Value: &memMetricsMSpanSys, Delta: nil}

	memMetricsMallocs := float64(memMetrics.Mallocs)
	m.memory["Mallocs"] = Metric{ID: "Mallocs", MetricType: MetricGauge, Value: &memMetricsMallocs, Delta: nil}

	memMetricsNextGC := float64(memMetrics.NextGC)
	m.memory["NextGC"] = Metric{ID: "NextGC", MetricType: MetricGauge, Value: &memMetricsNextGC, Delta: nil}

	memMetricsNumForcedGC := float64(memMetrics.NumForcedGC)
	m.memory["NumForcedGC"] = Metric{ID: "NumForcedGC", MetricType: MetricGauge, Value: &memMetricsNumForcedGC, Delta: nil}

	memMetricsNumGC := float64(memMetrics.NumGC)
	m.memory["NumGC"] = Metric{ID: "NumGC", MetricType: MetricGauge, Value: &memMetricsNumGC, Delta: nil}

	memMetricsOtherSys := float64(memMetrics.OtherSys)
	m.memory["OtherSys"] = Metric{ID: "OtherSys", MetricType: MetricGauge, Value: &memMetricsOtherSys, Delta: nil}

	memMetricsPauseTotalNs := float64(memMetrics.PauseTotalNs)
	m.memory["PauseTotalNs"] = Metric{ID: "PauseTotalNs", MetricType: MetricGauge, Value: &memMetricsPauseTotalNs, Delta: nil}

	memMetricsStackInuse := float64(memMetrics.StackInuse)
	m.memory["StackInuse"] = Metric{ID: "StackInuse", MetricType: MetricGauge, Value: &memMetricsStackInuse, Delta: nil}

	memMetricsStackSys := float64(memMetrics.StackSys)
	m.memory["StackSys"] = Metric{ID: "StackSys", MetricType: MetricGauge, Value: &memMetricsStackSys, Delta: nil}

	memMetricsSys := float64(memMetrics.Sys)
	m.memory["Sys"] = Metric{ID: "Sys", MetricType: MetricGauge, Value: &memMetricsSys, Delta: nil}

	memMetricsTotalAlloc := float64(memMetrics.TotalAlloc)
	m.memory["TotalAlloc"] = Metric{ID: "TotalAlloc", MetricType: MetricGauge, Value: &memMetricsTotalAlloc, Delta: nil}

	randomValue := float64(rand.Int()) //nolint:gosec // i know
	m.memory["RandomValue"] = Metric{ID: "RandomValue", MetricType: MetricGauge, Value: &randomValue, Delta: nil}
}

func (m *MetricsStore) Report(cfg config.Producer) error {
	err := m.reportURL(cfg)
	if err != nil {
		return fmt.Errorf("reporting url metrics: %w", err)
	}

	err = m.reportJSON(cfg)
	if err != nil {
		return fmt.Errorf("reporting json metrics: %w", err)
	}

	delete(m.memory, "PollCount")

	return nil
}

func (m *MetricsStore) reportURL(cfg config.Producer) error {
	urls, err := m.prepareURLs(cfg)
	if err != nil {
		return fmt.Errorf("prepare urls: %w", err)
	}

	m.sendRequest(urls)

	return nil
}

func (m *MetricsStore) reportJSON(cfg config.Producer) error {
	jsons, err := m.prepareJSONs()
	if err != nil {
		return fmt.Errorf("prepare urls: %w", err)
	}

	m.sendRequestJSON(cfg, jsons)

	return nil
}

func (m *MetricsStore) prepareURLs(cfg config.Producer) ([]string, error) {
	urls := make([]string, 0, len(m.memory))

	for _, metric := range m.memory {
		var err error
		var urlMetric string

		switch metric.MetricType {
		case MetricCounter:
			urlMetric, err = url.JoinPath(baseProtocol+cfg.Address.String(), "update", metric.MetricType, metric.ID, strconv.FormatInt(*metric.Delta, 10))
			if err != nil {
				return nil, fmt.Errorf("prepare urls: %w", err)
			}
		case MetricGauge:
			urlMetric, err = url.JoinPath(baseProtocol+cfg.Address.String(), "update", metric.MetricType, metric.ID, strconv.FormatFloat(*metric.Value, 'f', -1, 64))
			if err != nil {
				return nil, fmt.Errorf("prepare urls: %w", err)
			}
		default:
			return nil, ErrUnknownMetricType
		}

		urls = append(urls, urlMetric)
	}

	return urls, nil
}

func (m *MetricsStore) prepareJSONs() ([][]byte, error) {
	jsons := make([][]byte, 0, len(m.memory))

	for _, metric := range m.memory {
		jsonGauge, err := json.Marshal(metric)
		if err != nil {
			return nil, fmt.Errorf("prepare jsons: %w", err)
		}

		jsons = append(jsons, jsonGauge)
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
		request.Header.Del("Accept-Encoding")

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

func (*MetricsStore) compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	gzipWriter := gzip.NewWriter(&buf)

	if _, err := gzipWriter.Write(data); err != nil {
		return nil, fmt.Errorf("compressing data: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("closing gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

func (m *MetricsStore) sendRequestJSON(cfg config.Producer, jsons [][]byte) {
	const contentType = "application/json"
	var client = &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       clientTimeout,
	}

	for _, jsonMetric := range jsons {
		compressedJSONMetric, err := m.compress(jsonMetric)
		if err != nil {
			log.Error("Failed to compress json",
				log.ErrAttr(err),
				log.JSONAttr("url", jsonMetric))

			continue
		}

		request, err := http.NewRequest(http.MethodPost, baseProtocol+cfg.Address.String()+"/update/", bytes.NewReader(compressedJSONMetric)) //nolint:noctx //TODO: добавить контекст
		if err != nil {
			log.Error("Failed to create request",
				log.ErrAttr(err),
				log.StringAttr("url", string(jsonMetric)))

			continue
		}

		request.Header.Set("Content-Type", contentType)
		request.Header.Set("Content-Encoding", methodCompressGzip)

		response, err := client.Do(request)
		if err != nil {
			log.Error("Failed to send request with body",
				log.ErrAttr(err),
				log.StringAttr("json", string(jsonMetric)))

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
