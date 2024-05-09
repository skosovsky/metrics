package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"metrics/config"
	"metrics/internal/service"
)

const (
	ReadTimeout  = 60 * time.Second
	WriteTimeout = 60 * time.Second
	IdleTimeout  = 60 * time.Second
)

type KeyServiceCtx struct{}

func addMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

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

func RunServer(ctx context.Context, cfg config.Config) error {
	mux := http.NewServeMux()
	mux.HandleFunc(http.MethodPost+" /update/{kind}/{name}/{value}", addMetric)

	hostPort := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	server := http.Server{
		Addr:                         hostPort,
		Handler:                      mux,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  ReadTimeout,
		ReadHeaderTimeout:            0,
		WriteTimeout:                 WriteTimeout,
		IdleTimeout:                  IdleTimeout,
		MaxHeaderBytes:               0,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     nil,
		BaseContext:                  nil,
		ConnContext: func(_ context.Context, _ net.Conn) context.Context {
			return ctx
		},
	}

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("could not start server: %w", err)
	}

	return nil
}
