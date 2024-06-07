package consumer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"metrics/config"
	"metrics/internal/log"
)

const (
	ReadTimeout  = 10 * time.Second
	WriteTimeout = 10 * time.Second
	IdleTimeout  = 60 * time.Second
)

func RunServer(_ context.Context, handler Handler, cfg config.ConsumerConfig) error {
	server := http.Server{
		Addr:                         string(cfg.Consumer.Address),
		Handler:                      handler.InitRoutes(),
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
		ConnContext:                  nil,
	}

	log.Info("server starting", //nolint:contextcheck // no ctx
		log.StringAttr("host:port", string(cfg.Consumer.Address)))

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}

	err := server.Close()
	if err != nil {
		return fmt.Errorf("could not close server: %w", err)
	}

	return nil
}
