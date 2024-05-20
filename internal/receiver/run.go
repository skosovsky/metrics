package receiver

import (
	"context"

	"metrics/config"
	log "metrics/internal/logger"
	"metrics/internal/service"
	"metrics/internal/store"
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
