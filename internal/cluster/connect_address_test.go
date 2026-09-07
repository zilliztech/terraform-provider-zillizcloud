package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func TestIsConnectAddressEnabledChanged(t *testing.T) {
	plan := ClusterResourceModel{ConnectAddressEnabled: types.BoolValue(false)}
	state := ClusterResourceModel{ConnectAddressEnabled: types.BoolValue(true)}
	if !plan.isConnectAddressEnabledChanged(state) {
		t.Fatal("expected change to be detected")
	}

	plan.ConnectAddressEnabled = types.BoolValue(true)
	if plan.isConnectAddressEnabledChanged(state) {
		t.Fatal("expected no change")
	}

	// state from an older deployment that does not report the field
	state.ConnectAddressEnabled = types.BoolNull()
	if !plan.isConnectAddressEnabledChanged(state) {
		t.Fatal("expected change when state is null")
	}
}

// fakeConnectAddressStore records connect-address calls; every unrelated method
// panics so tests cannot accidentally depend on them.
type fakeConnectAddressStore struct {
	ClusterStore // nil; only the methods overridden below are callable

	privateLinkAddress string
	getErr             error
	enableErr          error
	disableErr         error
	enableCalled       bool
	disableCalled      bool
}

func (f *fakeConnectAddressStore) Get(ctx context.Context, clusterId string) (*ClusterResourceModel, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &ClusterResourceModel{
		ClusterId:          types.StringValue(clusterId),
		PrivateLinkAddress: types.StringValue(f.privateLinkAddress),
	}, nil
}

func (f *fakeConnectAddressStore) EnableConnectAddress(ctx context.Context, clusterId string) error {
	f.enableCalled = true
	return f.enableErr
}

func (f *fakeConnectAddressStore) DisableConnectAddress(ctx context.Context, clusterId string) error {
	f.disableCalled = true
	return f.disableErr
}

// runningClusterClient returns a client whose DescribeCluster always reports RUNNING,
// so waitForStatus succeeds on the first poll.
func runningClusterClient(t *testing.T) *zilliz.Client {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"clusterId": "in01-a", "status": "RUNNING"},
		}); err != nil {
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
	return c
}

func newConnectAddressResource(t *testing.T, store ClusterStore) *ClusterResource {
	t.Helper()
	return &ClusterResource{
		client:  runningClusterClient(t),
		store:   store,
		timeout: func() time.Duration { return time.Minute },
	}
}

func TestHandleConnectAddressUpdate(t *testing.T) {
	state := ClusterResourceModel{ClusterId: types.StringValue("in01-a")}

	t.Run("disable without private link fails fast before any API mutation", func(t *testing.T) {
		store := &fakeConnectAddressStore{privateLinkAddress: ""}
		plan := ClusterResourceModel{ConnectAddressEnabled: types.BoolValue(false)}

		diags := newConnectAddressResource(t, store).handleConnectAddressUpdate(context.Background(), plan, state)

		if !diags.HasError() {
			t.Fatal("expected an error diagnostic")
		}
		if got := diags.Errors()[0].Summary(); got != "Cannot disable connect address" {
			t.Fatalf("unexpected summary: %s", got)
		}
		if store.disableCalled || store.enableCalled {
			t.Fatal("no enable/disable call should be made when the precheck fails")
		}
	})

	t.Run("disable with private link calls the API", func(t *testing.T) {
		store := &fakeConnectAddressStore{privateLinkAddress: "https://in01-a-privatelink.example.com:19531"}
		plan := ClusterResourceModel{ConnectAddressEnabled: types.BoolValue(false)}

		diags := newConnectAddressResource(t, store).handleConnectAddressUpdate(context.Background(), plan, state)

		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if !store.disableCalled {
			t.Fatal("expected DisableConnectAddress to be called")
		}
		if store.enableCalled {
			t.Fatal("EnableConnectAddress must not be called")
		}
	})

	t.Run("enable calls the API without the private link precheck", func(t *testing.T) {
		store := &fakeConnectAddressStore{getErr: errors.New("get must not be called on the enable path")}
		plan := ClusterResourceModel{ConnectAddressEnabled: types.BoolValue(true)}

		diags := newConnectAddressResource(t, store).handleConnectAddressUpdate(context.Background(), plan, state)

		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if !store.enableCalled {
			t.Fatal("expected EnableConnectAddress to be called")
		}
	})

	t.Run("null or unknown plan value is a no-op", func(t *testing.T) {
		for _, v := range []types.Bool{types.BoolNull(), types.BoolUnknown()} {
			store := &fakeConnectAddressStore{}
			plan := ClusterResourceModel{ConnectAddressEnabled: v}

			diags := newConnectAddressResource(t, store).handleConnectAddressUpdate(context.Background(), plan, state)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if store.disableCalled || store.enableCalled {
				t.Fatal("no enable/disable call should be made for null/unknown plan values")
			}
		}
	})
}
