package receiver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"metrics/config"
	log "metrics/pkg/logger"
)

const (
	ReadTimeout  = 60 * time.Second
	WriteTimeout = 60 * time.Second
	IdleTimeout  = 60 * time.Second
)

func Handler() http.Handler {
	router := chi.NewRouter()
	router.Post("/update/{kind}/{name}/{value}", AddMetric)
	router.Get("/value/{kind}/{name}", GetMetric)
	router.Get("/", GetAllMetrics)

	return router
}

func RunServer(ctx context.Context, cfg config.ReceiverConfig) error {
	// hostPort := cfg.Receiver.Host + ":" + strconv.Itoa(cfg.Receiver.Port)
	server := http.Server{
		Addr:                         cfg.Address,
		Handler:                      Handler(),
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

	log.Info("server starting", log.StringAttr("host:port", cfg.Address)) //nolint:contextcheck // false positive
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("could not start server: %w", err)
	}

	return nil
}

func HandlerMux() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc(http.MethodPost+" /update/{kind}/{name}/{value}", AddMetric)

	return router
}
