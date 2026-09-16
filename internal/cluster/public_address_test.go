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

func TestIsPublicAddressEnabledChanged(t *testing.T) {
	plan := ClusterResourceModel{PublicAddressEnabled: types.BoolValue(false)}
	state := ClusterResourceModel{PublicAddressEnabled: types.BoolValue(true)}
	if !plan.isPublicAddressEnabledChanged(state) {
		t.Fatal("expected change to be detected")
	}

	plan.PublicAddressEnabled = types.BoolValue(true)
	if plan.isPublicAddressEnabledChanged(state) {
		t.Fatal("expected no change")
	}

	// state from a cluster the toggle does not apply to (BYOC — API omits the field)
	state.PublicAddressEnabled = types.BoolNull()
	if !plan.isPublicAddressEnabledChanged(state) {
		t.Fatal("expected change when state is null")
	}

	// unconfigured attribute tracks the remote value: after a refresh observes a
	// console-side disable, both plan and state hold false and no diff is produced
	plan.PublicAddressEnabled = types.BoolValue(false)
	state.PublicAddressEnabled = types.BoolValue(false)
	if plan.isPublicAddressEnabledChanged(state) {
		t.Fatal("expected no change when plan tracks the remote false")
	}
}

// fakePublicAddressStore records UpdatePublicAddressEnabled calls; every unrelated method
// panics so tests cannot accidentally depend on them.
type fakePublicAddressStore struct {
	ClusterStore // nil; only the methods overridden below are callable

	privateLinkAddress string
	getErr             error
	setErr             error
	setCalled          bool
	setEnabled         bool
}

func (f *fakePublicAddressStore) Get(ctx context.Context, clusterId string) (*ClusterResourceModel, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &ClusterResourceModel{
		ClusterId:          types.StringValue(clusterId),
		PrivateLinkAddress: types.StringValue(f.privateLinkAddress),
	}, nil
}

func (f *fakePublicAddressStore) UpdatePublicAddressEnabled(ctx context.Context, clusterId string, enabled bool) error {
	f.setCalled = true
	f.setEnabled = enabled
	return f.setErr
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

func newPublicAddressResource(t *testing.T, store ClusterStore) *ClusterResource {
	t.Helper()
	return &ClusterResource{
		client:  runningClusterClient(t),
		store:   store,
		timeout: func() time.Duration { return time.Minute },
	}
}

func TestHandlePublicAddressUpdate(t *testing.T) {
	enabledState := ClusterResourceModel{
		ClusterId:            types.StringValue("in01-a"),
		PublicAddressEnabled: types.BoolValue(true),
	}

	t.Run("disable without private link fails fast before any API mutation", func(t *testing.T) {
		store := &fakePublicAddressStore{privateLinkAddress: ""}
		plan := ClusterResourceModel{PublicAddressEnabled: types.BoolValue(false)}

		diags := newPublicAddressResource(t, store).handlePublicAddressUpdate(context.Background(), plan, enabledState)

		if !diags.HasError() {
			t.Fatal("expected an error diagnostic")
		}
		if got := diags.Errors()[0].Summary(); got != "Cannot disable public address" {
			t.Fatalf("unexpected summary: %s", got)
		}
		if store.setCalled {
			t.Fatal("no UpdatePublicAddressEnabled call should be made when the precheck fails")
		}
	})

	t.Run("disable with private link calls the API with false", func(t *testing.T) {
		store := &fakePublicAddressStore{privateLinkAddress: "https://in01-a-privatelink.example.com:19531"}
		plan := ClusterResourceModel{PublicAddressEnabled: types.BoolValue(false)}

		diags := newPublicAddressResource(t, store).handlePublicAddressUpdate(context.Background(), plan, enabledState)

		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if !store.setCalled {
			t.Fatal("expected UpdatePublicAddressEnabled to be called")
		}
		if store.setEnabled {
			t.Fatal("expected UpdatePublicAddressEnabled to be called with false")
		}
	})

	t.Run("enable calls the API with true without the private link precheck", func(t *testing.T) {
		store := &fakePublicAddressStore{getErr: errors.New("get must not be called on the enable path")}
		plan := ClusterResourceModel{PublicAddressEnabled: types.BoolValue(true)}
		disabledState := ClusterResourceModel{
			ClusterId:            types.StringValue("in01-a"),
			PublicAddressEnabled: types.BoolValue(false),
		}

		diags := newPublicAddressResource(t, store).handlePublicAddressUpdate(context.Background(), plan, disabledState)

		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if !store.setCalled {
			t.Fatal("expected UpdatePublicAddressEnabled to be called")
		}
		if !store.setEnabled {
			t.Fatal("expected UpdatePublicAddressEnabled to be called with true")
		}
	})

	t.Run("null or unknown plan value is a no-op", func(t *testing.T) {
		for _, v := range []types.Bool{types.BoolNull(), types.BoolUnknown()} {
			store := &fakePublicAddressStore{}
			plan := ClusterResourceModel{PublicAddressEnabled: v}

			diags := newPublicAddressResource(t, store).handlePublicAddressUpdate(context.Background(), plan, enabledState)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if store.setCalled {
				t.Fatal("no UpdatePublicAddressEnabled call should be made for null/unknown plan values")
			}
		}
	})

	t.Run("configured value against a null (not-applicable) state is rejected", func(t *testing.T) {
		// BYOC: the API omits the field, so state is null. Any explicit value is a
		// misconfiguration and must be rejected before any API call.
		nullState := ClusterResourceModel{
			ClusterId:            types.StringValue("in01-a"),
			PublicAddressEnabled: types.BoolNull(),
		}
		for _, v := range []types.Bool{types.BoolValue(true), types.BoolValue(false)} {
			store := &fakePublicAddressStore{}
			plan := ClusterResourceModel{PublicAddressEnabled: v}

			diags := newPublicAddressResource(t, store).handlePublicAddressUpdate(context.Background(), plan, nullState)

			if !diags.HasError() {
				t.Fatalf("expected an error diagnostic for value %v", v)
			}
			if got := diags.Errors()[0].Summary(); got != "public_address_enabled does not apply to this cluster" {
				t.Fatalf("unexpected summary: %s", got)
			}
			if store.setCalled {
				t.Fatal("no UpdatePublicAddressEnabled call should be made when the toggle does not apply")
			}
		}
	})
}
