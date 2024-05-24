package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	_ "modernc.org/sqlite"

	"metrics/config"
	"metrics/internal/server"
	"metrics/internal/service"
	"metrics/internal/store"
	log "metrics/pkg/logger"
)

func main() {
	loggerInit()
	loadEnv()
	logAppInfo()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("cfg", log.ErrAttr(err))
	}
	log.Info("config", log.AnyAttr("cfg", fmt.Sprint(cfg)))

	db, err := store.NewMemoryStore() // add defer db.CloseDBStore() - only for sqlite3
	if err != nil {
		log.Fatal("store", log.ErrAttr(err))
	}

	metricsGetter := service.NewMetricsGetterService(db)

	ctx := context.WithValue(context.Background(), server.KeyServiceCtx{}, metricsGetter)

	if err = server.RunServer(ctx, cfg); err != nil {
		log.Fatal("run", log.ErrAttr(err))
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

func setEnvDefault() {
	cfg := config.Config{} //nolint:exhaustruct // long struct
	cfg.App.Mode = "test"
	cfg.Server.Host = "localhost"
	cfg.Server.Port = 8080
	cfg.Store.DBDriver = "memory"
	cfg.Store.DBAddress = "map"

	_ = os.Setenv("APP_MODE", cfg.App.Mode)
	_ = os.Setenv("SRV_HOST", cfg.Server.Host)
	_ = os.Setenv("SRV_PORT", strconv.Itoa(cfg.Server.Port))
	_ = os.Setenv("DB_DRIVER", cfg.Store.DBDriver)
	_ = os.Setenv("DB_ADDRESS", cfg.Store.DBAddress)

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
