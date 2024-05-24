package receiver

import (
	"net/http"
	"strconv"

	"metrics/internal/service"
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

	kind := r.PathValue("kind")

	name := r.PathValue("name")
	if name == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	valueString := r.PathValue("value")
	if valueString == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
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

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
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

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}
