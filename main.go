package main

import (
	"github.com/sambarnes/demoware-consumer/metrics"
	log "github.com/sirupsen/logrus"
)

func main() {
	// TODO: use viper for configuration through commandline flags
	log.SetLevel(log.DebugLevel)
	metrics.DemowareMetricsURL = "http://localhost:8080/metrics"

	metricHandlers := map[metrics.MetricType]func(done <-chan interface{}, metricStream <-chan interface{}){
		metrics.LoadAverageMetric:       metrics.HandleLoadAverageMetric,
		metrics.CPUUsageMetric:          metrics.HandleCPUUsageMetric,
		metrics.LastKernelUpgradeMetric: metrics.HandleLastKernelUpgradeMetric,
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
		go handler(done, metricStreamsByType[metricType])
	}

	select {
	// TODO: run a server for outsiders to query current stats of those handler goroutines,
	//		 or make metrics available to prometheus (if available for the project)
	}
}
