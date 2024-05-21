package receiver

import (
	"context"
	"fmt"

	"metrics/config"
	"metrics/internal/service"
	"metrics/internal/store"
)

func Run(cfg config.ReceiverConfig) error {
	db := store.NewMemoryStore()

	receiver := service.NewReceiverService(db, cfg)

	handler := NewHandler(receiver)

	if err := RunServer(context.Background(), handler, cfg); err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}
