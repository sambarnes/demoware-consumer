package metrics

// repeatFn is a helper metrics function that calls the function passed to it
// indefinitely until told to stop
func repeatFn(done <-chan interface{}, fn func() interface{}) <-chan interface{} {
	valueStream := make(chan interface{})
	go func() {
		defer close(valueStream)
		for {
			select {
			case <-done:
				return
			case valueStream <- fn():
			}
		}
	}()
	return valueStream
}

// take is a helper metrics stage that takes the first num items off the
// incoming channel and then exits for demo and test purposes
func take(done <-chan interface{}, valueStream <-chan interface{}, num int) <-chan interface{} {
	takeStream := make(chan interface{})
	go func() {
		defer close(takeStream)
		for i := 0; i < num; i++ {
			select {
			case <-done:
				return
			case takeStream <- <-valueStream:
			}
		}
	}()
	return takeStream
}

// toResult wraps the given <-chan interface{} as a <-chan Result
func toResult(done <-chan interface{}, valueStream <-chan interface{}) <-chan Result {
	wrappedStream := make(chan Result)
	go func() {
		defer close(wrappedStream)
		for v := range valueStream {
			select {
			case <-done:
				return
			case wrappedStream <- v.(Result):
			}
		}
	}()
	return wrappedStream
}

// toFloat64Array converts the given []interface{} to []float64
func toFloat64Array(arr []interface{}) []float64 {
	result := make([]float64, len(arr))
	for i, x := range arr {
		result[i] = x.(float64)
	}
	return result
}
