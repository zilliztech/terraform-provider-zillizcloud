package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/zilliztech/terraform-provider-zillizcloud/client/retry"
)

const (
	requestMaxAttempts  = 4
	requestRetryTimeout = time.Minute
)

// doWithRetry is only for operations whose effects can safely be repeated.
// Callers must use doContext for creates without an idempotency key.
func (c *Client) doWithRetry(ctx context.Context, method, path string, body, result any) error {
	ctx, cancel := context.WithTimeout(ctx, requestRetryTimeout)
	defer cancel()

	attempts := 0
	_, err := retry.Poll[struct{}](ctx, requestRetryTimeout, func() (*struct{}, *retry.Err) {
		if err := ctx.Err(); err != nil {
			return nil, &retry.Err{Err: err, Halt: true}
		}
		attempts++
		err := c.doContext(ctx, method, path, body, result)
		if err == nil {
			return nil, nil
		}
		return nil, &retry.Err{
			Err:  err,
			Halt: ctx.Err() != nil || attempts >= requestMaxAttempts || !isRetryableRequestError(err),
		}
	})
	// Preserve cancellation/deadline identity even when cancellation occurred
	// during the shared poller's backoff rather than during an HTTP request.
	if err != nil && ctx.Err() != nil {
		return errors.Join(ctx.Err(), err)
	}
	return err
}

func isRetryableRequestError(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case http.StatusTooManyRequests, http.StatusInternalServerError,
			http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return true
		default:
			return false
		}
	}
	// Business errors are not transport failures, even if their message
	// contains words such as "timeout" used by the shared network classifier.
	var apiErr *Error
	var apiValue Error
	if errors.As(err, &apiErr) || errors.As(err, &apiValue) {
		return false
	}
	return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || retry.IsNetworkError(err)
}
