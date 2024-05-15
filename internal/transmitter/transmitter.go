package transmitter

import (
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"metrics/config"
	"metrics/internal/model"
	log "metrics/pkg/logger"
)

const (
	clientTimeout     = 1 * time.Second
	serverHost        = "http://localhost:8080/update"
	criticalErrorRate = 0.5
)

type Metrics struct {
	gauges    []model.Gauge
	PollCount []model.Counter
	config    config.TransmitterConfig
}

func NewMetrics(config config.TransmitterConfig) *Metrics {
	var metrics Metrics
	var memMetrics runtime.MemStats
	runtime.ReadMemStats(&memMetrics)

	metrics.gauges = append(metrics.gauges, model.Gauge{Name: "Alloc", Value: float64(memMetrics.Alloc)},
		model.Gauge{Name: "BuckHashSys", Value: float64(memMetrics.BuckHashSys)},
		model.Gauge{Name: "Frees", Value: float64(memMetrics.Frees)},
		model.Gauge{Name: "GCCPUFraction", Value: memMetrics.GCCPUFraction},
		model.Gauge{Name: "GCSys", Value: float64(memMetrics.GCSys)},
		model.Gauge{Name: "HeapAlloc", Value: float64(memMetrics.HeapAlloc)},
		model.Gauge{Name: "HeapIdle", Value: float64(memMetrics.HeapIdle)},
		model.Gauge{Name: "HeapInuse", Value: float64(memMetrics.HeapInuse)},
		model.Gauge{Name: "LastGC", Value: float64(memMetrics.LastGC)},
		model.Gauge{Name: "Lookups", Value: float64(memMetrics.Lookups)},
		model.Gauge{Name: "MCacheInuse", Value: float64(memMetrics.MCacheInuse)},
		model.Gauge{Name: "MCacheSys", Value: float64(memMetrics.MCacheSys)},
		model.Gauge{Name: "MSpanInuse", Value: float64(memMetrics.MSpanInuse)},
		model.Gauge{Name: "MSpanSys", Value: float64(memMetrics.MSpanSys)},
		model.Gauge{Name: "Mallocs", Value: float64(memMetrics.Mallocs)},
		model.Gauge{Name: "NextGC", Value: float64(memMetrics.NextGC)},
		model.Gauge{Name: "NumForcedGC", Value: float64(memMetrics.NumForcedGC)},
		model.Gauge{Name: "NumGC", Value: float64(memMetrics.NumGC)},
		model.Gauge{Name: "OtherSys", Value: float64(memMetrics.OtherSys)},
		model.Gauge{Name: "PauseTotalNs", Value: float64(memMetrics.PauseTotalNs)},
		model.Gauge{Name: "StackInuse", Value: float64(memMetrics.StackInuse)},
		model.Gauge{Name: "StackSys", Value: float64(memMetrics.StackSys)},
		model.Gauge{Name: "Sys", Value: float64(memMetrics.Sys)},
		model.Gauge{Name: "TotalAlloc", Value: float64(memMetrics.TotalAlloc)},
		model.Gauge{Name: "RandomValue", Value: float64(metrics.randomValueGenerate())})

	metrics.PollCount = append(metrics.PollCount, model.Counter{Name: "PollCount", Value: metrics.incrementPollCount()})
	metrics.config = config

	return &metrics
}

func (m *Metrics) Update() { //nolint:funlen // TODO: куда же тут еще уменьшить?
	var memMetrics runtime.MemStats
	runtime.ReadMemStats(&memMetrics)

	m.PollCount = append(m.PollCount, model.Counter{Name: "PollCount", Value: m.incrementPollCount()})

	for _, gauge := range m.gauges {
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
			gauge.Value = float64(m.randomValueGenerate())
		default:
			log.Warn("Unknown gauge metric: ", log.StringAttr("gauge name", gauge.Name))
		}
	}
}

func (m *Metrics) incrementPollCount() int64 {
	length := len(m.PollCount)
	if length == 0 {
		return 1
	}

	lastMetricValue := m.PollCount[length-1].Value
	newLastMetricValue := lastMetricValue + 1

	return newLastMetricValue
}

func (*Metrics) randomValueGenerate() int64 {
	const maxNum = 999999
	rnd := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec // no sec

	return rnd.Int63n(maxNum)
}

func (m *Metrics) Clear() {
	transmitterConfig := m.config
	*m = *NewMetrics(transmitterConfig)
}

func (m *Metrics) Report() {
	m.sendRequest(m.prepareUrls())
	m.Clear()
}

func (m *Metrics) prepareUrls() []string {
	// hostPort := m.config.Host + ":" + strconv.Itoa(m.config.Port)
	urls := make([]string, 0, len(m.gauges)+len(m.PollCount))

	for _, gauge := range m.gauges {
		url := "http://" + m.config.Address + "/update" + "/gauge/" + gauge.Name + "/" + strconv.FormatFloat(gauge.Value, 'f', -1, 64)
		urls = append(urls, url)
	}

	for _, poll := range m.PollCount {
		if poll.Name != "PollCount" {
			log.Warn("Unknown counter metric: ", log.StringAttr("counter name", poll.Name))

			continue
		}

		url := "http://" + m.config.Address + "/update" + "/counter/" + poll.Name + "/" + strconv.FormatInt(poll.Value, 10)
		urls = append(urls, url)
	}

	return urls
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

	for _, url := range urls {
		response, err := client.Post(url, contentType, http.NoBody) //nolint:noctx //TODO: добавить контекст, прокинуть от запуска
		if err != nil {
			countErr++
			log.Error("Failed to send request",
				log.ErrAttr(err),
				log.StringAttr("url", url),
				log.IntAttr("count errors", countErr),
			)

			return
		}

		err = response.Body.Close()
		if err != nil {
			log.Error("Failed to close response body", log.ErrAttr(err))
		}
	}

	if currentErrorRate := float64(countErr) / float64(len(urls)); currentErrorRate >= criticalErrorRate {
		log.Error("critical error rate alert",
			log.Float64Attr("current error rate", currentErrorRate),
			log.Float64Attr("critical error rate", criticalErrorRate))
	}
}
