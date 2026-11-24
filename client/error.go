package client

import (
	"fmt"
)

type Error struct {
	RequestId string `json:"requestId"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
}

func (err Error) Error() string {
	return fmt.Sprintf("Error[%d]:%s. RequestId:%s", err.Code, err.Message, err.RequestId)
}

func (err Error) Is(target error) bool {
	t, ok := target.(Error)
	if !ok {
		return false
	}
	return t.Code == err.Code
}

// HTTPError retains the HTTP status while preserving the underlying API error.
type HTTPError struct {
	StatusCode int
	RequestId  string
	Err        error
}

func (err *HTTPError) Error() string {
	return fmt.Sprintf("http status code: %d, error: %v, requestId: %s", err.StatusCode, err.Err, err.RequestId)
}

func (err *HTTPError) Unwrap() error {
	return err.Err
}
