package metrics

import "fmt"

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

// orDone wraps the given channel handling the done condition for the caller,
// enhancing for loop readability when ranging over a channel
func orDone(done, c <-chan interface{}) <-chan interface{} {
	valStream := make(chan interface{})
	go func() {
		defer close(valStream)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if ok == false {
					return
				}
				select {
				case valStream <- v:
				case <-done:
				}
			}
		}
	}()
	return valStream
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
func toFloat64Array(arr []interface{}) ([]float64, error) {
	result := make([]float64, len(arr))
	for i, x := range arr {
		f, ok := x.(float64)
		if ok == false {
			return nil, fmt.Errorf("failed to cast element to float64")
		}
		result[i] = f
	}
	return result, nil
}
