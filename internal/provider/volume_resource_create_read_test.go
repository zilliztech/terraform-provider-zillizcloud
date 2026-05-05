package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

type volumeResourceMockHTTPClient struct {
	t    *testing.T
	call int
	do   func(call int, req *http.Request, body []byte) (*http.Response, error)
}

func (m *volumeResourceMockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.t.Helper()
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			m.t.Fatalf("ReadAll request body: %v", err)
		}
	}
	m.call++
	return m.do(m.call, req, body)
}

func newTestVolumeResource(t *testing.T, do func(call int, req *http.Request, body []byte) (*http.Response, error)) *VolumeResource {
	t.Helper()
	client, err := zilliz.NewClient(
		zilliz.WithApiKey("test-key"),
		zilliz.WithBaseUrl("https://api.test/v2"),
		zilliz.WithHTTPClient(&volumeResourceMockHTTPClient{t: t, do: do}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return &VolumeResource{client: client}
}

func testVolumeResourceSchema(t *testing.T, r *VolumeResource) rschema.Schema {
	t.Helper()
	var resp fwresource.SchemaResponse
	r.Schema(context.Background(), fwresource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}
	return resp.Schema
}

func testVolumePlan(t *testing.T, ctx context.Context, schema rschema.Schema, model VolumeResourceModel) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: schema}
	diags := plan.Set(ctx, &model)
	if diags.HasError() {
		t.Fatalf("Plan.Set diagnostics: %s", diags.Errors()[0].Summary())
	}
	return plan
}

func testVolumeState(t *testing.T, ctx context.Context, schema rschema.Schema, model VolumeResourceModel) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: schema}
	diags := state.Set(ctx, &model)
	if diags.HasError() {
		t.Fatalf("State.Set diagnostics: %s", diags.Errors()[0].Summary())
	}
	return state
}

func volumeJSONResponse(t *testing.T, statusCode int, body any) *http.Response {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Marshal response: %v", err)
	}
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader(payload)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func TestVolumeResourceCreateRequiresStorageIntegrationForExternal(t *testing.T) {
	ctx := context.Background()
	clientCalled := false
	resource := newTestVolumeResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		clientCalled = true
		return nil, fmt.Errorf("client should not be called")
	})
	schema := testVolumeResourceSchema(t, resource)
	plan := testVolumePlan(t, ctx, schema, VolumeResourceModel{
		Id:                   types.StringUnknown(),
		ProjectId:            types.StringValue("proj-1"),
		RegionId:             types.StringValue("aws-us-west-2"),
		VolumeName:           types.StringValue("external-vol"),
		Type:                 types.StringValue("EXTERNAL"),
		StorageIntegrationId: types.StringNull(),
		Path:                 types.StringValue("datasets/"),
		Status:               types.StringUnknown(),
		CreateTime:           types.StringUnknown(),
	})

	var resp fwresource.CreateResponse
	resp.State = tfsdk.State{Schema: schema}
	resource.Create(ctx, fwresource.CreateRequest{Plan: plan}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected create diagnostics")
	}
	if resp.Diagnostics.Errors()[0].Summary() != "Missing storage integration ID" {
		t.Fatalf("diagnostic summary=%q", resp.Diagnostics.Errors()[0].Summary())
	}
	if clientCalled {
		t.Fatal("client should not be called when external storage_integration_id is missing")
	}
}

func TestVolumeResourceCreateCreatesVolumeAndHydratesState(t *testing.T) {
	ctx := context.Background()
	resource := newTestVolumeResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		switch call {
		case 1:
			if req.Method != http.MethodPost {
				t.Fatalf("create method=%s", req.Method)
			}
			if req.URL.Path != "/v2/volumes/create" {
				t.Fatalf("create path=%s", req.URL.Path)
			}
			var createReq zilliz.CreateVolumeRequest
			if err := json.Unmarshal(body, &createReq); err != nil {
				t.Fatalf("Unmarshal create request: %v", err)
			}
			if createReq.ProjectID != "proj-1" || createReq.RegionID != "aws-us-west-2" || createReq.VolumeName != "vol-1" || createReq.Type != "MANAGED" {
				t.Fatalf("create request=%+v", createReq)
			}
			if createReq.StorageIntegrationID != "" || createReq.Path != "" {
				t.Fatalf("unexpected optional create fields=%+v", createReq)
			}
			return volumeJSONResponse(t, http.StatusOK, map[string]any{
				"code": 0,
				"data": map[string]any{"volumeName": "vol-1"},
			}), nil
		case 2:
			if req.Method != http.MethodGet {
				t.Fatalf("describe method=%s", req.Method)
			}
			if req.URL.Path != "/v2/volumes/vol-1" {
				t.Fatalf("describe path=%s", req.URL.Path)
			}
			return volumeJSONResponse(t, http.StatusOK, map[string]any{
				"code": 0,
				"data": map[string]any{
					"volumeName": "vol-1",
					"type":       "MANAGED",
					"regionId":   "aws-us-west-2",
					"status":     "Available",
					"createTime": "2026-05-05T13:00:00Z",
				},
			}), nil
		default:
			return nil, fmt.Errorf("unexpected call %d", call)
		}
	})
	schema := testVolumeResourceSchema(t, resource)
	plan := testVolumePlan(t, ctx, schema, VolumeResourceModel{
		Id:                   types.StringUnknown(),
		ProjectId:            types.StringValue("proj-1"),
		RegionId:             types.StringValue("aws-us-west-2"),
		VolumeName:           types.StringValue("vol-1"),
		Type:                 types.StringValue("MANAGED"),
		StorageIntegrationId: types.StringNull(),
		Path:                 types.StringNull(),
		Status:               types.StringUnknown(),
		CreateTime:           types.StringUnknown(),
	})

	var resp fwresource.CreateResponse
	resp.State = tfsdk.State{Schema: schema}
	resource.Create(ctx, fwresource.CreateRequest{Plan: plan}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Create diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}

	var state VolumeResourceModel
	diags := resp.State.Get(ctx, &state)
	if diags.HasError() {
		t.Fatalf("State.Get diagnostics: %s", diags.Errors()[0].Summary())
	}
	if state.Id.ValueString() != "vol-1" {
		t.Fatalf("id=%q", state.Id.ValueString())
	}
	if state.Status.ValueString() != "Available" {
		t.Fatalf("status=%q", state.Status.ValueString())
	}
	if state.CreateTime.ValueString() != "2026-05-05T13:00:00Z" {
		t.Fatalf("create_time=%q", state.CreateTime.ValueString())
	}
	if !state.StorageIntegrationId.IsNull() {
		t.Fatalf("storage_integration_id should remain null, got %s", state.StorageIntegrationId.String())
	}
	if !state.Path.IsNull() {
		t.Fatalf("path should remain null, got %s", state.Path.String())
	}
}

func TestVolumeResourceReadRefreshesState(t *testing.T) {
	ctx := context.Background()
	resource := newTestVolumeResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		if call != 1 {
			return nil, fmt.Errorf("unexpected call %d", call)
		}
		if req.Method != http.MethodGet {
			t.Fatalf("describe method=%s", req.Method)
		}
		if req.URL.Path != "/v2/volumes/external-vol" {
			t.Fatalf("describe path=%s", req.URL.Path)
		}
		return volumeJSONResponse(t, http.StatusOK, map[string]any{
			"code": 0,
			"data": map[string]any{
				"volumeName":           "external-vol",
				"type":                 "EXTERNAL",
				"regionId":             "aws-us-west-2",
				"storageIntegrationId": "si-1",
				"path":                 "datasets/",
				"status":               "Available",
				"createTime":           "2026-05-05T13:10:00Z",
			},
		}), nil
	})
	schema := testVolumeResourceSchema(t, resource)
	state := testVolumeState(t, ctx, schema, VolumeResourceModel{
		Id:                   types.StringValue("external-vol"),
		ProjectId:            types.StringValue("proj-1"),
		RegionId:             types.StringValue("aws-us-west-2"),
		VolumeName:           types.StringValue("external-vol"),
		Type:                 types.StringValue("EXTERNAL"),
		StorageIntegrationId: types.StringValue("si-1"),
		Path:                 types.StringValue("datasets/"),
		Status:               types.StringValue("Creating"),
		CreateTime:           types.StringValue(""),
	})

	var resp fwresource.ReadResponse
	resp.State = tfsdk.State{Schema: schema}
	resource.Read(ctx, fwresource.ReadRequest{State: state}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Read diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}

	var refreshed VolumeResourceModel
	diags := resp.State.Get(ctx, &refreshed)
	if diags.HasError() {
		t.Fatalf("State.Get diagnostics: %s", diags.Errors()[0].Summary())
	}
	if refreshed.Status.ValueString() != "Available" {
		t.Fatalf("status=%q", refreshed.Status.ValueString())
	}
	if refreshed.CreateTime.ValueString() != "2026-05-05T13:10:00Z" {
		t.Fatalf("create_time=%q", refreshed.CreateTime.ValueString())
	}
	if refreshed.StorageIntegrationId.ValueString() != "si-1" {
		t.Fatalf("storage_integration_id=%q", refreshed.StorageIntegrationId.ValueString())
	}
	if refreshed.Path.ValueString() != "datasets/" {
		t.Fatalf("path=%q", refreshed.Path.ValueString())
	}
}

func TestVolumeResourceReadRemovesStateWhenVolumeNotFound(t *testing.T) {
	ctx := context.Background()
	resource := newTestVolumeResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		if call != 1 {
			return nil, fmt.Errorf("unexpected call %d", call)
		}
		if req.Method != http.MethodGet {
			t.Fatalf("describe method=%s", req.Method)
		}
		if req.URL.Path != "/v2/volumes/missing-vol" {
			t.Fatalf("describe path=%s", req.URL.Path)
		}
		return volumeJSONResponse(t, http.StatusNotFound, map[string]any{
			"code":      404,
			"message":   "volume not found",
			"requestId": "req-1",
		}), nil
	})
	schema := testVolumeResourceSchema(t, resource)
	state := testVolumeState(t, ctx, schema, VolumeResourceModel{
		Id:                   types.StringValue("missing-vol"),
		ProjectId:            types.StringValue("proj-1"),
		RegionId:             types.StringValue("aws-us-west-2"),
		VolumeName:           types.StringValue("missing-vol"),
		Type:                 types.StringValue("MANAGED"),
		StorageIntegrationId: types.StringNull(),
		Path:                 types.StringNull(),
		Status:               types.StringValue("Available"),
		CreateTime:           types.StringValue("2026-05-05T13:00:00Z"),
	})

	var resp fwresource.ReadResponse
	resp.State = tfsdk.State{Schema: schema}
	resource.Read(ctx, fwresource.ReadRequest{State: state}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Read diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}
	if !resp.State.Raw.IsNull() {
		t.Fatalf("state should be removed, got %#v", resp.State.Raw)
	}
}

func testVolumeDeletePolling(t *testing.T, timeout, interval time.Duration) {
	t.Helper()
	originalTimeout := volumeDeleteTimeout
	originalInterval := volumeDeletePollInterval
	volumeDeleteTimeout = timeout
	volumeDeletePollInterval = interval
	t.Cleanup(func() {
		volumeDeleteTimeout = originalTimeout
		volumeDeletePollInterval = originalInterval
	})
}

func TestVolumeResourceDeleteDeletesVolumeAndWaitsUntilNotFound(t *testing.T) {
	ctx := context.Background()
	testVolumeDeletePolling(t, time.Second, time.Millisecond)
	resource := newTestVolumeResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		switch call {
		case 1:
			if req.Method != http.MethodDelete {
				t.Fatalf("delete method=%s", req.Method)
			}
			if req.URL.Path != "/v2/volumes/vol-1" {
				t.Fatalf("delete path=%s", req.URL.Path)
			}
			return volumeJSONResponse(t, http.StatusOK, map[string]any{
				"code": 0,
				"data": map[string]any{"volumeName": "vol-1"},
			}), nil
		case 2:
			if req.Method != http.MethodGet {
				t.Fatalf("describe method=%s", req.Method)
			}
			if req.URL.Path != "/v2/volumes/vol-1" {
				t.Fatalf("describe path=%s", req.URL.Path)
			}
			return volumeJSONResponse(t, http.StatusOK, map[string]any{
				"code": 0,
				"data": map[string]any{
					"volumeName": "vol-1",
					"type":       "MANAGED",
					"regionId":   "aws-us-west-2",
					"status":     "Deleting",
				},
			}), nil
		case 3:
			if req.Method != http.MethodGet {
				t.Fatalf("describe method=%s", req.Method)
			}
			if req.URL.Path != "/v2/volumes/vol-1" {
				t.Fatalf("describe path=%s", req.URL.Path)
			}
			return volumeJSONResponse(t, http.StatusNotFound, map[string]any{
				"code":      404,
				"message":   "volume not found",
				"requestId": "req-1",
			}), nil
		default:
			return nil, fmt.Errorf("unexpected call %d", call)
		}
	})
	schema := testVolumeResourceSchema(t, resource)
	state := testVolumeState(t, ctx, schema, VolumeResourceModel{
		Id:                   types.StringValue("vol-1"),
		ProjectId:            types.StringValue("proj-1"),
		RegionId:             types.StringValue("aws-us-west-2"),
		VolumeName:           types.StringValue("vol-1"),
		Type:                 types.StringValue("MANAGED"),
		StorageIntegrationId: types.StringNull(),
		Path:                 types.StringNull(),
		Status:               types.StringValue("Available"),
		CreateTime:           types.StringValue("2026-05-05T13:00:00Z"),
	})

	var resp fwresource.DeleteResponse
	resource.Delete(ctx, fwresource.DeleteRequest{State: state}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Delete diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}
}

func TestVolumeResourceDeleteSucceedsWhenAlreadyGone(t *testing.T) {
	ctx := context.Background()
	resource := newTestVolumeResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		if call != 1 {
			return nil, fmt.Errorf("unexpected call %d", call)
		}
		if req.Method != http.MethodDelete {
			t.Fatalf("delete method=%s", req.Method)
		}
		if req.URL.Path != "/v2/volumes/missing-vol" {
			t.Fatalf("delete path=%s", req.URL.Path)
		}
		return volumeJSONResponse(t, http.StatusNotFound, map[string]any{
			"code":      404,
			"message":   "volume not found",
			"requestId": "req-1",
		}), nil
	})
	schema := testVolumeResourceSchema(t, resource)
	state := testVolumeState(t, ctx, schema, VolumeResourceModel{
		Id:                   types.StringValue("missing-vol"),
		ProjectId:            types.StringValue("proj-1"),
		RegionId:             types.StringValue("aws-us-west-2"),
		VolumeName:           types.StringValue("missing-vol"),
		Type:                 types.StringValue("MANAGED"),
		StorageIntegrationId: types.StringNull(),
		Path:                 types.StringNull(),
		Status:               types.StringValue("Available"),
		CreateTime:           types.StringValue("2026-05-05T13:00:00Z"),
	})

	var resp fwresource.DeleteResponse
	resource.Delete(ctx, fwresource.DeleteRequest{State: state}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Delete diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}
}

func TestVolumeResourceDeleteReportsDeleteError(t *testing.T) {
	ctx := context.Background()
	resource := newTestVolumeResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		if call != 1 {
			return nil, fmt.Errorf("unexpected call %d", call)
		}
		if req.Method != http.MethodDelete {
			t.Fatalf("delete method=%s", req.Method)
		}
		return volumeJSONResponse(t, http.StatusInternalServerError, map[string]any{
			"code":      500,
			"message":   "temporary failure",
			"requestId": "req-1",
		}), nil
	})
	schema := testVolumeResourceSchema(t, resource)
	state := testVolumeState(t, ctx, schema, VolumeResourceModel{
		Id:                   types.StringValue("vol-1"),
		ProjectId:            types.StringValue("proj-1"),
		RegionId:             types.StringValue("aws-us-west-2"),
		VolumeName:           types.StringValue("vol-1"),
		Type:                 types.StringValue("MANAGED"),
		StorageIntegrationId: types.StringNull(),
		Path:                 types.StringNull(),
		Status:               types.StringValue("Available"),
		CreateTime:           types.StringValue("2026-05-05T13:00:00Z"),
	})

	var resp fwresource.DeleteResponse
	resource.Delete(ctx, fwresource.DeleteRequest{State: state}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected delete diagnostics")
	}
	if resp.Diagnostics.Errors()[0].Summary() != "Failed to delete volume" {
		t.Fatalf("diagnostic summary=%q", resp.Diagnostics.Errors()[0].Summary())
	}
}

func TestVolumeResourceDeleteReportsTimeout(t *testing.T) {
	ctx := context.Background()
	testVolumeDeletePolling(t, 5*time.Millisecond, time.Millisecond)
	resource := newTestVolumeResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		switch call {
		case 1:
			if req.Method != http.MethodDelete {
				t.Fatalf("delete method=%s", req.Method)
			}
			return volumeJSONResponse(t, http.StatusOK, map[string]any{
				"code": 0,
				"data": map[string]any{"volumeName": "vol-1"},
			}), nil
		default:
			if req.Method != http.MethodGet {
				t.Fatalf("describe method=%s", req.Method)
			}
			if req.URL.Path != "/v2/volumes/vol-1" {
				t.Fatalf("describe path=%s", req.URL.Path)
			}
			return volumeJSONResponse(t, http.StatusOK, map[string]any{
				"code": 0,
				"data": map[string]any{
					"volumeName": "vol-1",
					"type":       "MANAGED",
					"regionId":   "aws-us-west-2",
					"status":     "Deleting",
				},
			}), nil
		}
	})
	schema := testVolumeResourceSchema(t, resource)
	state := testVolumeState(t, ctx, schema, VolumeResourceModel{
		Id:                   types.StringValue("vol-1"),
		ProjectId:            types.StringValue("proj-1"),
		RegionId:             types.StringValue("aws-us-west-2"),
		VolumeName:           types.StringValue("vol-1"),
		Type:                 types.StringValue("MANAGED"),
		StorageIntegrationId: types.StringNull(),
		Path:                 types.StringNull(),
		Status:               types.StringValue("Available"),
		CreateTime:           types.StringValue("2026-05-05T13:00:00Z"),
	})

	var resp fwresource.DeleteResponse
	resource.Delete(ctx, fwresource.DeleteRequest{State: state}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected timeout diagnostics")
	}
	if resp.Diagnostics.Errors()[0].Summary() != "Timed out deleting volume" {
		t.Fatalf("diagnostic summary=%q", resp.Diagnostics.Errors()[0].Summary())
	}
}
