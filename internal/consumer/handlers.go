package consumer

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"

	"metrics/internal/consumer/internal/mux"
	"metrics/internal/consumer/internal/service"
	"metrics/internal/log"
)

const (
	metricCounter = "counter"
	metricGauge   = "gauge"
)

type Metric struct {
	ID    string   `json:"id"              validate:"required"`
	MType string   `json:"type"            validate:"required,oneof=gauge counter"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type Handler struct {
	service service.Consumer
}

func NewHandler(service service.Consumer) Handler {
	return Handler{service: service}
}

func (h Handler) InitRoutes() http.Handler {
	router := mux.NewRouter()

	router.Use(WithLogging)
	router.Use(WithGzipCompress)

	router.Post("/", h.BadRequest)
	router.Post("/update/{$}", h.AddMetricJSON)
	router.Post("/update/{kind}/{name}/{value}", h.AddMetric)
	router.Post("/value/{$}", h.GetMetricJSON)
	router.Get("/value/{kind}/{name}", h.GetMetric)
	router.Get("/", h.GetAllMetrics)

	return router
}

func (h Handler) BadRequest(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func (h Handler) AddMetricJSON(w http.ResponseWriter, r *http.Request) {
	var metric Metric

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("error writing response", //nolint:contextcheck // false positive
			log.ErrAttr(err))

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	if len(body) == 0 {
		log.Debug("empty body") //nolint:contextcheck // false positive

		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	err = json.Unmarshal(body, &metric)
	if err != nil {
		log.Debug("error decode to json", //nolint:contextcheck // false positive
			log.ErrAttr(err))

		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	switch metric.MType {
	case metricCounter:
		_ = h.service.AddCounter(metric.ID, *metric.Delta)
	case metricGauge:
		_ = h.service.AddGauge(metric.ID, *metric.Value)
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Encoding", "gzip") //TODO: костыль
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(metric)
	if err != nil {
		log.Error("error encode to json", //nolint:contextcheck // false positive
			log.ErrAttr(err))

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}
}

func (h Handler) AddMetric(w http.ResponseWriter, r *http.Request) {
	kind := r.PathValue("kind")

	name := r.PathValue("name")
	if name == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	valueString := r.PathValue("value")
	if valueString == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	switch kind {
	case metricCounter:
		value, err := strconv.ParseInt(valueString, 10, 64)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}

		_ = h.service.AddCounter(name, value)

	case metricGauge:
		value, err := strconv.ParseFloat(valueString, 64)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}

		_ = h.service.AddGauge(name, value)

	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (h Handler) GetMetric(w http.ResponseWriter, r *http.Request) {
	kind := r.PathValue("kind")

	name := r.PathValue("name")
	if name == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	switch kind {
	case metricCounter:
		counter, err := h.service.GetCounter(name)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		_, err = io.WriteString(w, strconv.FormatInt(counter.Value, 10))
		if err != nil {
			log.Error("Error writing response", //nolint:contextcheck // no ctx
				log.ErrAttr(err))

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

	case metricGauge:
		gauge, err := h.service.GetGauge(name)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}

		gaugeValue := strconv.FormatFloat(gauge.Value, 'f', -1, 64)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		_, err = io.WriteString(w, gaugeValue)
		if err != nil {
			log.Error("Error writing response", //nolint:contextcheck // no ctx
				log.ErrAttr(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}
}

func (h Handler) GetMetricJSON(w http.ResponseWriter, r *http.Request) { //nolint:funlen // agree
	var metric Metric

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("error writing response", //nolint:contextcheck // false positive
			log.ErrAttr(err))

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	if len(body) == 0 {
		log.Debug("empty body") //nolint:contextcheck // false positive

		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	err = json.Unmarshal(body, &metric)
	if err != nil {
		log.Debug("error decode to json", //nolint:contextcheck // false positive
			log.ErrAttr(err))

		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	defer func(Body io.ReadCloser) { //nolint:contextcheck // false positive
		err = Body.Close()
		if err != nil {
			log.Error("error close body",
				log.ErrAttr(err))
		}
	}(r.Body)

	if !h.IsValidRequest(metric) {
		log.Debug("invalid requestBody", //nolint:contextcheck // false positive
			log.StringAttr("metric", fmt.Sprint(metric)))

		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	switch metric.MType {
	case metricCounter:
		var counter service.Counter
		counter, err = h.service.GetCounter(metric.ID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}

		metric.Delta = &counter.Value

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Encoding", "gzip") //TODO: костыль
		w.WriteHeader(http.StatusOK)

		err = json.NewEncoder(w).Encode(metric)
		if err != nil {
			log.Error("error encode to json", //nolint:contextcheck // false positive
				log.ErrAttr(err))

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}
	case metricGauge:
		var gauge service.Gauge
		gauge, err = h.service.GetGauge(metric.ID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}

		metric.Value = &gauge.Value

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Encoding", "gzip") //TODO: костыль
		w.WriteHeader(http.StatusOK)

		err = json.NewEncoder(w).Encode(metric)
		if err != nil {
			log.Error("error encode to json", //nolint:contextcheck // false positive
				log.ErrAttr(err))

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}
}

func (h Handler) GetAllMetrics(w http.ResponseWriter, _ *http.Request) {
	var answer string

	counters := h.service.GetAllCounters()
	if len(counters) == 0 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	answer = h.prepareAllCounters(counters)

	gauges := h.service.GetAllGauges()
	if len(gauges) == 0 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	answer += h.prepareAllGauges(gauges)

	const templateHTML = `
		<!DOCTYPE html>
		<html lang="en">

		<body>
			<pre>{{.}}</pre>
		</body>
		</html>
	`

	templateWithValues, err := template.New("all metrics template").Parse(templateHTML)
	if err != nil {
		log.Error("error parsing template", //nolint:contextcheck // no ctx
			log.ErrAttr(err))

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Encoding", "gzip") //TODO: костыль
	w.WriteHeader(http.StatusOK)

	if err = templateWithValues.Execute(w, answer); err != nil {
		log.Error("error executing template", //nolint:contextcheck // no ctx
			log.ErrAttr(err))

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}
}

func (Handler) prepareAllCounters(counters []service.Counter) string {
	var answer strings.Builder

	for _, counter := range counters {
		_, _ = answer.WriteString(counter.Name)                         // always returns nil error
		_, _ = answer.WriteString(" ")                                  // always returns nil error
		_, _ = answer.WriteString(strconv.FormatInt(counter.Value, 10)) // always returns nil error
		_, _ = answer.WriteString("\n")                                 // always returns nil error
	}

	return answer.String()
}

func (Handler) prepareAllGauges(gauges []service.Gauge) string {
	var answer strings.Builder

	for _, gauge := range gauges {
		_, _ = answer.WriteString(gauge.Name)                                    // always returns nil error
		_, _ = answer.WriteString(" ")                                           // always returns nil error
		_, _ = answer.WriteString(strconv.FormatFloat(gauge.Value, 'f', -1, 64)) // always returns nil error
		_, _ = answer.WriteString("\n")                                          // always returns nil error
	}

	return answer.String()
}

func (Handler) IsValidRequest(metric Metric) bool {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(metric)

	return err == nil
}
