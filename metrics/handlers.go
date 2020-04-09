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
	Stats LoadStats
}

// LoadStats keeps track of the min and max load seen
type LoadStats struct {
	n   int
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
	return h.Stats.Update(load)
}

// CurrentStats returns the current LoadStats in a concurrent-safe manner
func (h LoadMetricsHandler) CurrentStats() LoadStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.Stats
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

// CPUMetricsHandler handles all "cpu_usage" metrics and manages CPUUsageStats
type CPUMetricsHandler struct {
	mu    sync.RWMutex
	Stats CPUUsageStats
}

// CPUUsageStats keeps track of the running average CPU usage per core
type CPUUsageStats struct {
	n        int
	cpuCount int
	totals   []float64
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
	return h.Stats.Update(usages)
}

// CurrentStats returns the current CPUUsageStats in a concurrent-safe manner
func (h CPUMetricsHandler) CurrentStats() CPUUsageStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.Stats
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

// KernelMetricsHandler handles all "last_kernel_upgrade" metrics and manages KernelUpgradeStats
type KernelMetricsHandler struct {
	mu    sync.RWMutex
	Stats KernelUpgradeStats
}

// KernelUpgradeStats keeps track of the most recent timestamp seen
type KernelUpgradeStats struct {
	n          int
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
	return h.Stats.Update(timestamp)
}

// CurrentStats returns the current KernelUpgradeStats in a concurrent-safe manner
func (h KernelMetricsHandler) CurrentStats() KernelUpgradeStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.Stats
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
