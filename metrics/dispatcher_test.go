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
	t.Run("DispatchesCorrectly", func(t *testing.T) {
		done := make(chan interface{})
		defer close(done)

		metricSubscriptions := []MetricType{LoadAverageMetric}
		resultStream := make(chan Result)
		metricStreamsByType := RunDispatcher(done, metricSubscriptions, resultStream)
		good := Metric{LoadAverageMetric, MetricPayload{}}
		bad := Metric{CPUUsageMetric, MetricPayload{}}
		resultStream <- Result{
			Metrics: []Metric{
				good,
				good,
				bad,
				bad,
				bad,
				good,
				good,
			},
		}
		close(resultStream)

		metricsExpected := 4
		metricsObserved := 0
	loop:
		for {
			select {
			case <-done:
				break loop
			case _, ok := <-metricStreamsByType[LoadAverageMetric]:
				if ok == false {
					break loop
				}
				metricsObserved++
			}
		}
		if metricsObserved != metricsExpected {
			t.Errorf("unexpected metrics observed: %v != %v (observed, expected)", metricsObserved, metricsExpected)
		}
	})
}
