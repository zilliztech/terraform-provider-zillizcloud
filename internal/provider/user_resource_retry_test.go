package provider

import (
	"fmt"
	"testing"
)

func TestIsRetryableClusterError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "nil error",
			err:       nil,
			retryable: false,
		},
		{
			name:      "401 auth not ready",
			err:       fmt.Errorf("http status code: 401, error: user hasn't authenticated"),
			retryable: true,
		},
		{
			name:      "503 service unavailable",
			err:       fmt.Errorf("http status code: 503, error: service unavailable"),
			retryable: true,
		},
		{
			name:      "connection refused",
			err:       fmt.Errorf("dial tcp 10.0.0.1:19539: connection refused"),
			retryable: true,
		},
		{
			name:      "dns not ready",
			err:       fmt.Errorf("dial tcp: lookup in01-xxx.vectordb.zillizcloud.com: no such host"),
			retryable: true,
		},
		{
			name:      "400 bad request - not retryable",
			err:       fmt.Errorf("http status code: 400, error: invalid parameter"),
			retryable: false,
		},
		{
			name:      "user already exists - not retryable",
			err:       fmt.Errorf("Error[65535]:user already exists"),
			retryable: false,
		},
		{
			name:      "permission denied - not retryable",
			err:       fmt.Errorf("http status code: 403, error: permission denied"),
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryableClusterError(tt.err)
			if got != tt.retryable {
				t.Errorf("isRetryableClusterError(%v) = %v, want %v", tt.err, got, tt.retryable)
			}
		})
	}
}
