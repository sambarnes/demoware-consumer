package metrics

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// StatsExporter is a struct that provides an introspection point for
// other goroutines to check in on the current stats of the given goroutine
type StatsExporter struct {
	requestChannel chan chan interface{}
}

func (s StatsExporter) CurrentStats() (interface{}, bool) {
	responseChannel := make(chan interface{})
	s.requestChannel <- responseChannel
	currentStats, ok := <-responseChannel
	return currentStats, ok
}

// LoadStats keeps track of the min and max load seen
type LoadStats struct {
	n   int
	Min float64
	Max float64
}

// Update determines if the newLoadMetric is the new maximum or minimum and if
// so, changes that value
func (s *LoadStats) Update(newLoadMetric float64) error {
	s.n++
	if s.n == 1 {
		s.Min, s.Max = newLoadMetric, newLoadMetric
	} else if newLoadMetric < s.Min {
		s.Min = newLoadMetric
	} else if s.Max < newLoadMetric {
		s.Max = newLoadMetric
	}
	return nil
}

// HandleLoadAverageMetric handles all Metric payloads of type "load_avg"
// and updates LoadStats
func HandleLoadAverageMetric(done <-chan interface{}, metricStream <-chan interface{}) StatsExporter {
	log.Debug(fmt.Sprintf("Starting %v handler...", LoadAverageMetric))
	processedLogMsg := fmt.Sprintf("Processed %v metric", LoadAverageMetric)

	exporter := StatsExporter{requestChannel: make(chan chan interface{})}
	go func() {
		defer close(exporter.requestChannel)

		stats := LoadStats{}
		for {
			select {
			case <-done:
				return
			case v, ok := <-metricStream:
				if ok == false {
					return
				}
				metric := v.(float64)
				err := stats.Update(metric)
				if err != nil {
					log.Error(err)
					continue
				}
				log.WithFields(log.Fields{
					"newValue": metric,
					"max":      stats.Max,
					"min":      stats.Min,
					"n":        stats.n,
				}).Trace(processedLogMsg)
			case responseChannel, ok := <-exporter.requestChannel:
				if ok == false {
					return
				}
				responseChannel <- stats
			}
		}
	}()

	return exporter
}

// CPUUsageStats keeps track of the running average CPU usage per core
type CPUUsageStats struct {
	n        int
	cpuCount int
	totals   []float64
	Averages []float64
}

// Update calculates the new average CPU usage for each core
func (s *CPUUsageStats) Update(usages []float64) error {
	if 0 < s.n && len(usages) != s.cpuCount {
		// Assumption: constant CPU count for all requests
		return fmt.Errorf("invalid length of usages array: expected %v, got %v", s.cpuCount, len(usages))
	}

	s.n++
	if s.n == 1 {
		s.cpuCount = len(usages)
		s.totals = make([]float64, s.cpuCount)
		s.Averages = make([]float64, s.cpuCount)
	}
	for i, usage := range usages {
		s.totals[i] += usage
		s.Averages[i] = s.totals[i] / float64(s.n)
	}
	return nil
}

// HandleCPUUsageMetric handles all Metric payloads of type "cpu_usage" by
// keeping track of the running average usage for each CPU
func HandleCPUUsageMetric(done <-chan interface{}, metricStream <-chan interface{}) StatsExporter {
	log.Debug(fmt.Sprintf("Starting %v handler...", CPUUsageMetric))
	processedLogMsg := fmt.Sprintf("Processed %v metric", CPUUsageMetric)

	exporter := StatsExporter{requestChannel: make(chan chan interface{})}
	go func() {
		defer close(exporter.requestChannel)

		stats := CPUUsageStats{}
		for {
			select {
			case <-done:
				return
			case v, ok := <-metricStream:
				if ok == false {
					return
				}
				metric := toFloat64Array(v.([]interface{}))
				err := stats.Update(metric)
				if err != nil {
					log.Error(err)
					continue
				}
				log.WithFields(log.Fields{
					"newValues": metric,
					"averages":  stats.Averages,
					"n":         stats.n,
				}).Trace(processedLogMsg)
			case responseChannel, ok := <-exporter.requestChannel:
				if ok == false {
					return
				}
				responseChannel <- stats
			}
		}
	}()

	return exporter
}

// KernelUpgradeStats keeps track of the most recent "last_kernel_upgrade"
// timestamp
type KernelUpgradeStats struct {
	n          int
	MostRecent time.Time
}

// Update takes a new timestamp string and compares it against the current
// most recent upgrade time, replacing it if it's more recent
func (s *KernelUpgradeStats) Update(newTimestamp string) error {
	newTime, err := time.Parse(time.RFC3339, newTimestamp)
	if err != nil {
		return fmt.Errorf("unable to parse newTimestamp: %v", err)
	}

	s.n++
	if newTime.After(s.MostRecent) {
		// Assumption: "keep track of the most recent timestamp" means comparing timestamps rather
		// than just storing the timestamp that was received most recently. Even though in the case
		// of this demo, both would behave the same: https://github.com/juju/demoware/blob/master/main.go#L206
		s.MostRecent = newTime
	}
	return nil
}

// HandleLastKernelUpgradeMetric handles all Metric payloads of type
// "last_kernel_upgrade" by keeping track of the most recent timestamp
func HandleLastKernelUpgradeMetric(done <-chan interface{}, metricStream <-chan interface{}) StatsExporter {
	log.Debug(fmt.Sprintf("Starting %v handler...", LastKernelUpgradeMetric))
	processedLogMsg := fmt.Sprintf("Processed %v metric", LastKernelUpgradeMetric)

	exporter := StatsExporter{requestChannel: make(chan chan interface{})}
	go func() {
		defer close(exporter.requestChannel)

		stats := KernelUpgradeStats{}
		for {
			select {
			case <-done:
				return
			case v, ok := <-metricStream:
				if ok == false {
					return
				}
				metric := v.(string)
				err := stats.Update(metric)
				if err != nil {
					log.Error(err)
					continue
				}
				log.WithFields(log.Fields{
					"newValue":   metric,
					"mostRecent": stats.MostRecent,
					"n":          stats.n,
				}).Trace(processedLogMsg)
			case responseChannel, ok := <-exporter.requestChannel:
				if ok == false {
					return
				}
				responseChannel <- stats
			}
		}
	}()

	return exporter
}
