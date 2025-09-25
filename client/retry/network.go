package retry

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/url"
	"strings"
	"time"
)

// DefaultMaxNetworkFailures is the default maximum number of network failures before giving up
const DefaultMaxNetworkFailures = 3

// Network error patterns that suggest temporary network issues
var networkErrorPatterns = []string{
	"connection refused",
	"connection reset",
	"connection timeout",
	"network unreachable",
	"temporary failure",
	"timeout",
	"no such host",
	"i/o timeout",
	"tls handshake timeout",
	"context deadline exceeded",
}

// IsNetworkError checks if the error appears to be a network-related issue
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// Check for specific network error types
	if _, ok := err.(net.Error); ok {
		return true
	}

	if _, ok := err.(*url.Error); ok {
		return true
	}

	// Check error message patterns
	for _, pattern := range networkErrorPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// NetworkBackoff provides more generous backoff for network errors
func NetworkBackoff(attempts int) time.Duration {
	// For network errors, use a longer minimum wait and higher max
	networkMinWait := 2 * time.Second
	networkMaxWait := 120 * time.Second // 2 minutes max for network issues

	wait := time.Duration(float64(networkMinWait) * math.Pow(2, float64(attempts)))

	if wait > networkMaxWait {
		wait = networkMaxWait
	}

	// Apply jitter
	jitterValue := jitterCoefficient * float64(wait)
	var jitterDuration time.Duration

	if jitterValue > 0 && jitterValue < float64(math.MaxInt64) {
		unitDuration := int64(jitterValue)
		jitterDuration = 2*time.Duration(rand.Int63n(unitDuration)) - time.Duration(unitDuration)
	}

	wait += jitterDuration

	return wait
}

// Err represents a retriable error.

// NetworkGiveUpError indicates that network retries have been exhausted
type NetworkGiveUpError struct {
	Attempts int
	Message  string
}

func (e *NetworkGiveUpError) Error() string {
	return fmt.Sprintf("network retries exhausted after %d attempts: %s", e.Attempts, e.Message)
}

// IsNetworkGiveUpError checks if an error is a NetworkGiveUpError
func IsNetworkGiveUpError(err error) bool {
	_, ok := err.(*NetworkGiveUpError)
	return ok
}
