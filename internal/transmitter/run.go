package transmitter

import (
	"time"

	"metrics/config"
	log "metrics/pkg/logger"
)

func Run(cfg config.TransmitterConfig) {
	tickReport := time.NewTicker(time.Duration(cfg.Transmitter.ReportInterval) * time.Second)
	defer tickReport.Stop()

	tickPool := time.NewTicker(time.Duration(cfg.Transmitter.PollInterval) * time.Second)
	defer tickPool.Stop()

	stats := NewMetrics()

	for {
		select {
		case <-tickPool.C:
			stats.Update()
			log.Debug("Updated metrics", log.AnyAttr("PollCount", stats.PollCount.Value))
		case <-tickReport.C:
			err := stats.Report(cfg.Transmitter)
			if err != nil {
				log.Error("report error",
					log.ErrAttr(err))

				continue
			}

			log.Debug("Reported metrics")
		}
	}
}
