package receiver

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	log "metrics/internal/logger"
	"metrics/internal/model"
	"metrics/internal/service"
)

type Handler struct {
	service service.Receiver
}

func NewHandler(service service.Receiver) Handler {
	return Handler{service: service}
}

func (h Handler) InitRoutes() http.Handler {
	router := chi.NewRouter()

	router.Post("/update/{kind}/{name}/{value}", h.AddMetric)
	router.Get("/value/{kind}/{name}", h.GetMetric)
	router.Get("/", h.GetAllMetrics)

	return router
}

func (h Handler) AddMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	kind := chi.URLParam(r, "kind")

	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	valueString := chi.URLParam(r, "value")
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

		_ = h.service.AddCounter(name, value)

	case "gauge":
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
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	kind := chi.URLParam(r, "kind")

	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	switch kind {
	case "counter":
		counter, err := h.service.GetCounter(name)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		_, err = io.WriteString(w, strconv.FormatInt(counter.Value, 10))
		if err != nil {
			log.Error("Error writing response", log.ErrAttr(err)) //nolint:contextcheck // false positive
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

	case "gauge":
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
			log.Error("Error writing response", log.ErrAttr(err)) //nolint:contextcheck // false positive
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}
}

func (h Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

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

	_, err := io.WriteString(w, answer)
	if err != nil {
		log.Error("Error writing response", log.ErrAttr(err)) //nolint:contextcheck // false positive
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (Handler) prepareAllCounters(counters []model.Counter) string {
	var answer string

	for _, counter := range counters {
		answer += counter.Name + " " + strconv.FormatInt(counter.Value, 10) + "\n"
	}

	return answer
}

func (Handler) prepareAllGauges(gauges []model.Gauge) string {
	var answer string

	for _, gauge := range gauges {
		answer += gauge.Name + " " + strconv.FormatFloat(gauge.Value, 'f', -1, 64) + "\n"
	}

	return answer
}
