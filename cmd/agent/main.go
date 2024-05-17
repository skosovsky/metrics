package main

import (
	"metrics/config"
	"metrics/internal/transmitter"
	log "metrics/pkg/logger"
)

func main() {
	log.Prepare()

	config.LoadTransmitterEnv()

	err := config.LogAppInfo()
	if err != nil {
		log.Fatal("appInfo",
			log.ErrAttr(err))
	}

	cfg, err := config.NewTransmitterConfig()
	if err != nil {
		log.Fatal("cfg",
			log.ErrAttr(err))
	}

	log.Info("config",
		log.StringAttr("address", string(cfg.Transmitter.Address)),
		log.IntAttr("poll interval", cfg.Transmitter.PollInterval),
		log.IntAttr("report interval", cfg.Transmitter.ReportInterval),
	)

	transmitter.Run(cfg)
}
