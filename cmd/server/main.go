package main

import (
	"metrics/config"
	"metrics/internal/consumer"
	"metrics/internal/log"
)

func main() {
	log.Prepare()

	config.LoadConsumerEnv()

	err := config.LogAppInfo()
	if err != nil {
		log.Fatal("appInfo",
			log.ErrAttr(err))
	}

	cfg, err := config.NewConsumerConfig()
	if err != nil {
		log.Fatal("cfg",
			log.ErrAttr(err))
	}

	log.Info("config",
		log.StringAttr("address", string(cfg.Consumer.Address)),
		log.Uint64Attr("store interval", cfg.Store.StoreInterval),
		log.StringAttr("filepath", cfg.Store.FileStoragePath),
		log.BoolAttr("should restore", cfg.Store.ShouldRestore))

	err = consumer.Run(cfg)
	if err != nil {
		log.Fatal("consumer run error",
			log.ErrAttr(err))
	}
}
