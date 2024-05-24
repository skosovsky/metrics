package main

import (
	"metrics/config"
	"metrics/internal/log"
	"metrics/internal/receiver"
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

	err = receiver.Run(cfg)
	if err != nil {
		log.Fatal("receiver run error",
			log.ErrAttr(err))
	}
}
