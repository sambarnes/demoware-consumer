package metrics

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// TODO: better configuration mangagement for remote API url
var DemowareMetricsURL = "http://localhost:8080/metrics"

type Result struct {
	Error   error
	Metrics []Metric
}

type Metric struct {
	Type    MetricType    `json:"type"`
	Payload MetricPayload `json:"payload"`
}

type MetricType string

type MetricPayload struct {
	Value interface{} `json:"value"`
}

const (
	LoadAverageMetric       MetricType = "load_avg"
	CPUUsageMetric          MetricType = "cpu_usage"
	LastKernelUpgradeMetric MetricType = "last_kernel_upgrade"
)

func runGenerator(done <-chan interface{}) <-chan interface{} {
	log.WithFields(log.Fields{
		"demowareMetricsURL": DemowareMetricsURL,
	}).Debug("Starting metric ingestion...")
	return repeatFn(done, ingest)
}

// RunGenerator repeatedly calls the metrics API and returns a channel that
// streams responses as Result structs
func RunGenerator(done <-chan interface{}) <-chan Result {
	return toResult(done, runGenerator(done))
}

// RunGeneratorN calls the metrics API n times and returns a channel that
// streams those responses as Result structs
func RunGeneratorN(done <-chan interface{}, n int) <-chan Result {
	return toResult(done, take(done, runGenerator(done), n))
}

// ingest makes a call to the remote demoware API returns the Result
func ingest() interface{} {
	resp, err := http.Get(DemowareMetricsURL)
	if err != nil {
		return Result{
			Error:   err,
			Metrics: nil,
		}
	} else if 400 <= resp.StatusCode && resp.StatusCode <= 599 {
		return Result{
			Error:   fmt.Errorf("unsucessful request: %v", resp.Status),
			Metrics: nil,
		}
	}
	// TODO: exponential backoff on certain types of errors
	defer resp.Body.Close()

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Result{
			Error:   err,
			Metrics: nil,
		}
	}
	metrics, err := unmarshalMetricsBatch(responseData)
	if err != nil {
		return Result{
			Error:   err,
			Metrics: nil,
		}
	}
	return Result{
		Error:   nil,
		Metrics: metrics,
	}
}

func unmarshalMetricsBatch(data []byte) ([]Metric, error) {
	metrics := make([]Metric, 0)
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}
