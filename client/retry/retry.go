package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

var maxWait = 10 * time.Second
var minWait = 500 * time.Millisecond
var jitterCoefficient = 0.1

func Backoff(attempts int) time.Duration {
	wait := time.Duration(float64(minWait) * math.Pow(2, float64(attempts)))
	unitDuration := int64(jitterCoefficient * float64(wait))
	jitterDuration := 2*time.Duration(rand.Int63n(unitDuration)) - time.Duration(unitDuration)
	wait += jitterDuration

	if wait > maxWait {
		wait = maxWait
	}

	return wait
}

func Poll[T any](pctx context.Context, timeout time.Duration, fn func() (*T, *Err)) (*T, error) {
	ctx, cancel := context.WithTimeout(pctx, timeout)
	defer cancel()
	var attempt int
	var lastErr error
	for {
		attempt++
		entity, err := fn()
		if err == nil {
			return entity, nil
		}
		if err.Halt {
			return nil, err.Err
		}
		lastErr = err.Err
		wait := Backoff(attempt)
		timer := time.NewTimer(wait)
		select {
		// stop when either this or parent context times out
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("timed out: %w", lastErr)
		case <-timer.C:
		}
	}
}

// Err represents a retriable error
type Err struct {
	Err  error
	Halt bool
}
