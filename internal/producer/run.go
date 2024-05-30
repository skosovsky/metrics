package producer

import (
	"time"

	"metrics/config"
	"metrics/internal/log"
)

func Run(cfg config.ProducerConfig) {
	tickReport := time.NewTicker(time.Duration(cfg.Producer.ReportInterval) * time.Second)
	defer tickReport.Stop()

	tickPool := time.NewTicker(time.Duration(cfg.Producer.PollInterval) * time.Second)
	defer tickPool.Stop()

	stats := NewMetrics()

	for {
		select {
		case <-tickPool.C:
			stats.Update()
			log.Debug("Updated metrics", log.AnyAttr("PollCount", stats.PollCount.Value))
		case <-tickReport.C:
			err := stats.Report(cfg.Producer)
			if err != nil {
				log.Error("report error",
					log.ErrAttr(err))

				continue
			}

			log.Debug("Reported metrics")
		}
	}
}
