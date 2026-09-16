package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func writePublicAddressResponse(t *testing.T, w http.ResponseWriter, data any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"code": 0,
		"data": data,
	}); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}

func TestClient_UpdatePublicAddressEnabled(t *testing.T) {
	enabled := true
	type request struct{ method, path, body string }
	var requests []request

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests = append(requests, request{r.Method, r.URL.Path, string(body)})
		clusterId := "in01-abc123"
		switch {
		case r.URL.Path == "/clusters/"+clusterId && r.Method == http.MethodPatch:
			var params UpdatePublicAddressEnabledParams
			if err := json.Unmarshal(body, &params); err != nil {
				t.Fatalf("decode patch body: %v", err)
			}
			enabled = params.PublicAddressEnabled != nil && *params.PublicAddressEnabled
			writePublicAddressResponse(t, w, map[string]any{"clusterId": clusterId})
		case r.URL.Path == "/clusters/"+clusterId && r.Method == http.MethodGet:
			writePublicAddressResponse(t, w, map[string]any{
				"clusterId":            clusterId,
				"clusterName":          "test",
				"status":               "RUNNING",
				"publicAddressEnabled": enabled,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c, err := NewClient(
		WithBaseUrl(server.URL),
		WithApiKey("gibberish_key"),
		WithCloudRegionId("gcp-us-west1"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	clusterId := "in01-abc123"

	got, err := c.UpdatePublicAddressEnabled(clusterId, false)
	if err != nil {
		t.Fatalf("failed to disable public address: %v", err)
	}
	if got == nil || *got != clusterId {
		t.Fatalf("want = %s, got = %v", clusterId, got)
	}

	cluster, err := c.DescribeCluster(clusterId)
	if err != nil {
		t.Fatalf("failed to describe cluster: %v", err)
	}
	if cluster.PublicAddressEnabled == nil || *cluster.PublicAddressEnabled {
		t.Fatalf("want publicAddressEnabled = false, got = %v", cluster.PublicAddressEnabled)
	}

	got, err = c.UpdatePublicAddressEnabled(clusterId, true)
	if err != nil {
		t.Fatalf("failed to enable public address: %v", err)
	}
	if got == nil || *got != clusterId {
		t.Fatalf("want = %s, got = %v", clusterId, got)
	}

	cluster, err = c.DescribeCluster(clusterId)
	if err != nil {
		t.Fatalf("failed to describe cluster: %v", err)
	}
	if cluster.PublicAddressEnabled == nil || !*cluster.PublicAddressEnabled {
		t.Fatalf("want publicAddressEnabled = true, got = %v", cluster.PublicAddressEnabled)
	}

	wantRequests := []request{
		{http.MethodPatch, "/clusters/in01-abc123", `{"publicAddressEnabled":false}`},
		{http.MethodGet, "/clusters/in01-abc123", ""},
		{http.MethodPatch, "/clusters/in01-abc123", `{"publicAddressEnabled":true}`},
		{http.MethodGet, "/clusters/in01-abc123", ""},
	}
	if len(requests) != len(wantRequests) {
		t.Fatalf("want %d requests, got %d: %v", len(wantRequests), len(requests), requests)
	}
	for i, want := range wantRequests {
		got := requests[i]
		if got.method != want.method || got.path != want.path ||
			strings.TrimSpace(got.body) != strings.TrimSpace(want.body) {
			t.Fatalf("request %d: want %v, got %v", i, want, got)
		}
	}
}

// BYOC clusters (and older deployments) omit publicAddressEnabled; the client
// must keep it nil so the provider can treat the toggle as not applicable.
func TestClient_PublicAddressEnabledAbsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writePublicAddressResponse(t, w, map[string]any{
			"clusterId":   "in01-abc123",
			"clusterName": "test",
			"status":      "RUNNING",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c, err := NewClient(
		WithBaseUrl(server.URL),
		WithApiKey("gibberish_key"),
		WithCloudRegionId("gcp-us-west1"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	cluster, err := c.DescribeCluster("in01-abc123")
	if err != nil {
		t.Fatalf("failed to describe cluster: %v", err)
	}
	if cluster.PublicAddressEnabled != nil {
		t.Fatalf("want publicAddressEnabled = nil, got = %v", *cluster.PublicAddressEnabled)
	}
}
