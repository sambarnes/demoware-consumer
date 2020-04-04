package metrics

import "testing"

func TestRunDispatcher(t *testing.T) {
	t.Run("CreateSubscriptionStreams", func(t *testing.T) {
		done := make(chan interface{})
		resultStream := make(chan Result)
		defer close(done)
		defer close(resultStream)

		metricSubscriptions := []MetricType{
			LoadAverageMetric,
			CPUUsageMetric,
			LastKernelUpgradeMetric,
		}
		metricStreamsByType := RunDispatcher(done, metricSubscriptions, resultStream)
		for _, metricType := range metricSubscriptions {
			if _, ok := metricStreamsByType[metricType]; !ok {
				t.Errorf("Missing %v stream", metricType)
			}
		}
	})
}
