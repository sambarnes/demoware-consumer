package metrics

import (
	"testing"
)

func TestResultStreamDispatcher_Dispatch(t *testing.T) {
	dispatcher := &ResultStreamDispatcher{}
	subscriptionStream := dispatcher.Subscribe(LoadAverageMetric)

	good := Metric{LoadAverageMetric, MetricPayload{}}
	bad := Metric{CPUUsageMetric, MetricPayload{}}
	batch := []Metric{
		good,
		good,
		bad,
		bad,
		bad,
		good,
		good,
	}
	go func() {
		dispatcher.Dispatch(batch)
		dispatcher.Close()
	}()

	metricsExpected := 4
	metricsObserved := 0
	for range subscriptionStream {
		metricsObserved++
	}
	if metricsObserved != metricsExpected {
		t.Errorf("unexpected metrics observed: %v != %v (observed, expected)", metricsObserved, metricsExpected)
	}
}
