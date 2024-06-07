package main

import (
	"metrics/config"
	"metrics/internal/log"
	"metrics/internal/producer"
)

func main() {
	log.Prepare()

	config.LoadProducerEnv()

	err := config.LogAppInfo()
	if err != nil {
		log.Fatal("appInfo",
			log.ErrAttr(err))
	}

	cfg, err := config.NewProducerConfig()
	if err != nil {
		log.Fatal("cfg",
			log.ErrAttr(err))
	}

	log.Info("config",
		log.StringAttr("address", string(cfg.Producer.Address)),
		log.IntAttr("poll interval", cfg.Producer.PollInterval),
		log.IntAttr("report interval", cfg.Producer.ReportInterval),
	)

	producer.Run(cfg)
}
