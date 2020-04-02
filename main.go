package main

import (
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
	go metrics.HandleLoadAverageMetric(done, metricStreamsByType[metrics.LoadAverageMetric])
	go metrics.HandleCPUUsageMetric(done, metricStreamsByType[metrics.CPUUsageMetric])
	go metrics.HandleLastKernelUpgradeMetric(done, metricStreamsByType[metrics.LastKernelUpgradeMetric])

	select {
	// TODO: run a server for outsiders to query current stats of those handler goroutines,
	//		 or make metrics available to prometheus (if available for the project)
	}
}
