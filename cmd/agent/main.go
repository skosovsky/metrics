package main

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"

	"metrics/config"
	"metrics/internal/transmitter"
	log "metrics/pkg/logger"
)

func main() {
	loggerInit()
	loadEnv()
	logAppInfo()

	cfg, err := config.NewTransmitterConfig()
	if err != nil {
		log.Fatal("cfg", log.ErrAttr(err))
	}
	log.Info("config", log.AnyAttr("cfg", fmt.Sprint(cfg)))

	tickerReport := time.NewTicker(cfg.ReportInterval)
	defer tickerReport.Stop()

	tickerPool := time.NewTicker(cfg.PollInterval)
	defer tickerPool.Stop()

	statistics := transmitter.NewMetrics(cfg)

	for {
		select {
		case <-tickerPool.C:
			statistics.Update()
			log.Info("Updated metrics", log.AnyAttr("PollCount", statistics.PollCount))
		case <-tickerReport.C:
			statistics.Report()
			statistics.Clear()
			log.Info("Reported metrics")
		}
	}
}

func loggerInit() {
	log.NewLogger(
		log.WithLevel("DEBUG"),
		log.WithAddSource(false),
		log.WithIsJSON(true),
		log.WithMiddleware(false),
		log.WithSetDefault(true))
}

func loadEnv() {
	if os.Getenv("APP_MODE") == "test" || os.Getenv("APP_MODE") == "production" { //nolint:goconst // not applicable
		return
	}

	if err := godotenv.Load(".env"); err != nil {
		workDir, errGetWD := os.Getwd()
		if errGetWD != nil {
			log.Error("Error getting work dir", log.ErrAttr(errGetWD))
		}
		log.Error("Error loading .env file", log.ErrAttr(err), log.StringAttr("work dir", workDir))
		setEnvDefault()
	}
}

func setEnvDefault() { // TODO: Обновить или удалить уже
	cfg := config.TransmitterConfig{} //nolint:exhaustruct // long struct
	cfg.App.Mode = "test"

	_ = os.Setenv("APP_MODE", cfg.App.Mode)

	log.Info("Environment variables set default")
}

func logAppInfo() {
	if os.Getenv("APP_MODE") == "test" || os.Getenv("APP_MODE") == "production" {
		return
	}

	appInfo, err := config.NewAppInfo()
	if err != nil {
		log.Fatal("appInfo", log.ErrAttr(err))
	}
	log.Info("appInfo", log.AnyAttr("app", fmt.Sprint(appInfo)))
}
