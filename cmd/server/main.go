package main

import (
	"metrics/config"
	"metrics/internal/receiver"
	log "metrics/pkg/logger"
)

func main() {
	log.Prepare()

	config.LoadReceiverEnv()

	err := config.LogAppInfo()
	if err != nil {
		log.Fatal("appInfo",
			log.ErrAttr(err))
	}

	cfg, err := config.NewReceiverConfig()
	if err != nil {
		log.Fatal("cfg",
			log.ErrAttr(err))
	}

	log.Info("config",
		log.StringAttr("address", string(cfg.Receiver.Address)),
	)

	receiver.Run(cfg)
}
