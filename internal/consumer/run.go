package consumer

import (
	"context"
	"fmt"

	"metrics/config"
	"metrics/internal/consumer/internal/service"
	"metrics/internal/consumer/internal/store"
)

func Run(cfg config.ConsumerConfig) error {
	db := store.NewMemoryStore()

	consumer := service.NewConsumerService(db, cfg)

	handler := NewHandler(consumer)

	if err := RunServer(context.Background(), handler, cfg); err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}
