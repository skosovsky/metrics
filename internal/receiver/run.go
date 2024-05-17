package receiver

import (
	"context"

	"metrics/config"
	"metrics/internal/service"
	"metrics/internal/store"
	log "metrics/pkg/logger"
)

func Run(cfg config.ReceiverConfig) {
	db := store.NewMemoryStore()

	receiver := service.NewReceiverService(db, cfg)

	handler := NewHandler(receiver)

	if err := RunServer(context.Background(), handler, cfg); err != nil {
		log.Fatal("run server",
			log.ErrAttr(err))
	}
}
