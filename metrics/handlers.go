package metrics

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Handler interface {
	Handle(metric interface{}) error
}

// RunMetricStreamHandler pulls metrics off of the metricStream channel and
// passes them to a handler for processing, stopping when a signal is sent over
// the done channel
func RunMetricStreamHandler(done <-chan interface{}, metricStream <-chan interface{}, handler Handler) {
	for metric := range orDone(done, metricStream) {
		if err := handler.Handle(metric); err != nil {
			log.Error(err)
			continue
		}
	}
}

// LoadMetricsHandler handles all "load_avg" metrics and manages LoadStats
type LoadMetricsHandler struct {
	mu    sync.RWMutex
	stats LoadStats
}

// LoadStats keeps track of the min and max load seen
type LoadStats struct {
	N   int
	Min float64
	Max float64
}

// Handle updates the LoadStats with a new metric
func (h *LoadMetricsHandler) Handle(metric interface{}) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	load, ok := metric.(float64)
	if ok == false {
		return fmt.Errorf("failed to cast metric to float64")
	}
	return h.stats.Update(load)
}

// CurrentStats returns the current LoadStats in a concurrent-safe manner
func (h LoadMetricsHandler) CurrentStats() LoadStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.stats
}

// Update determines if the newLoadMetric is the new maximum or minimum and if
// so, changes that value
func (s *LoadStats) Update(newLoadMetric float64) error {
	s.N++
	if s.N == 1 {
		s.Min, s.Max = newLoadMetric, newLoadMetric
	} else if newLoadMetric < s.Min {
		s.Min = newLoadMetric
	} else if s.Max < newLoadMetric {
		s.Max = newLoadMetric
	}
	return nil
}

// CPUMetricsHandler handles all "cpu_usage" metrics and manages CPUUsageStats
type CPUMetricsHandler struct {
	mu    sync.RWMutex
	stats CPUUsageStats
}

// CPUUsageStats keeps track of the running average CPU usage per core
type CPUUsageStats struct {
	cpuCount int
	totals   []float64
	N        int
	Averages []float64
}

// Handle updates the CPUUsageStats with a new metric
func (h *CPUMetricsHandler) Handle(metric interface{}) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	tmp, ok := metric.([]interface{})
	if ok == false {
		return fmt.Errorf("failed to cast metric to []interface{}")
	}
	usages, err := toFloat64Array(tmp)
	if err != nil {
		return err
	}
	return h.stats.Update(usages)
}

// CurrentStats returns the current CPUUsageStats in a concurrent-safe manner
func (h CPUMetricsHandler) CurrentStats() CPUUsageStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// TODO: evaluate use of a deep copy library
	return CPUUsageStats{
		cpuCount: h.stats.cpuCount,
		totals:   append([]float64{}, h.stats.totals...),
		N:        h.stats.N,
		Averages: append([]float64{}, h.stats.Averages...),
	}
}

// Update calculates the new average CPU usage for each core
func (s *CPUUsageStats) Update(usages []float64) error {
	if 0 < s.N && len(usages) != s.cpuCount {
		// Assumption: constant CPU count for all requests
		return fmt.Errorf("invalid length of usages array: expected %v, got %v", s.cpuCount, len(usages))
	}

	s.N++
	if s.N == 1 {
		s.cpuCount = len(usages)
		s.totals = make([]float64, s.cpuCount)
		s.Averages = make([]float64, s.cpuCount)
	}
	for i, usage := range usages {
		s.totals[i] += usage
		s.Averages[i] = s.totals[i] / float64(s.N)
	}
	return nil
}

// KernelMetricsHandler handles all "last_kernel_upgrade" metrics and manages KernelUpgradeStats
type KernelMetricsHandler struct {
	mu    sync.RWMutex
	stats KernelUpgradeStats
}

// KernelUpgradeStats keeps track of the most recent timestamp seen
type KernelUpgradeStats struct {
	N          int
	MostRecent time.Time
}

// Handle updates the KernelUpgradeStats with a new metric
func (h *KernelMetricsHandler) Handle(metric interface{}) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	timestamp, ok := metric.(string)
	if ok == false {
		return fmt.Errorf("failed to cast metric to string")
	}
	return h.stats.Update(timestamp)
}

// CurrentStats returns the current KernelUpgradeStats in a concurrent-safe manner
func (h KernelMetricsHandler) CurrentStats() KernelUpgradeStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.stats
}

// Update takes a new timestamp string and compares it against the current
// most recent upgrade time, replacing it if it's more recent
func (s *KernelUpgradeStats) Update(newTimestamp string) error {
	newTime, err := time.Parse(time.RFC3339, newTimestamp)
	if err != nil {
		return fmt.Errorf("unable to parse newTimestamp: %v", err)
	}

	s.N++
	if newTime.After(s.MostRecent) {
		// Assumption: "keep track of the most recent timestamp" means comparing timestamps rather
		// than just storing the timestamp that was received most recently. Even though in the case
		// of this demo, both would behave the same: https://github.com/juju/demoware/blob/master/main.go#L206
		s.MostRecent = newTime
	}
	return nil
}
