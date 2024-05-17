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
	"metrics/internal/model"
	log "metrics/pkg/logger"
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

	for _, gauge := range m.Gauges {
		switch gauge.Name {
		case "Alloc":
			gauge.Value = float64(memMetrics.Alloc)
		case "BuckHashSys":
			gauge.Value = float64(memMetrics.BuckHashSys)
		case "Frees":
			gauge.Value = float64(memMetrics.Frees)
		case "GCCPUFraction":
			gauge.Value = memMetrics.GCCPUFraction
		case "GCSys":
			gauge.Value = float64(memMetrics.GCSys)
		case "HeapAlloc":
			gauge.Value = float64(memMetrics.HeapAlloc)
		case "HeapIdle":
			gauge.Value = float64(memMetrics.HeapIdle)
		case "HeapInuse":
			gauge.Value = float64(memMetrics.HeapInuse)
		case "LastGC":
			gauge.Value = float64(memMetrics.LastGC)
		case "Lookups":
			gauge.Value = float64(memMetrics.Lookups)
		case "MCacheInuse":
			gauge.Value = float64(memMetrics.MCacheInuse)
		case "MCacheSys":
			gauge.Value = float64(memMetrics.MCacheSys)
		case "MSpanInuse":
			gauge.Value = float64(memMetrics.MSpanInuse)
		case "MSpanSys":
			gauge.Value = float64(memMetrics.MSpanSys)
		case "Mallocs":
			gauge.Value = float64(memMetrics.Mallocs)
		case "NextGC":
			gauge.Value = float64(memMetrics.NextGC)
		case "NumForcedGC":
			gauge.Value = float64(memMetrics.NumForcedGC)
		case "NumGC":
			gauge.Value = float64(memMetrics.NumGC)
		case "OtherSys":
			gauge.Value = float64(memMetrics.OtherSys)
		case "PauseTotalNs":
			gauge.Value = float64(memMetrics.PauseTotalNs)
		case "StackInuse":
			gauge.Value = float64(memMetrics.StackInuse)
		case "StackSys":
			gauge.Value = float64(memMetrics.StackSys)
		case "Sys":
			gauge.Value = float64(memMetrics.Sys)
		case "TotalAlloc":
			gauge.Value = float64(memMetrics.TotalAlloc)
		case "RandomValue":
			gauge.Value = float64(rand.Int())
		default:
			log.Warn("Unknown gauge metric: ", log.StringAttr("gauge name", gauge.Name))
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
	var countErr int
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
				log.IntAttr("count errors", countErr),
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
