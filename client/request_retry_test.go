package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRequestRetryClassification(t *testing.T) {
	for _, status := range []int{400, 401, 403, 404, 409, 422, 429, 500, 501, 502, 503, 504} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			err := &HTTPError{StatusCode: status, Err: Error{Code: 123, Message: "timeout"}}
			want := status == 429 || status == 500 || status == 502 || status == 503 || status == 504
			require.Equal(t, want, isRetryableRequestError(err))
		})
	}
	for _, err := range []error{nil, context.Canceled, context.DeadlineExceeded,
		Error{Code: 403, Message: "timeout"}, &Error{Code: 403, Message: "timeout"}, errors.New("invalid argument")} {
		require.False(t, isRetryableRequestError(err), "%v", err)
	}
	require.True(t, isRetryableRequestError(io.EOF))
	require.True(t, isRetryableRequestError(io.ErrUnexpectedEOF))
}

func TestServiceAccountRetryableOperations(t *testing.T) {
	t.Parallel()
	status := "disable"
	description := ""
	empty := []string{}
	assignments := []ServiceAccountProjectRoles{}
	cases := []struct {
		name   string
		call   func(context.Context, *Client) error
		result string
	}{
		{"get account", func(ctx context.Context, c *Client) error { _, err := c.GetServiceAccount(ctx, "sa-test"); return err }, `{}`},
		{"list accounts", func(ctx context.Context, c *Client) error {
			_, err := c.ListServiceAccounts(ctx, ServiceAccountListOptions{})
			return err
		}, `{}`},
		{"update status", func(ctx context.Context, c *Client) error {
			_, err := c.UpdateServiceAccount(ctx, "sa-test", &UpdateServiceAccountRequest{Status: &status})
			return err
		}, `{}`},
		{"update metadata", func(ctx context.Context, c *Client) error {
			_, err := c.UpdateServiceAccount(ctx, "sa-test", &UpdateServiceAccountRequest{Description: &description})
			return err
		}, `{}`},
		{"assign", func(ctx context.Context, c *Client) error {
			_, err := c.AssignServiceAccount(ctx, "sa-test", "proj-test", &AssignServiceAccountRequest{ProjectRoleIds: empty})
			return err
		}, `{}`},
		{"unassign", func(ctx context.Context, c *Client) error {
			return c.UnassignServiceAccount(ctx, "sa-test", "proj-test")
		}, `true`},
		{"replace access", func(ctx context.Context, c *Client) error {
			_, err := c.ReplaceServiceAccountAccess(ctx, "sa-test", &ReplaceServiceAccountAccessRequest{ProjectAssignments: &assignments})
			return err
		}, `{}`},
		{"update project roles", func(ctx context.Context, c *Client) error {
			_, err := c.UpdateServiceAccountProjectRoles(ctx, "sa-test", "proj-test", &UpdateServiceAccountProjectRolesRequest{ProjectRoleIds: empty})
			return err
		}, `{}`},
		{"delete account", func(ctx context.Context, c *Client) error { return c.DeleteServiceAccount(ctx, "sa-test") }, `true`},
		{"get key", func(ctx context.Context, c *Client) error {
			_, err := c.GetServiceAccountKey(ctx, "sa-test", "key-test")
			return err
		}, `{}`},
		{"list keys", func(ctx context.Context, c *Client) error {
			_, err := c.ListServiceAccountKeys(ctx, "sa-test", ServiceAccountKeyListOptions{})
			return err
		}, `{}`},
		{"update key", func(ctx context.Context, c *Client) error {
			_, err := c.UpdateServiceAccountKey(ctx, "sa-test", "key-test", &UpdateServiceAccountKeyRequest{Description: &description})
			return err
		}, `{}`},
		{"delete key", func(ctx context.Context, c *Client) error {
			return c.DeleteServiceAccountKey(ctx, "sa-test", "key-test")
		}, `true`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			calls := 0
			var firstBody string
			c, err := NewClient(WithApiKey("test"), WithHTTPClient(&mockHTTPClient{do: func(req *http.Request) (*http.Response, error) {
				calls++
				body := ""
				if req.Body != nil {
					raw, err := io.ReadAll(req.Body)
					require.NoError(t, err)
					body = string(raw)
				}
				if calls == 1 {
					firstBody = body
				} else {
					require.Equal(t, firstBody, body, "request body must be replayed intact")
				}
				deadline, ok := req.Context().Deadline()
				require.True(t, ok)
				require.LessOrEqual(t, time.Until(deadline), time.Minute)
				if calls == 1 {
					return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("upstream unavailable"))}, nil
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"code":0,"data":` + tc.result + `}`))}, nil
			}}))
			require.NoError(t, err)
			require.NoError(t, tc.call(context.Background(), c))
			require.Equal(t, 2, calls)
		})
	}
}

func TestServiceAccountDoesNotRetryUnsafeWrites(t *testing.T) {
	name, status := "new-name", "disable"
	calls := 0
	c, err := NewClient(WithApiKey("test"), WithHTTPClient(&mockHTTPClient{do: func(req *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader(`{"code":500,"message":"unavailable"}`))}, nil
	}}))
	require.NoError(t, err)
	_, err = c.CreateServiceAccount(context.Background(), &CreateServiceAccountRequest{Name: "test"})
	require.Error(t, err)
	_, err = c.CreateProjectServiceAccount(context.Background(), "proj-test", &CreateProjectServiceAccountRequest{Name: "test"})
	require.Error(t, err)
	_, err = c.CreateServiceAccountKey(context.Background(), "sa-test", &CreateServiceAccountKeyRequest{Name: "test", ExpireDays: 1})
	require.Error(t, err)
	_, err = c.UpdateServiceAccount(context.Background(), "sa-test", &UpdateServiceAccountRequest{Name: &name, Status: &status})
	require.Error(t, err)
	require.Equal(t, 4, calls)
}

func TestRequestRetryExhaustionPreservesError(t *testing.T) {
	t.Parallel()
	calls := 0
	c, err := NewClient(WithApiKey("test"), WithHTTPClient(&mockHTTPClient{do: func(req *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 500, Header: http.Header{"Requestid": []string{"req-test"}}, Body: io.NopCloser(strings.NewReader(`{"code":42,"message":"sync failed"}`))}, nil
	}}))
	require.NoError(t, err)
	_, err = c.GetServiceAccount(context.Background(), "sa-test")
	require.Error(t, err)
	require.Equal(t, 4, calls)
	var httpErr *HTTPError
	require.ErrorAs(t, err, &httpErr)
	require.Equal(t, 500, httpErr.StatusCode)
	require.Equal(t, "req-test", httpErr.RequestId)
	var apiErr Error
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, 42, apiErr.Code)
	require.ErrorIs(t, err, Error{Code: 42})
}

func TestRequestRetryCancellation(t *testing.T) {
	for _, inFlight := range []bool{false, true} {
		t.Run(fmt.Sprintf("in-flight=%t", inFlight), func(t *testing.T) {
			calls := 0
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			started := make(chan struct{})
			c, err := NewClient(WithApiKey("test"), WithHTTPClient(&mockHTTPClient{do: func(req *http.Request) (*http.Response, error) {
				calls++
				close(started)
				if inFlight {
					<-req.Context().Done()
					return nil, req.Context().Err()
				}
				return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("unavailable"))}, nil
			}}))
			require.NoError(t, err)
			done := make(chan error, 1)
			go func() { _, err := c.GetServiceAccount(ctx, "sa-test"); done <- err }()
			<-started
			// Cancel after the first response has entered its backoff.
			if !inFlight {
				time.Sleep(20 * time.Millisecond)
			}
			cancel()
			select {
			case err := <-done:
				require.ErrorIs(t, err, context.Canceled)
			case <-time.After(time.Second):
				t.Fatal("request did not stop after cancellation")
			}
			require.Equal(t, 1, calls)
		})
	}
}

func TestRequestRetryDeadlineAndNetworkFailure(t *testing.T) {
	t.Run("deadline", func(t *testing.T) {
		calls := 0
		c, err := NewClient(WithApiKey("test"), WithHTTPClient(&mockHTTPClient{do: func(req *http.Request) (*http.Response, error) {
			calls++
			return nil, io.ErrUnexpectedEOF
		}}))
		require.NoError(t, err)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		_, err = c.GetServiceAccount(ctx, "sa-test")
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Equal(t, 1, calls)
	})
	t.Run("network recovers", func(t *testing.T) {
		t.Parallel()
		calls := 0
		c, err := NewClient(WithApiKey("test"), WithHTTPClient(&mockHTTPClient{do: func(req *http.Request) (*http.Response, error) {
			calls++
			if calls == 1 {
				return nil, errors.New("connection reset by peer")
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"code":0,"data":{}}`))}, nil
		}}))
		require.NoError(t, err)
		_, err = c.GetServiceAccount(context.Background(), "sa-test")
		require.NoError(t, err)
		require.Equal(t, 2, calls)
	})
}

func TestServiceAccountPermanentErrorsAreNotRetried(t *testing.T) {
	for _, status := range []int{200, 400, 401, 403, 404, 409, 422} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			calls := 0
			c, err := NewClient(WithApiKey("test"), WithHTTPClient(&mockHTTPClient{do: func(req *http.Request) (*http.Response, error) {
				calls++
				return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(`{"code":1100,"message":"invalid timeout parameter"}`))}, nil
			}}))
			require.NoError(t, err)
			_, err = c.GetServiceAccount(context.Background(), "sa-test")
			require.Error(t, err)
			require.Equal(t, 1, calls)
		})
	}
}
