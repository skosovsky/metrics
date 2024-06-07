package consumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"metrics/config"
	"metrics/internal/consumer/internal/service"
	"metrics/internal/consumer/internal/store"
	"metrics/internal/log"
)

func Run(cfg config.ConsumerConfig) error {
	var db service.Store
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.Store.FileStoragePath == "" || cfg.Store.StoreInterval != 0 {
		db, err = store.NewMemoryStore(cfg.Store)
		if err != nil {
			return fmt.Errorf("create memory store: %w", err)
		}

		go autosave(ctx, cfg.Store, db)
	} else {
		db, err = store.NewFileStore(cfg.Store)
		if err != nil {
			return fmt.Errorf("create file store: %w", err)
		}

		defer db.Close()
	}

	go gracefulShutdown(ctx, cancel, cfg.Store, db)

	consumer := service.NewConsumerService(db, cfg)

	handler := NewHandler(consumer)

	if err = RunServer(ctx, handler, cfg); err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}

func autosave(ctx context.Context, cfg config.Store, db service.Store) {
	tickSave := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
	defer tickSave.Stop()

	go func(ctx context.Context) {
		<-ctx.Done()

		tickSave.Stop()
	}(ctx)

	for range tickSave.C {
		err := saveAll(cfg, db) //nolint:contextcheck // no ctx
		if err != nil {
			log.Error("error saving data", //nolint:contextcheck // no ctx
				log.ErrAttr(err))
		}
	}
}

func saveAll(cfg config.Store, db service.Store) error {
	fileStore, err := store.NewFileStore(cfg)
	if err != nil {
		return fmt.Errorf("create file store: %w", err)
	}

	defer fileStore.Close()

	metrics := db.GetAllMetrics()

	for _, metric := range metrics {
		switch metric.MetricType {
		case service.MetricCounter:
			err = fileStore.AddCounter(metric, false)
			if err != nil {
				log.Error("add counter to file store",
					log.ErrAttr(err))
			}
		case service.MetricGauge:
			err = fileStore.AddGauge(metric)
			if err != nil {
				log.Error("add gauge to file store",
					log.ErrAttr(err))
			}
		default:
			log.Error("unknown metric type",
				log.ErrAttr(service.ErrUnknownMetricType))
		}
	}

	log.Debug("All metrics saved")

	return nil
}

func gracefulShutdown(ctx context.Context, cancel context.CancelFunc, cfg config.Store, db service.Store) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigs
		cancel()
	}()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()

		err := saveAll(cfg, db) //nolint:contextcheck // no ctx
		if err != nil {
			log.Error("error saving data", //nolint:contextcheck // no ctx
				log.ErrAttr(err))
		}
	}()

	wg.Wait()
}
