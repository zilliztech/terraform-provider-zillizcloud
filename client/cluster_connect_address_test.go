package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func writeConnectAddressResponse(t *testing.T, w http.ResponseWriter, data any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"code": 0,
		"data": data,
	}); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}

func TestClient_ConnectAddress(t *testing.T) {
	enabled := true
	type request struct{ method, path string }
	var requests []request

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, request{r.Method, r.URL.Path})
		clusterId := "in01-abc123"
		switch r.URL.Path {
		case "/clusters/" + clusterId + "/enableConnectAddress":
			enabled = true
			writeConnectAddressResponse(t, w, map[string]any{"clusterId": clusterId})
		case "/clusters/" + clusterId + "/disableConnectAddress":
			enabled = false
			writeConnectAddressResponse(t, w, map[string]any{"clusterId": clusterId})
		case "/clusters/" + clusterId:
			writeConnectAddressResponse(t, w, map[string]any{
				"clusterId":             clusterId,
				"clusterName":           "test",
				"status":                "RUNNING",
				"connectAddressEnabled": enabled,
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

	got, err := c.DisableConnectAddress(clusterId)
	if err != nil {
		t.Fatalf("failed to disable public endpoint: %v", err)
	}
	if got == nil || *got != clusterId {
		t.Fatalf("want = %s, got = %v", clusterId, got)
	}

	cluster, err := c.DescribeCluster(clusterId)
	if err != nil {
		t.Fatalf("failed to describe cluster: %v", err)
	}
	if cluster.ConnectAddressEnabled == nil || *cluster.ConnectAddressEnabled {
		t.Fatalf("want connectAddressEnabled = false, got = %v", cluster.ConnectAddressEnabled)
	}

	got, err = c.EnableConnectAddress(clusterId)
	if err != nil {
		t.Fatalf("failed to enable public endpoint: %v", err)
	}
	if got == nil || *got != clusterId {
		t.Fatalf("want = %s, got = %v", clusterId, got)
	}

	cluster, err = c.DescribeCluster(clusterId)
	if err != nil {
		t.Fatalf("failed to describe cluster: %v", err)
	}
	if cluster.ConnectAddressEnabled == nil || !*cluster.ConnectAddressEnabled {
		t.Fatalf("want connectAddressEnabled = true, got = %v", cluster.ConnectAddressEnabled)
	}

	wantRequests := []request{
		{"POST", "/clusters/in01-abc123/disableConnectAddress"},
		{"GET", "/clusters/in01-abc123"},
		{"POST", "/clusters/in01-abc123/enableConnectAddress"},
		{"GET", "/clusters/in01-abc123"},
	}
	if len(requests) != len(wantRequests) {
		t.Fatalf("want %d requests, got %d: %v", len(wantRequests), len(requests), requests)
	}
	for i, want := range wantRequests {
		if requests[i] != want {
			t.Fatalf("request %d: want %v, got %v", i, want, requests[i])
		}
	}
}

// Older deployments do not return connectAddressEnabled; the client must keep it nil.
func TestClient_ConnectAddressEnabledAbsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeConnectAddressResponse(t, w, map[string]any{
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
	if cluster.ConnectAddressEnabled != nil {
		t.Fatalf("want connectAddressEnabled = nil, got = %v", *cluster.ConnectAddressEnabled)
	}
}
