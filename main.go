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

	dispatcher := metrics.ResultStreamDispatcher{}
	defer dispatcher.Close()

	loadMetricsHandler := &metrics.LoadMetricsHandler{}
	cpuMetricsHandler := &metrics.CPUMetricsHandler{}
	kernelMetricsHandler := &metrics.KernelMetricsHandler{}
	metricSubscriptions := map[metrics.Handler]<-chan interface{}{
		loadMetricsHandler:   dispatcher.Subscribe(metrics.LoadAverageMetric),
		cpuMetricsHandler:    dispatcher.Subscribe(metrics.CPUUsageMetric),
		kernelMetricsHandler: dispatcher.Subscribe(metrics.LastKernelUpgradeMetric),
	}

	done := make(chan interface{})
	defer close(done)
	ingestedMetrics := metrics.RunGenerator(done)
	go dispatcher.Run(done, ingestedMetrics)
	for handler, stream := range metricSubscriptions {
		go metrics.RunMetricStreamHandler(done, stream, handler)
	}

introspectionLoop:
	for {
		select {
		case <-done:
			break introspectionLoop
		case <-time.After(5 * time.Second):
			loadStats := loadMetricsHandler.CurrentStats()
			log.WithFields(log.Fields{
				"n":   loadStats.N,
				"min": loadStats.Min,
				"max": loadStats.Max,
			}).Debug("Current LoadStats")

			cpuStats := cpuMetricsHandler.CurrentStats()
			log.WithFields(log.Fields{
				"n":        cpuStats.N,
				"averages": cpuStats.Averages,
			}).Debug("Current CPUUsageStats")

			kernelStats := kernelMetricsHandler.CurrentStats()
			log.WithFields(log.Fields{
				"n":           kernelStats.N,
				"most_recent": kernelStats.MostRecent,
			}).Debug("Current KernelUpgradeStats")
		}
	}
}
