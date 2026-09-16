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

func TestStoreGetPublicAddressPassthrough(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		wantNull bool
		want     bool
	}{
		{
			name:     "API reports disabled",
			data:     map[string]any{"clusterId": "in01-a", "status": "RUNNING", "publicAddressEnabled": false},
			wantNull: false,
			want:     false,
		},
		{
			name:     "API reports enabled",
			data:     map[string]any{"clusterId": "in01-a", "status": "RUNNING", "publicAddressEnabled": true},
			wantNull: false,
			want:     true,
		},
		{
			name:     "field absent (BYOC or older deployment) stays null",
			data:     map[string]any{"clusterId": "in01-a", "status": "RUNNING"},
			wantNull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newStoreWithDescribeBody(t, tt.data)
			state, err := store.Get(context.Background(), "in01-a")
			if err != nil {
				t.Fatalf("failed to get cluster: %v", err)
			}
			if tt.wantNull {
				if !state.PublicAddressEnabled.IsNull() {
					t.Fatalf("want null (not applicable), got = %v", state.PublicAddressEnabled)
				}
				return
			}
			if state.PublicAddressEnabled.IsNull() {
				t.Fatal("state must not be null when the API reports the field")
			}
			if got := state.PublicAddressEnabled.ValueBool(); got != tt.want {
				t.Fatalf("want = %v, got = %v", tt.want, got)
			}
		})
	}
}
