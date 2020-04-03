package main

import (
	"time"

	"github.com/sambarnes/demoware-consumer/metrics"
	log "github.com/sirupsen/logrus"
)

func main() {
	// TODO: use viper for configuration through commandline flags
	log.SetLevel(log.DebugLevel)
	metrics.DemowareMetricsURL = "http://localhost:8080/metrics"

	done := make(chan interface{})
	defer close(done)
	ingestedMetrics := metrics.RunGenerator(done)
	metricStreamsByType := metrics.RunDispatcher(done, ingestedMetrics)

	loadStatsExporter := metrics.HandleLoadAverageMetric(done, metricStreamsByType[metrics.LoadAverageMetric])
	cpuUsageStatsExporter := metrics.HandleCPUUsageMetric(done, metricStreamsByType[metrics.CPUUsageMetric])
	kernelUpgradeStatsExporter := metrics.HandleLastKernelUpgradeMetric(done, metricStreamsByType[metrics.LastKernelUpgradeMetric])
	for {
		select {
		case <-time.After(5 * time.Second):

			if stats, ok := loadStatsExporter.CurrentStats(); ok {
				loadStats := stats.(metrics.LoadStats)
				log.WithFields(log.Fields{
					"max": loadStats.Max,
					"min": loadStats.Min,
				}).Info("Current LoadStats")
			} else {
				log.Error("Failed to read from LoadStats Exporter")
			}

			if stats, ok := cpuUsageStatsExporter.CurrentStats(); ok {
				cpuStats := stats.(metrics.CPUUsageStats)
				log.WithFields(log.Fields{
					"average_cpu_usage": cpuStats.Averages,
				}).Info("Current CPUUsageStats")
			} else {
				log.Error("Failed to read from CPUUsageStats Exporter")
			}

			if stats, ok := kernelUpgradeStatsExporter.CurrentStats(); ok {
				kernelStats := stats.(metrics.KernelUpgradeStats)
				log.WithFields(log.Fields{
					"most_recent_upgrade": kernelStats.MostRecent,
				}).Info("Current KernelUpgradeStats")
			} else {
				log.Error("Failed to read from KernelUpgradeStats Exporter")
			}

		}
	}
}
