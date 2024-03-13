package boilergo

import (
    "time"
    "fmt"
)

func WithRetry(fn func() (interface{}, error), condition func(interface{}) bool, maxRetries int, retryInterval time.Duration) (interface{}, error) {
	var (
		res     interface{}
		err     error
		retry   = make(chan bool, 1)
		retries = 0
	)

	go func() {
		retry <- true // Start with a retry signal to kick off the process
	}()

	for shouldRetry := range retry {
		if !shouldRetry || retries >= maxRetries {
			close(retry) // Important to close the channel to break the loop
			break
		}

		res, err = fn()
		if err != nil {
			close(retry)    // Make sure to close the channel on error too
			return nil, err // Return early if condition results in an error
		}

		if condition(res) {
			close(retry)    // Condition met, close the channel
			return res, nil // and return successfully
		}

		retries++
		if retries < maxRetries {
			time.Sleep(retryInterval) // Wait before retrying
			retry <- true             // Signal to retry
		} else {
			retry <- false // Signal to stop
		}
	}

	if retries >= maxRetries {
		return nil, fmt.Errorf("condition not met after %d attempts", maxRetries)
	}

	return res, nil
}
