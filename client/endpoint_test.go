package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

type mockHTTPClient struct{ do func(*http.Request) (*http.Response, error) }

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) { return m.do(req) }

func newMockClient(t *testing.T, handler func(*http.Request) (*http.Response, error)) *Client {
	t.Helper()
	c, err := NewClient(
		WithApiKey("test-key"),
		WithBaseUrl("https://api.test/v2"),
		WithHTTPClient(&mockHTTPClient{do: handler}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func jsonResponse(t *testing.T, status int, body any) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader(b)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func TestUnitListEndpointServices(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "GET" {
			t.Errorf("method=%s", req.Method)
		}
		if !strings.Contains(req.URL.String(), "/endpointServices") {
			t.Errorf("url=%s", req.URL.String())
		}
		if req.URL.Query().Get("regionId") != "aws-us-west-2" {
			t.Errorf("regionId=%s", req.URL.Query().Get("regionId"))
		}
		return jsonResponse(t, 200, map[string]any{
			"code": 0,
			"data": map[string]any{
				"endpointServices": []map[string]any{
					{"regionId": "aws-us-west-2", "cloudId": "aws", "endpointService": "svc-x", "whitelistRequired": false},
				},
				"currentPage": 1, "pageSize": 10, "count": 1,
			},
		}), nil
	})

	svcs, page, err := c.ListEndpointServices("aws-us-west-2", 1, 10)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(svcs) != 1 || svcs[0].EndpointService != "svc-x" {
		t.Errorf("svcs=%+v", svcs)
	}
	if page.Count != 1 {
		t.Errorf("count=%d", page.Count)
	}
}

func TestUnitListEndpoints(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if !strings.Contains(req.URL.Path, "/projects/proj-1/endpoints") {
			t.Errorf("path=%s", req.URL.Path)
		}
		return jsonResponse(t, 200, map[string]any{
			"code": 0,
			"data": map[string]any{
				"endpoints": []map[string]any{
					{"regionId": "aws-us-west-2", "cloudId": "aws", "endpointService": "svc-x",
						"endpointServiceStatus": "Available", "endpointId": "vpce-abc",
						"endpointStatus": "accepted", "gcpProjectId": nil},
				},
				"currentPage": 1, "pageSize": 10, "count": 1,
			},
		}), nil
	})

	eps, _, err := c.ListEndpoints("proj-1", 1, 10)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(eps) != 1 || eps[0].EndpointId != "vpce-abc" {
		t.Errorf("eps=%+v", eps)
	}
	if eps[0].GcpProjectId != nil {
		t.Errorf("expected gcpProjectId nil, got %v", eps[0].GcpProjectId)
	}
}

func TestUnitCreateEndpoint(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "POST" {
			t.Errorf("method=%s", req.Method)
		}
		var body CreateEndpointRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.EndpointId != "vpce-abc" || body.RegionId != "aws-us-west-2" {
			t.Errorf("body=%+v", body)
		}
		return jsonResponse(t, 200, map[string]any{
			"code": 0,
			"data": map[string]any{"endpointId": "vpce-abc", "regionId": "aws-us-west-2"},
		}), nil
	})

	resp, err := c.CreateEndpoint("proj-1", &CreateEndpointRequest{
		RegionId: "aws-us-west-2", EndpointId: "vpce-abc",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp.EndpointId != "vpce-abc" {
		t.Errorf("resp=%+v", resp)
	}
}

func TestUnitDeleteEndpoint(t *testing.T) {
	t.Run("no gcpProjectId", func(t *testing.T) {
		c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
			if req.Method != "DELETE" {
				t.Errorf("method=%s", req.Method)
			}
			if req.URL.Query().Get("regionId") != "aws-us-west-2" {
				t.Errorf("regionId missing")
			}
			if _, ok := req.URL.Query()["gcpProjectId"]; ok {
				t.Errorf("gcpProjectId should not be present")
			}
			return jsonResponse(t, 200, map[string]any{
				"code": 0, "data": map[string]any{"endpointId": "vpce-abc"},
			}), nil
		})
		if err := c.DeleteEndpoint("proj-1", "vpce-abc", "aws-us-west-2", nil); err != nil {
			t.Fatalf("err: %v", err)
		}
	})

	t.Run("with gcpProjectId", func(t *testing.T) {
		gcp := "my-gcp-proj"
		c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
			if req.URL.Query().Get("gcpProjectId") != "my-gcp-proj" {
				t.Errorf("gcpProjectId=%s", req.URL.Query().Get("gcpProjectId"))
			}
			return jsonResponse(t, 200, map[string]any{
				"code": 0, "data": map[string]any{"endpointId": "vpce-abc"},
			}), nil
		})
		if err := c.DeleteEndpoint("proj-1", "vpce-abc", "gcp-us-west1", &gcp); err != nil {
			t.Fatalf("err: %v", err)
		}
	})
}

func TestUnitAddEndpointWhitelist(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "POST" {
			t.Errorf("method=%s", req.Method)
		}
		if !strings.HasSuffix(req.URL.Path, "/projects/proj-1/endpointWhitelist") {
			t.Errorf("path=%s", req.URL.Path)
		}
		var body AddEndpointWhitelistRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.OuterUserId != "user-abc" {
			t.Errorf("body=%+v", body)
		}
		return jsonResponse(t, 200, map[string]any{"code": 0, "data": "success"}), nil
	})

	err := c.AddEndpointWhitelist("proj-1", &AddEndpointWhitelistRequest{
		RegionId: "azure-eastus2", OuterUserId: "user-abc",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
}
