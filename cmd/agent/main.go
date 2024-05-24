package main

import (
	"time"

	"metrics/internal/transmitter"
	log "metrics/pkg/logger"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

func main() {
	loggerInit()

	tickerReport := time.NewTicker(reportInterval)
	defer tickerReport.Stop()

	tickerPool := time.NewTicker(pollInterval)
	defer tickerPool.Stop()

	statistics := transmitter.NewMetrics()

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
