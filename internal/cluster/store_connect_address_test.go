package cluster

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func newStoreWithDescribeBody(t *testing.T, data map[string]any) ClusterStore {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": data}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	c, err := zilliz.NewClient(
		zilliz.WithBaseUrl(server.URL),
		zilliz.WithApiKey("gibberish_key"),
		zilliz.WithCloudRegionId("gcp-us-west1"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return &ClusterStoreImpl{client: c}
}

func TestStoreGetConnectAddressFallback(t *testing.T) {
	tests := []struct {
		name string
		data map[string]any
		want bool
	}{
		{
			name: "API reports disabled",
			data: map[string]any{"clusterId": "in01-a", "status": "RUNNING", "connectAddressEnabled": false},
			want: false,
		},
		{
			name: "API reports enabled",
			data: map[string]any{"clusterId": "in01-a", "status": "RUNNING", "connectAddressEnabled": true},
			want: true,
		},
		{
			name: "field absent (BYOC or older deployment) falls back to true",
			data: map[string]any{"clusterId": "in01-a", "status": "RUNNING"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newStoreWithDescribeBody(t, tt.data)
			state, err := store.Get(context.Background(), "in01-a")
			if err != nil {
				t.Fatalf("failed to get cluster: %v", err)
			}
			if state.ConnectAddressEnabled.IsNull() {
				t.Fatal("state must never be null")
			}
			if got := state.ConnectAddressEnabled.ValueBool(); got != tt.want {
				t.Fatalf("want = %v, got = %v", tt.want, got)
			}
		})
	}
}
