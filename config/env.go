package config

import (
	"os"

	"github.com/joho/godotenv"

	log "metrics/pkg/logger"
)

func LoadTransmitterEnv() {
	if os.Getenv("APP_MODE") == testMode || os.Getenv("APP_MODE") == prodMode {
		return
	}

	if err := godotenv.Load(".env"); err != nil {
		workDir, errGetWD := os.Getwd()

		if errGetWD != nil {
			log.Error("Error getting work dir",
				log.ErrAttr(errGetWD))
		}

		log.Error("Error loading .env file",
			log.StringAttr("work dir", workDir),
			log.ErrAttr(err),
		)

		setTransmitterEnvDefault()
	}
}

func setTransmitterEnvDefault() {
	var cfg TransmitterConfig
	cfg.App.Mode = testMode

	err := os.Setenv("APP_MODE", cfg.App.Mode)
	if err != nil {
		log.Error("Error setting APP_MODE", log.ErrAttr(err))
	}

	log.Info("Environment variables set default",
		log.StringAttr("app mode", cfg.App.Mode))
}

func LoadReceiverEnv() {
	if os.Getenv("APP_MODE") == testMode || os.Getenv("APP_MODE") == prodMode {
		return
	}

	if err := godotenv.Load(".env"); err != nil {
		workDir, errGetWD := os.Getwd()

		if errGetWD != nil {
			log.Error("Error getting work dir",
				log.ErrAttr(errGetWD))
		}

		log.Error("Error loading .env file",
			log.StringAttr("work dir", workDir),
			log.ErrAttr(err),
		)

		setReceiverEnvDefault()
	}
}

func setReceiverEnvDefault() {
	var cfg ReceiverConfig
	cfg.App.Mode = testMode
	cfg.Store.DBDriver = "memory"
	cfg.Store.DBAddress = "map"

	err := os.Setenv("APP_MODE", cfg.App.Mode)
	if err != nil {
		log.Error("Error setting APP_MODE",
			log.ErrAttr(err))
	}

	err = os.Setenv("DB_DRIVER", cfg.Store.DBDriver)
	if err != nil {
		log.Error("Error setting DB_DRIVER",
			log.ErrAttr(err))
	}

	err = os.Setenv("DB_ADDRESS", cfg.Store.DBAddress)
	if err != nil {
		log.Error("Error setting DB_ADDRESS",
			log.ErrAttr(err))
	}

	log.Info("Environment variables set default",
		log.StringAttr("app mode", cfg.App.Mode))
}
