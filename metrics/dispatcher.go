package metrics

import (
	log "github.com/sirupsen/logrus"
)

// RunDispatcher routes the individual metrics of ingested Result structs to
// their payload-specific handlers and returns a map of the channels that will
// feed those handlers
func RunDispatcher(done <-chan interface{}, resultStream <-chan Result) map[MetricType]chan interface{} {
	log.Debug("Starting dispatcher...")
	metricStreamsByType := map[MetricType]chan interface{}{
		LoadAverageMetric:       make(chan interface{}),
		CPUUsageMetric:          make(chan interface{}),
		LastKernelUpgradeMetric: make(chan interface{}),
	}

	go func() {
		defer func() {
			for _, stream := range metricStreamsByType {
				close(stream)
			}
		}()

		for result := range resultStream {
			if result.Error != nil {
				log.Error(result.Error)
				continue
			}
			select {
			case <-done:
				return
			default:
				dispatch(result.Metrics, metricStreamsByType)
			}
		}
	}()
	return metricStreamsByType
}

// dispatch sends the payload of each metric in a batch to its designated handler
func dispatch(metricsBatch []Metric, metricStreamsByType map[MetricType]chan interface{}) {
	for _, metric := range metricsBatch {
		if metricStream, ok := metricStreamsByType[metric.Type]; ok {
			metricStream <- metric.Payload.Value
		} else {
			log.WithFields(log.Fields{
				"type":    metric.Type,
				"payload": metric.Payload,
			}).Warn("Unknown Metric type")
		}
	}
}
