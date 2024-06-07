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

	router.Post("/update/{$}", h.AddMetricJSON)
	router.Post("/update/{type}/{id}/{value}", h.AddMetric)
	router.Post("/value/{$}", h.GetMetricJSON)
	router.Get("/value/{type}/{id}", h.GetMetric)
	router.Get("/", h.GetAllMetrics)

	router.Post("/", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	return router
}

func (h Handler) AddMetricJSON(w http.ResponseWriter, r *http.Request) {
	var metric service.Metric

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

	switch metric.MetricType {
	case service.MetricCounter:
		_, err = h.service.AddCounter(metric.ID, *metric.Delta)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}
	case service.MetricGauge:
		_, err = h.service.AddGauge(metric.ID, *metric.Value)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}
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
	metricType := r.PathValue("type")

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	valueString := r.PathValue("value")
	if valueString == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	switch metricType {
	case service.MetricCounter:
		value, err := strconv.ParseInt(valueString, 10, 64)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}

		_, err = h.service.AddCounter(id, value)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}
	case service.MetricGauge:
		value, err := strconv.ParseFloat(valueString, 64)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}

		_, err = h.service.AddGauge(id, value)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (h Handler) GetMetric(w http.ResponseWriter, r *http.Request) {
	metricType := r.PathValue("type")

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	switch metricType {
	case service.MetricCounter:
		counter, err := h.service.GetMetric(id)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		_, err = io.WriteString(w, strconv.FormatInt(*counter.Delta, 10))
		if err != nil {
			log.Error("Error writing response", //nolint:contextcheck // no ctx
				log.ErrAttr(err))

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}
	case service.MetricGauge:
		gauge, err := h.service.GetMetric(id)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}

		gaugeValue := strconv.FormatFloat(*gauge.Value, 'f', -1, 64)

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
	var metric service.Metric

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

	switch metric.MetricType {
	case service.MetricCounter:
		var counter service.Metric
		counter, err = h.service.GetMetric(metric.ID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}

		metric.Delta = counter.Delta

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
	case service.MetricGauge:
		var gauge service.Metric
		gauge, err = h.service.GetMetric(metric.ID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}

		metric.Value = gauge.Value

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
	metrics := h.service.GetAllMetrics()
	if len(metrics) == 0 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	answer, err := h.prepareAllMetrics(metrics)
	if err != nil {
		log.Error("error prepare all metrics", //nolint:contextcheck // false positive
			log.ErrAttr(err))

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

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

func (Handler) prepareAllMetrics(metrics []service.Metric) (string, error) {
	var answer strings.Builder

	for _, metric := range metrics {
		_, _ = answer.WriteString(metric.ID) // always returns nil error
		_, _ = answer.WriteString(" ")       // always returns nil error

		switch metric.MetricType {
		case service.MetricCounter:
			_, _ = answer.WriteString(strconv.FormatInt(*metric.Delta, 10)) // always returns nil error
		case service.MetricGauge:
			_, _ = answer.WriteString(strconv.FormatFloat(*metric.Value, 'f', -1, 64)) // always returns nil error
		default:
			return "", service.ErrUnknownMetricType
		}

		_, _ = answer.WriteString("\n") // always returns nil error
	}

	return answer.String(), nil
}

func (Handler) IsValidRequest(metric service.Metric) bool {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(metric)

	return err == nil
}
