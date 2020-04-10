package metrics

import (
	log "github.com/sirupsen/logrus"
)

// MetricDispatcher routes Metric values to their payload specific handlers
type MetricDispatcher interface {
	Dispatch([]Metric)
}

// ResultStreamDispatcher is a Dispatcher based on channels
type ResultStreamDispatcher struct {
	subscriptions map[MetricType]chan interface{}
}

// Subscribe returns a new channel such that all metrics of that type will be
// sent through that channel
func (d *ResultStreamDispatcher) Subscribe(t MetricType) <-chan interface{} {
	if d.subscriptions == nil {
		d.subscriptions = make(map[MetricType]chan interface{})
	}
	d.subscriptions[t] = make(chan interface{})
	return d.subscriptions[t]
}

// Close closes the Dispatcher's Subscription channels
func (d *ResultStreamDispatcher) Close() {
	for _, stream := range d.subscriptions {
		close(stream)
	}
}

// Run pulls Results off the resultStream and dispatches each batch of Metrics
func (d ResultStreamDispatcher) Run(done <-chan interface{}, resultStream <-chan Result) {
	for {
		select {
		case <-done:
			return
		case result, ok := <-resultStream:
			if ok == false {
				return
			} else if result.Error != nil {
				// TODO: evaluate if errors should be routed to their own handler
				log.Error(result.Error)
				continue
			}
			d.Dispatch(result.Metrics)
		}
	}
}

// Dispatch sends the payload of each metric in a batch to its designated handler
func (d ResultStreamDispatcher) Dispatch(metricsBatch []Metric) {
	for _, metric := range metricsBatch {
		metricStream, ok := d.subscriptions[metric.Type]
		if ok == false {
			continue
		}
		metricStream <- metric.Payload.Value
	}
}
