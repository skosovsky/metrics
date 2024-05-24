package receiver

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"metrics/internal/model"
	"metrics/internal/service"
	log "metrics/pkg/logger"
)

type KeyServiceCtx struct{}

func AddMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	metricsGetter, ok := r.Context().Value(KeyServiceCtx{}).(service.MetricsGetter)
	if !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	kind := chi.URLParam(r, "kind") // for mux: kind := r.PathValue("kind")

	name := chi.URLParam(r, "name") // for mux: name := r.PathValue("name")
	if name == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	valueString := chi.URLParam(r, "value") // for mux: valueString := r.PathValue("value")
	if valueString == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	switch kind {
	case "counter":
		value, err := strconv.ParseInt(valueString, 10, 64)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}

		_, err = metricsGetter.AddCounter(name, value)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}
	case "gauge":
		value, err := strconv.ParseFloat(valueString, 64)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}

		_, err = metricsGetter.AddGauge(name, value)
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

func GetMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	metricsGetter, ok := r.Context().Value(KeyServiceCtx{}).(service.MetricsGetter)
	if !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	kind := chi.URLParam(r, "kind") // for mux: kind := r.PathValue("kind")

	name := chi.URLParam(r, "name") // for mux: name := r.PathValue("name")
	if name == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	switch kind {
	case "counter":
		counters, err := metricsGetter.GetCounters(name)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}

		var counterValues int64

		for _, counter := range counters {
			counterValues += counter.Value
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		_, err = io.WriteString(w, strconv.Itoa(int(counterValues)))
		if err != nil {
			log.Error("Error writing response", log.ErrAttr(err)) //nolint:contextcheck // false positive
		}

	case "gauge":
		gauge, err := metricsGetter.GetGauge(name)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}

		gaugeValue := strconv.FormatFloat(gauge.Value, 'f', -1, 64)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		_, err = io.WriteString(w, gaugeValue)
		if err != nil {
			log.Error("Error writing response", log.ErrAttr(err)) //nolint:contextcheck // false positive
		}
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}
}

func GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	metricsGetter, ok := r.Context().Value(KeyServiceCtx{}).(service.MetricsGetter)
	if !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	var answer string

	counters := metricsGetter.GetAllCounters()
	if len(counters) == 0 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	answer = prepareAllCounters(counters)

	gauges := metricsGetter.GetAllGauges()
	if len(gauges) == 0 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	answer += prepareAllGauges(gauges)

	_, err := io.WriteString(w, answer)
	if err != nil {
		log.Error("Error writing response", log.ErrAttr(err)) //nolint:contextcheck // false positive
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func prepareAllCounters(counters [][]model.Counter) string {
	var answer string

	for _, counter := range counters {
		for _, counterUnit := range counter {
			answer += counterUnit.Name + " " + strconv.Itoa(int(counterUnit.Value)) + "\n"
		}
	}

	return answer
}

func prepareAllGauges(gauges []model.Gauge) string {
	var answer string

	for _, gauge := range gauges {
		answer += gauge.Name + " " + strconv.FormatFloat(gauge.Value, 'f', -1, 64) + "\n"
	}

	return answer
}
