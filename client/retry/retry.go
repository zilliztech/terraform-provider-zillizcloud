package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var maxWait = 10 * time.Second
var minWait = 500 * time.Millisecond
var jitterCoefficient = 0.1

func Backoff(attempts int) time.Duration {
	wait := time.Duration(float64(minWait) * math.Pow(2, float64(attempts)))

	// Apply maxWait limit before jitter calculation to prevent overflow
	if wait > maxWait {
		wait = maxWait
	}

	jitterValue := jitterCoefficient * float64(wait)
	var jitterDuration time.Duration

	// Ensure jitterValue is positive and within int64 bounds
	if jitterValue > 0 && jitterValue < float64(math.MaxInt64) {
		unitDuration := int64(jitterValue)
		jitterDuration = 2*time.Duration(rand.Int63n(unitDuration)) - time.Duration(unitDuration)
	}

	wait += jitterDuration

	return wait
}

func Poll[T any](pctx context.Context, timeout time.Duration, fn func() (*T, *Err)) (*T, error) {
	return PollWithNetworkResilience(pctx, timeout, fn, DefaultMaxNetworkFailures)
}

// PollWithNetworkResilience allows configuring the maximum number of network failures before giving up
func PollWithNetworkResilience[T any](pctx context.Context, timeout time.Duration, fn func() (*T, *Err), maxNetworkFailures int) (*T, error) {
	ctx, cancel := context.WithTimeout(pctx, timeout)
	defer cancel()
	var attempt int
	var networkFailureCount int
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

		// Track network failures
		if IsNetworkError(err.Err) {
			networkFailureCount++
			tflog.Info(pctx, fmt.Sprintf("Network failure count: %d", networkFailureCount))
			// If we've hit the network failure limit, return give-up error
			if networkFailureCount > maxNetworkFailures {
				tflog.Info(pctx, fmt.Sprintf("Network failure limit reached: %d", networkFailureCount))
				giveUpErr := &NetworkGiveUpError{
					Attempts: networkFailureCount,
					Message:  fmt.Sprintf("network errors exceeded limit of %d", maxNetworkFailures),
				}
				return nil, giveUpErr
			}
		}

		// Use network-specific backoff for network errors
		var wait time.Duration
		if IsNetworkError(err.Err) {
			wait = NetworkBackoff(attempt)
		} else {
			wait = Backoff(attempt)
		}
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

type Err struct {
	Err  error
	Halt bool
}

// NetworkResilientPoll is a variant of Poll that allows network failures up to a specified limit
// before giving up gracefully. This is useful for operations that can tolerate some network
// instability but should eventually succeed or fail gracefully.
func NetworkResilientPoll[T any](pctx context.Context, timeout time.Duration, fn func() (*T, *Err), maxNetworkFailures int) (*T, error) {
	return PollWithNetworkResilience(pctx, timeout, fn, maxNetworkFailures)
}
