package main

import (
	"fmt"
	"github.com/sambarnes/demoware-consumer/metrics"
	log "github.com/sirupsen/logrus"
	"time"
)

func main() {
	// TODO: use viper for configuration through commandline flags
	log.SetLevel(log.DebugLevel)
	metrics.DemowareMetricsURL = "http://localhost:8080/metrics"

	loadMetricsHandler := &metrics.LoadMetricsHandler{}
	cpuMetricsHandler := &metrics.CPUMetricsHandler{}
	kernelMetricsHandler := &metrics.KernelMetricsHandler{}
	metricHandlers := map[metrics.MetricType]metrics.Handler{
		metrics.LoadAverageMetric:       loadMetricsHandler,
		metrics.CPUUsageMetric:          cpuMetricsHandler,
		metrics.LastKernelUpgradeMetric: kernelMetricsHandler,
	}
	var metricSubscriptions []metrics.MetricType
	for k, _ := range metricHandlers {
		metricSubscriptions = append(metricSubscriptions, k)
	}

	done := make(chan interface{})
	defer close(done)
	ingestedMetrics := metrics.RunGenerator(done)
	metricStreamsByType := metrics.RunDispatcher(done, metricSubscriptions, ingestedMetrics)
	for metricType, handler := range metricHandlers {
		log.Debug(fmt.Sprintf("Starting %v handler...", metricType))
		go metrics.RunMetricStreamHandler(done, metricStreamsByType[metricType], handler)
	}

introspectionLoop:
	for {
		select {
		case <-done:
			break introspectionLoop
		case <-time.After(5 * time.Second):
			//
			loadStats := loadMetricsHandler.CurrentStats()
			log.WithFields(log.Fields{
				"min": loadStats.Min,
				"max": loadStats.Max,
			}).Debug("Current LoadStats")

			cpuStats := cpuMetricsHandler.CurrentStats()
			log.WithFields(log.Fields{
				"averages": cpuStats.Averages,
			}).Debug("Current CPUUsageStats")

			kernelStats := kernelMetricsHandler.CurrentStats()
			log.WithFields(log.Fields{
				"most_recent": kernelStats.MostRecent,
			}).Debug("Current KernelUpgradeStats")
		}
	}
}
