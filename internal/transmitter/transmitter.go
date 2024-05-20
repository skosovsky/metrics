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
	log "metrics/internal/logger"
	"metrics/internal/model"
)

const (
	clientTimeout     = 1 * time.Second
	criticalErrorRate = 0.5
	baseProtocol      = "http://"
)

type Metrics struct {
	Gauges    []model.Gauge
	PollCount model.Counter
}

func NewMetrics() *Metrics {
	var metrics Metrics

	metrics.Gauges = append(metrics.Gauges,
		model.Gauge{Name: "Alloc", Value: 0.0},
		model.Gauge{Name: "BuckHashSys", Value: 0.0},
		model.Gauge{Name: "Frees", Value: 0.0},
		model.Gauge{Name: "GCCPUFraction", Value: 0.0},
		model.Gauge{Name: "GCSys", Value: 0.0},
		model.Gauge{Name: "HeapAlloc", Value: 0.0},
		model.Gauge{Name: "HeapIdle", Value: 0.0},
		model.Gauge{Name: "HeapInuse", Value: 0.0},
		model.Gauge{Name: "LastGC", Value: 0.0},
		model.Gauge{Name: "Lookups", Value: 0.0},
		model.Gauge{Name: "MCacheInuse", Value: 0.0},
		model.Gauge{Name: "MCacheSys", Value: 0.0},
		model.Gauge{Name: "MSpanInuse", Value: 0.0},
		model.Gauge{Name: "MSpanSys", Value: 0.0},
		model.Gauge{Name: "Mallocs", Value: 0.0},
		model.Gauge{Name: "NextGC", Value: 0.0},
		model.Gauge{Name: "NumForcedGC", Value: 0.0},
		model.Gauge{Name: "NumGC", Value: 0.0},
		model.Gauge{Name: "OtherSys", Value: 0.0},
		model.Gauge{Name: "PauseTotalNs", Value: 0.0},
		model.Gauge{Name: "StackInuse", Value: 0.0},
		model.Gauge{Name: "StackSys", Value: 0.0},
		model.Gauge{Name: "Sys", Value: 0.0},
		model.Gauge{Name: "TotalAlloc", Value: 0.0},
		model.Gauge{Name: "RandomValue", Value: 0.0})

	return &metrics
}

func (m *Metrics) Update() { //nolint:funlen // long func
	var memMetrics runtime.MemStats
	runtime.ReadMemStats(&memMetrics)

	m.PollCount.Value++

	for i := range m.Gauges {
		switch m.Gauges[i].Name {
		case "Alloc":
			m.Gauges[i].Value = float64(memMetrics.Alloc)
		case "BuckHashSys":
			m.Gauges[i].Value = float64(memMetrics.BuckHashSys)
		case "Frees":
			m.Gauges[i].Value = float64(memMetrics.Frees)
		case "GCCPUFraction":
			m.Gauges[i].Value = memMetrics.GCCPUFraction
		case "GCSys":
			m.Gauges[i].Value = float64(memMetrics.GCSys)
		case "HeapAlloc":
			m.Gauges[i].Value = float64(memMetrics.HeapAlloc)
		case "HeapIdle":
			m.Gauges[i].Value = float64(memMetrics.HeapIdle)
		case "HeapInuse":
			m.Gauges[i].Value = float64(memMetrics.HeapInuse)
		case "LastGC":
			m.Gauges[i].Value = float64(memMetrics.LastGC)
		case "Lookups":
			m.Gauges[i].Value = float64(memMetrics.Lookups)
		case "MCacheInuse":
			m.Gauges[i].Value = float64(memMetrics.MCacheInuse)
		case "MCacheSys":
			m.Gauges[i].Value = float64(memMetrics.MCacheSys)
		case "MSpanInuse":
			m.Gauges[i].Value = float64(memMetrics.MSpanInuse)
		case "MSpanSys":
			m.Gauges[i].Value = float64(memMetrics.MSpanSys)
		case "Mallocs":
			m.Gauges[i].Value = float64(memMetrics.Mallocs)
		case "NextGC":
			m.Gauges[i].Value = float64(memMetrics.NextGC)
		case "NumForcedGC":
			m.Gauges[i].Value = float64(memMetrics.NumForcedGC)
		case "NumGC":
			m.Gauges[i].Value = float64(memMetrics.NumGC)
		case "OtherSys":
			m.Gauges[i].Value = float64(memMetrics.OtherSys)
		case "PauseTotalNs":
			m.Gauges[i].Value = float64(memMetrics.PauseTotalNs)
		case "StackInuse":
			m.Gauges[i].Value = float64(memMetrics.StackInuse)
		case "StackSys":
			m.Gauges[i].Value = float64(memMetrics.StackSys)
		case "Sys":
			m.Gauges[i].Value = float64(memMetrics.Sys)
		case "TotalAlloc":
			m.Gauges[i].Value = float64(memMetrics.TotalAlloc)
		case "RandomValue":
			m.Gauges[i].Value = float64(rand.Int()) //nolint:gosec // non secure
		default:
			log.Warn("Unknown gauge metric: ", log.StringAttr("gauge name", m.Gauges[i].Name))
		}
	}
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
