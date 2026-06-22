package global_cluster

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

type globalClusterResourceMockHTTPClient struct {
	t    *testing.T
	call int
	do   func(call int, req *http.Request, body []byte) (*http.Response, error)
}

func (m *globalClusterResourceMockHTTPClient) Do(req *http.Request) (*http.Response, error) {
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

func newTestGlobalClusterClient(t *testing.T, do func(call int, req *http.Request, body []byte) (*http.Response, error)) *zilliz.Client {
	t.Helper()
	client, err := zilliz.NewClient(
		zilliz.WithApiKey("test-key"),
		zilliz.WithBaseUrl("https://api.test/v2"),
		zilliz.WithHTTPClient(&globalClusterResourceMockHTTPClient{t: t, do: do}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

func newTestGlobalClusterResource(t *testing.T, do func(call int, req *http.Request, body []byte) (*http.Response, error)) *GlobalClusterResource {
	t.Helper()
	client := newTestGlobalClusterClient(t, do)
	resource := newGlobalClusterResource()
	resource.client = client
	resource.store = NewGlobalClusterStore(client)
	return resource
}

func testGlobalClusterPostCreateDescribeDelay(t *testing.T, delay time.Duration) {
	t.Helper()
	originalDelay := globalClusterPostCreateDescribeDelay
	globalClusterPostCreateDescribeDelay = delay
	t.Cleanup(func() {
		globalClusterPostCreateDescribeDelay = originalDelay
	})
}

func testGlobalClusterSecondaryDeleteWait(t *testing.T, pollInterval time.Duration, timeout time.Duration) {
	t.Helper()
	originalPollInterval := globalClusterSecondaryPollInterval
	originalTimeout := globalClusterSecondaryDeleteTimeout
	globalClusterSecondaryPollInterval = pollInterval
	globalClusterSecondaryDeleteTimeout = timeout
	t.Cleanup(func() {
		globalClusterSecondaryPollInterval = originalPollInterval
		globalClusterSecondaryDeleteTimeout = originalTimeout
	})
}

func testGlobalClusterSecondaryRunningWait(t *testing.T, pollInterval time.Duration, timeout time.Duration) {
	t.Helper()
	originalPollInterval := globalClusterSecondaryPollInterval
	originalTimeout := globalClusterSecondaryRunningTimeout
	globalClusterSecondaryPollInterval = pollInterval
	globalClusterSecondaryRunningTimeout = timeout
	t.Cleanup(func() {
		globalClusterSecondaryPollInterval = originalPollInterval
		globalClusterSecondaryRunningTimeout = originalTimeout
	})
}

func testGlobalClusterResourceSchema(t *testing.T, r fwresource.Resource) rschema.Schema {
	t.Helper()
	var resp fwresource.SchemaResponse
	r.Schema(context.Background(), fwresource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}
	return resp.Schema
}

func TestGlobalClusterResourceSchemaDoesNotPreserveRegionIDsDuringPlanning(t *testing.T) {
	schema := testGlobalClusterResourceSchema(t, newGlobalClusterResource())
	attr, ok := schema.Attributes["region_ids"].(rschema.ListAttribute)
	if !ok {
		t.Fatalf("region_ids attribute = %T, want schema.ListAttribute", schema.Attributes["region_ids"])
	}
	if len(attr.PlanModifiers) != 0 {
		t.Fatalf("region_ids should not preserve prior state during planning; got %d plan modifiers", len(attr.PlanModifiers))
	}
}

func testGlobalClusterPlan(t *testing.T, ctx context.Context, schema rschema.Schema, model GlobalClusterResourceModel) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: schema}
	diags := plan.Set(ctx, &model)
	if diags.HasError() {
		t.Fatalf("Plan.Set diagnostics: %s: %s", diags.Errors()[0].Summary(), diags.Errors()[0].Detail())
	}
	return plan
}

func testGlobalClusterState(t *testing.T, ctx context.Context, schema rschema.Schema, model GlobalClusterResourceModel) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: schema}
	diags := state.Set(ctx, &model)
	if diags.HasError() {
		t.Fatalf("State.Set diagnostics: %s: %s", diags.Errors()[0].Summary(), diags.Errors()[0].Detail())
	}
	return state
}

func globalClusterJSONResponse(t *testing.T, statusCode int, body any) *http.Response {
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

func describeGlobalClusterPayload(cuSize int) map[string]any {
	payload, _, _ := describeGlobalClusterPayloadParts(cuSize)
	return payload
}

func describeGlobalClusterPayloadParts(cuSize int) (map[string]any, map[string]any, []map[string]any) {
	clusters := []map[string]any{
		{
			"clusterId":   "in01-primary",
			"clusterName": "primary-a",
			"regionId":    "aws-us-west-2",
			"role":        "PRIMARY",
			"status":      "RUNNING",
		},
		{
			"clusterId":   "in01-secondary",
			"clusterName": "secondary-eu",
			"regionId":    "aws-eu-west-1",
			"role":        "SECONDARY",
			"status":      "RUNNING",
		},
		{
			"clusterId":   "in01-secondary-ap",
			"clusterName": "secondary-ap",
			"regionId":    "aws-ap-southeast-1",
			"role":        "SECONDARY",
			"status":      "RUNNING",
		},
	}
	data := map[string]any{
		"globalClusterId":   "glo-1",
		"globalClusterName": "global-a",
		"projectId":         "proj-1",
		"regionIds":         []string{"aws-us-west-2", "aws-eu-west-1", "aws-ap-southeast-1"},
		"cuType":            "Performance-optimized",
		"cuSize":            cuSize,
		"connectAddress":    "https://glo-1.global-cluster.vectordb.zillizcloud.com",
		"createTime":        "2026-06-04T10:00:00Z",
		"clusters":          clusters,
	}
	return map[string]any{
		"code": 0,
		"data": data,
	}, data, clusters
}

func describeGlobalClusterPayloadWithSecondaryAPDeleting(cuSize int) map[string]any {
	payload, _, clusters := describeGlobalClusterPayloadParts(cuSize)
	clusters[2]["status"] = "DELETING"
	return payload
}

func describeGlobalClusterPayloadWithSecondaryAPCreating(cuSize int) map[string]any {
	payload, _, clusters := describeGlobalClusterPayloadParts(cuSize)
	clusters[2]["status"] = "CREATING"
	return payload
}

func describeGlobalClusterPayloadWithSecondaryAU(cuSize int, status string) map[string]any {
	payload, data, clusters := describeGlobalClusterPayloadParts(cuSize)
	clusters = append(clusters, map[string]any{
		"clusterId":   "in01-secondary-au",
		"clusterName": "secondary-au",
		"regionId":    "aws-ap-southeast-2",
		"role":        "SECONDARY",
		"status":      status,
	})
	data["clusters"] = clusters
	data["regionIds"] = []string{"aws-us-west-2", "aws-eu-west-1", "aws-ap-southeast-1", "aws-ap-southeast-2"}
	return payload
}

func describeGlobalClusterPayloadWithoutSecondaryAP(cuSize int) map[string]any {
	payload, data, clusters := describeGlobalClusterPayloadParts(cuSize)
	data["clusters"] = clusters[:2]
	data["regionIds"] = []string{"aws-us-west-2", "aws-eu-west-1"}
	return payload
}

func describeGlobalClusterPayloadWithoutSecondaryEU(cuSize int) map[string]any {
	payload, data, clusters := describeGlobalClusterPayloadParts(cuSize)
	data["clusters"] = []map[string]any{clusters[0], clusters[2]}
	data["regionIds"] = []string{"aws-us-west-2", "aws-ap-southeast-1"}
	return payload
}

func describeGlobalClusterPayloadWithoutSecondaries() map[string]any {
	payload, data, clusters := describeGlobalClusterPayloadParts(4)
	data["clusters"] = clusters[:1]
	data["regionIds"] = []string{"aws-us-west-2"}
	return payload
}

func describeGlobalClusterPayloadWithLockedPrimary(t *testing.T) map[string]any {
	t.Helper()
	payload := describeGlobalClusterPayloadWithoutSecondaries()
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("payload data type = %T, want map[string]any", payload["data"])
	}
	clusters, ok := data["clusters"].([]map[string]any)
	if !ok {
		t.Fatalf("payload clusters type = %T, want []map[string]any", data["clusters"])
	}
	clusters[0]["status"] = "LOCKED"
	return payload
}

func testGlobalClusterBaseModel() GlobalClusterResourceModel {
	return GlobalClusterResourceModel{
		ID:                types.StringValue("glo-1"),
		GlobalClusterName: types.StringValue("global-a"),
		ProjectID:         types.StringValue("proj-1"),
		CUType:            types.StringValue("Performance-optimized"),
		CUSize:            types.Int64Value(4),
		Cluster: []GlobalClusterMemberModel{
			{ClusterID: types.StringValue("in01-primary"), ClusterName: types.StringValue("primary-a"), RegionID: types.StringValue("aws-us-west-2"), Role: types.StringValue("PRIMARY"), Status: types.StringValue("RUNNING")},
			{ClusterID: types.StringValue("in01-secondary"), ClusterName: types.StringValue("secondary-eu"), RegionID: types.StringValue("aws-eu-west-1"), Role: types.StringValue("SECONDARY"), Status: types.StringValue("RUNNING")},
			{ClusterID: types.StringValue("in01-secondary-ap"), ClusterName: types.StringValue("secondary-ap"), RegionID: types.StringValue("aws-ap-southeast-1"), Role: types.StringValue("SECONDARY"), Status: types.StringValue("RUNNING")},
		},
		ConnectAddress: types.StringValue("https://glo-1.global-cluster.vectordb.zillizcloud.com"),
		CreateTime:     types.StringValue("2026-06-04T10:00:00Z"),
		RegionIDs:      types.ListUnknown(types.StringType),
		Username:       types.StringValue("db_admin"),
		Password:       types.StringValue("password"),
		CreateJobID:    types.StringValue("job-create-1"),
	}
}

func TestGlobalClusterResourceCreateCreatesAndHydratesState(t *testing.T) {
	ctx := context.Background()
	testGlobalClusterPostCreateDescribeDelay(t, 0)
	resource := newTestGlobalClusterResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		switch call {
		case 1:
			if req.Method != http.MethodPost || req.URL.Path != "/v2/globalClusters/create" {
				t.Fatalf("create %s %s", req.Method, req.URL.Path)
			}
			var createReq zilliz.CreateGlobalClusterParams
			if err := json.Unmarshal(body, &createReq); err != nil {
				t.Fatalf("Unmarshal create request: %v", err)
			}
			if createReq.GlobalClusterName != "global-a" || createReq.ProjectId != "proj-1" || createReq.CuSize != 4 {
				t.Fatalf("create request=%+v", createReq)
			}
			if createReq.PrimaryCluster.ClusterName != "primary-a" || createReq.PrimaryCluster.RegionId != "aws-us-west-2" {
				t.Fatalf("primary request=%+v", createReq.PrimaryCluster)
			}
			if len(createReq.SecondaryClusters) != 2 {
				t.Fatalf("secondary request=%+v", createReq.SecondaryClusters)
			}
			if createReq.SecondaryClusters[0].ClusterName != "secondary-eu" || createReq.SecondaryClusters[0].RegionId != "aws-eu-west-1" {
				t.Fatalf("secondary request=%+v", createReq.SecondaryClusters[0])
			}
			if createReq.SecondaryClusters[1].ClusterName != "secondary-ap" || createReq.SecondaryClusters[1].RegionId != "aws-ap-southeast-1" {
				t.Fatalf("secondary request=%+v", createReq.SecondaryClusters[1])
			}
			return globalClusterJSONResponse(t, http.StatusOK, map[string]any{
				"code": 0,
				"data": map[string]any{
					"globalClusterId": "glo-1",
					"username":        "db_admin",
					"password":        "password",
					"jobId":           "job-create-1",
				},
			}), nil
		case 2:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayload(4)), nil
		default:
			return nil, fmt.Errorf("unexpected call %d", call)
		}
	})
	schema := testGlobalClusterResourceSchema(t, resource)
	plan := testGlobalClusterPlan(t, ctx, schema, GlobalClusterResourceModel{
		ID:                types.StringUnknown(),
		GlobalClusterName: types.StringValue("global-a"),
		ProjectID:         types.StringValue("proj-1"),
		CUType:            types.StringValue("Performance-optimized"),
		CUSize:            types.Int64Value(4),
		Cluster: []GlobalClusterMemberModel{
			{
				ClusterID:   types.StringUnknown(),
				ClusterName: types.StringValue("primary-a"),
				RegionID:    types.StringValue("aws-us-west-2"),
				Role:        types.StringUnknown(),
				Status:      types.StringUnknown(),
			},
			{
				ClusterID:   types.StringUnknown(),
				ClusterName: types.StringValue("secondary-eu"),
				RegionID:    types.StringValue("aws-eu-west-1"),
				Role:        types.StringUnknown(),
				Status:      types.StringUnknown(),
			},
			{
				ClusterID:   types.StringUnknown(),
				ClusterName: types.StringValue("secondary-ap"),
				RegionID:    types.StringValue("aws-ap-southeast-1"),
				Role:        types.StringUnknown(),
				Status:      types.StringUnknown(),
			},
		},
		ConnectAddress: types.StringUnknown(),
		CreateTime:     types.StringUnknown(),
		RegionIDs:      types.ListUnknown(types.StringType),
		Username:       types.StringUnknown(),
		Password:       types.StringUnknown(),
		CreateJobID:    types.StringUnknown(),
	})

	var resp fwresource.CreateResponse
	resp.State = tfsdk.State{Schema: schema}
	resource.Create(ctx, fwresource.CreateRequest{Plan: plan}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Create diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}

	var state GlobalClusterResourceModel
	diags := resp.State.Get(ctx, &state)
	if diags.HasError() {
		t.Fatalf("State.Get diagnostics: %s", diags.Errors()[0].Summary())
	}
	if state.ID.ValueString() != "glo-1" {
		t.Fatalf("unexpected global id=%s", state.ID.ValueString())
	}
	if len(state.Cluster) != 3 {
		t.Fatalf("unexpected cluster length=%d", len(state.Cluster))
	}
	if state.Cluster[0].ClusterName.ValueString() != "primary-a" || state.Cluster[0].RegionID.ValueString() != "aws-us-west-2" {
		t.Fatalf("unexpected primary cluster=%+v", state.Cluster[0])
	}
	if state.Cluster[1].ClusterName.ValueString() != "secondary-eu" || state.Cluster[1].RegionID.ValueString() != "aws-eu-west-1" {
		t.Fatalf("unexpected first secondary cluster=%+v", state.Cluster[1])
	}
	if state.Cluster[2].ClusterName.ValueString() != "secondary-ap" || state.Cluster[2].RegionID.ValueString() != "aws-ap-southeast-1" {
		t.Fatalf("unexpected second secondary cluster=%+v", state.Cluster[2])
	}
	if state.Cluster[0].ClusterID.ValueString() != "in01-primary" || state.Cluster[0].Role.ValueString() != "PRIMARY" || state.Cluster[0].Status.ValueString() != "RUNNING" {
		t.Fatalf("unexpected primary computed fields=%+v", state.Cluster[0])
	}
	if state.Cluster[1].ClusterID.ValueString() != "in01-secondary" || state.Cluster[1].Role.ValueString() != "SECONDARY" || state.Cluster[1].Status.ValueString() != "RUNNING" {
		t.Fatalf("unexpected first secondary computed fields=%+v", state.Cluster[1])
	}
	if state.ConnectAddress.ValueString() == "" || state.CreateTime.ValueString() == "" {
		t.Fatalf("missing computed endpoint/time")
	}
	if state.Username.ValueString() != "db_admin" || state.Password.ValueString() != "password" {
		t.Fatalf("unexpected credentials")
	}
	if state.CreateJobID.ValueString() != "job-create-1" {
		t.Fatalf("unexpected create job id=%s", state.CreateJobID.ValueString())
	}
}

func TestGlobalClusterResourceCreateDelegatesMemberValidationToBackend(t *testing.T) {
	ctx := context.Background()
	resource := newTestGlobalClusterResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		if call != 1 || req.Method != http.MethodPost || req.URL.Path != "/v2/globalClusters/create" {
			t.Fatalf("unexpected call %d %s %s", call, req.Method, req.URL.Path)
		}
		return globalClusterJSONResponse(t, http.StatusBadRequest, map[string]any{"code": 400, "message": "secondaryClusters is required"}), nil
	})
	schema := testGlobalClusterResourceSchema(t, resource)
	model := testGlobalClusterBaseModel()
	model.ID = types.StringUnknown()
	model.Cluster = model.Cluster[:1]
	model.ConnectAddress = types.StringUnknown()
	model.CreateTime = types.StringUnknown()
	model.RegionIDs = types.ListUnknown(types.StringType)
	model.Username = types.StringUnknown()
	model.Password = types.StringUnknown()
	model.CreateJobID = types.StringUnknown()
	plan := testGlobalClusterPlan(t, ctx, schema, model)

	var resp fwresource.CreateResponse
	resp.State = tfsdk.State{Schema: schema}
	resource.Create(ctx, fwresource.CreateRequest{Plan: plan}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected create diagnostics")
	}
}

func TestGlobalClusterResourceUpdateModifiesCU(t *testing.T) {
	ctx := context.Background()
	resource := newTestGlobalClusterResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		switch call {
		case 1:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe before update %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayload(4)), nil
		case 2:
			if req.Method != http.MethodPost || req.URL.Path != "/v2/globalClusters/glo-1/modifyCU" {
				t.Fatalf("modify %s %s", req.Method, req.URL.Path)
			}
			var modifyReq zilliz.ModifyGlobalClusterCUParams
			if err := json.Unmarshal(body, &modifyReq); err != nil {
				t.Fatalf("Unmarshal modify request: %v", err)
			}
			if modifyReq.CuSize != 8 {
				t.Fatalf("cuSize=%d", modifyReq.CuSize)
			}
			return globalClusterJSONResponse(t, http.StatusOK, map[string]any{"code": 0, "data": map[string]any{"jobId": "job-1"}}), nil
		case 3:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe after cu modify %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayload(8)), nil
		case 4:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("final describe after update %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayload(8)), nil
		default:
			return nil, fmt.Errorf("unexpected call %d", call)
		}
	})
	schema := testGlobalClusterResourceSchema(t, resource)
	base := testGlobalClusterBaseModel()
	state := testGlobalClusterState(t, ctx, schema, base)
	base.CUSize = types.Int64Value(8)
	plan := testGlobalClusterPlan(t, ctx, schema, base)

	var resp fwresource.UpdateResponse
	resp.State = tfsdk.State{Schema: schema}
	resource.Update(ctx, fwresource.UpdateRequest{Plan: plan, State: state}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Update diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}
}

func TestGlobalClusterResourceUpdateWaitsForCUSizeToConverge(t *testing.T) {
	ctx := context.Background()
	testGlobalClusterSecondaryRunningWait(t, time.Millisecond, time.Second)
	resource := newTestGlobalClusterResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		switch call {
		case 1:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe before update %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayload(1)), nil
		case 2:
			if req.Method != http.MethodPost || req.URL.Path != "/v2/globalClusters/glo-1/modifyCU" {
				t.Fatalf("modify %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, map[string]any{"code": 0, "data": map[string]any{"jobId": "job-1"}}), nil
		case 3:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe while cu modifying %s %s", req.Method, req.URL.Path)
			}
			payload, _, clusters := describeGlobalClusterPayloadParts(1)
			for _, cluster := range clusters {
				cluster["status"] = "CU_MODIFYING"
			}
			return globalClusterJSONResponse(t, http.StatusOK, payload), nil
		case 4:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe after cu converged %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayload(2)), nil
		case 5:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("final describe after update %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayload(2)), nil
		default:
			return nil, fmt.Errorf("unexpected call %d", call)
		}
	})
	schema := testGlobalClusterResourceSchema(t, resource)
	base := testGlobalClusterBaseModel()
	base.CUSize = types.Int64Value(1)
	state := testGlobalClusterState(t, ctx, schema, base)
	base.CUSize = types.Int64Value(2)
	plan := testGlobalClusterPlan(t, ctx, schema, base)

	var resp fwresource.UpdateResponse
	resp.State = tfsdk.State{Schema: schema}
	resource.Update(ctx, fwresource.UpdateRequest{Plan: plan, State: state}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Update diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}

	var got GlobalClusterResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("State.Get diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}
	if got.CUSize.ValueInt64() != 2 {
		t.Fatalf("cu_size after update=%d, want 2", got.CUSize.ValueInt64())
	}
}

func TestGlobalClusterResourceUpdateAddsSecondaryCluster(t *testing.T) {
	ctx := context.Background()
	testGlobalClusterSecondaryRunningWait(t, time.Millisecond, time.Second)
	resource := newTestGlobalClusterResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		switch call {
		case 1:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe before update %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayload(4)), nil
		case 2:
			if req.Method != http.MethodPost || req.URL.Path != "/v2/globalClusters/glo-1/secondaryClusters" {
				t.Fatalf("add secondary %s %s", req.Method, req.URL.Path)
			}
			var addReq zilliz.AddSecondaryClustersParams
			if err := json.Unmarshal(body, &addReq); err != nil {
				t.Fatalf("Unmarshal add request: %v", err)
			}
			if len(addReq.SecondaryClusters) != 1 || addReq.SecondaryClusters[0].ClusterName != "secondary-au" || addReq.SecondaryClusters[0].RegionId != "aws-ap-southeast-2" {
				t.Fatalf("add request=%+v", addReq.SecondaryClusters)
			}
			return globalClusterJSONResponse(t, http.StatusOK, map[string]any{"code": 0, "data": map[string]any{"jobId": "job-add-1"}}), nil
		case 3:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe while secondary is creating %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayloadWithSecondaryAU(4, "CREATING")), nil
		case 4:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe after secondary running %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayloadWithSecondaryAU(4, "RUNNING")), nil
		case 5:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("final describe after update %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayloadWithSecondaryAU(4, "RUNNING")), nil
		default:
			return nil, fmt.Errorf("unexpected call %d", call)
		}
	})
	schema := testGlobalClusterResourceSchema(t, resource)
	base := testGlobalClusterBaseModel()
	state := testGlobalClusterState(t, ctx, schema, base)
	base.Cluster = append(base.Cluster, GlobalClusterMemberModel{ClusterID: types.StringUnknown(), ClusterName: types.StringValue("secondary-au"), RegionID: types.StringValue("aws-ap-southeast-2"), Role: types.StringUnknown(), Status: types.StringUnknown()})
	plan := testGlobalClusterPlan(t, ctx, schema, base)

	var resp fwresource.UpdateResponse
	resp.State = tfsdk.State{Schema: schema}
	resource.Update(ctx, fwresource.UpdateRequest{Plan: plan, State: state}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Update diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}

	var got GlobalClusterResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("State.Get diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}
	var regionIDs []string
	resp.Diagnostics.Append(got.RegionIDs.ElementsAs(ctx, &regionIDs, false)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("RegionIDs.ElementsAs diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}
	if len(regionIDs) != 4 || regionIDs[3] != "aws-ap-southeast-2" {
		t.Fatalf("unexpected region_ids after secondary add: %#v", regionIDs)
	}
}

func TestGlobalClusterResourceUpdateRemovesSecondaryCluster(t *testing.T) {
	ctx := context.Background()
	testGlobalClusterSecondaryDeleteWait(t, time.Millisecond, time.Second)
	resource := newTestGlobalClusterResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		switch call {
		case 1:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe before update %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayload(4)), nil
		case 2:
			if req.Method != http.MethodDelete || req.URL.Path != "/v2/globalClusters/glo-1/clusters/in01-secondary-ap" {
				t.Fatalf("delete secondary %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, map[string]any{"code": 0, "data": map[string]any{"globalClusterId": "glo-1", "clusterId": "in01-secondary-ap", "prompt": "deleted"}}), nil
		case 3:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe while secondary is deleting %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayloadWithSecondaryAPDeleting(4)), nil
		case 4:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe after secondary deleted %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayloadWithoutSecondaryAP(4)), nil
		case 5:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("final describe after update %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayloadWithoutSecondaryAP(4)), nil
		default:
			return nil, fmt.Errorf("unexpected call %d", call)
		}
	})
	schema := testGlobalClusterResourceSchema(t, resource)
	base := testGlobalClusterBaseModel()
	state := testGlobalClusterState(t, ctx, schema, base)
	base.Cluster = base.Cluster[:2]
	plan := testGlobalClusterPlan(t, ctx, schema, base)

	var resp fwresource.UpdateResponse
	resp.State = tfsdk.State{Schema: schema}
	resource.Update(ctx, fwresource.UpdateRequest{Plan: plan, State: state}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Update diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}

	var got GlobalClusterResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("State.Get diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}
	if len(got.Cluster) != 2 {
		t.Fatalf("deleted secondary was written back to state: %+v", got.Cluster)
	}
	if got.Cluster[1].ClusterID.ValueString() != "in01-secondary" {
		t.Fatalf("unexpected remaining secondary: %+v", got.Cluster[1])
	}
}

func TestGlobalClusterResourceUpdateRejectsSecondaryModification(t *testing.T) {
	ctx := context.Background()
	resource := newTestGlobalClusterResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		if call != 1 || req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
			t.Fatalf("describe before update %d %s %s", call, req.Method, req.URL.Path)
		}
		return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayload(4)), nil
	})
	schema := testGlobalClusterResourceSchema(t, resource)
	base := testGlobalClusterBaseModel()
	state := testGlobalClusterState(t, ctx, schema, base)
	base.Cluster[1] = GlobalClusterMemberModel{ClusterID: types.StringUnknown(), ClusterName: types.StringValue("secondary-eu-renamed"), RegionID: types.StringValue("aws-eu-west-1"), Role: types.StringUnknown(), Status: types.StringUnknown()}
	plan := testGlobalClusterPlan(t, ctx, schema, base)

	var resp fwresource.UpdateResponse
	resp.State = tfsdk.State{Schema: schema}
	resource.Update(ctx, fwresource.UpdateRequest{Plan: plan, State: state}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected secondary modification diagnostic")
	}
}

func TestGlobalClusterResourceDeleteRemovesSecondariesBeforePrimary(t *testing.T) {
	ctx := context.Background()
	testGlobalClusterSecondaryDeleteWait(t, time.Millisecond, time.Second)
	resource := newTestGlobalClusterResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		switch call {
		case 1:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayload(4)), nil
		case 2:
			if req.Method != http.MethodDelete || req.URL.Path != "/v2/globalClusters/glo-1/clusters/in01-secondary" {
				t.Fatalf("delete first secondary %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, map[string]any{
				"code": 0,
				"data": map[string]any{"globalClusterId": "glo-1", "clusterId": "in01-secondary", "prompt": "deleted"},
			}), nil
		case 3:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe after first secondary deleted %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayloadWithoutSecondaryEU(4)), nil
		case 4:
			if req.Method != http.MethodDelete || req.URL.Path != "/v2/globalClusters/glo-1/clusters/in01-secondary-ap" {
				t.Fatalf("delete second secondary %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, map[string]any{
				"code": 0,
				"data": map[string]any{"globalClusterId": "glo-1", "clusterId": "in01-secondary-ap", "prompt": "deleted"},
			}), nil
		case 5:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe after second secondary deleted %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayloadWithoutSecondaries()), nil
		case 6:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe primary deletable %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayloadWithoutSecondaries()), nil
		case 7:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("final describe before primary delete %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayloadWithoutSecondaries()), nil
		case 8:
			if req.Method != http.MethodDelete || req.URL.Path != "/v2/globalClusters/glo-1/clusters/in01-primary" {
				t.Fatalf("delete primary %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, map[string]any{
				"code": 0,
				"data": map[string]any{"globalClusterId": "glo-1", "clusterId": "in01-primary", "prompt": "deleted"},
			}), nil
		default:
			return nil, fmt.Errorf("unexpected call %d", call)
		}
	})
	schema := testGlobalClusterResourceSchema(t, resource)
	state := testGlobalClusterState(t, ctx, schema, testGlobalClusterBaseModel())

	var resp fwresource.DeleteResponse
	resource.Delete(ctx, fwresource.DeleteRequest{State: state}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Delete diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}
}

func TestGlobalClusterResourceDeleteWaitsForPrimaryMemberRunningBeforePrimaryDelete(t *testing.T) {
	ctx := context.Background()
	testGlobalClusterSecondaryDeleteWait(t, time.Millisecond, time.Second)
	resource := newTestGlobalClusterResource(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		switch call {
		case 1:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe %s %s", req.Method, req.URL.Path)
			}
			payload := describeGlobalClusterPayloadWithLockedPrimary(t)
			return globalClusterJSONResponse(t, http.StatusOK, payload), nil
		case 2:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe primary deletable %s %s", req.Method, req.URL.Path)
			}
			payload := describeGlobalClusterPayloadWithLockedPrimary(t)
			return globalClusterJSONResponse(t, http.StatusOK, payload), nil
		case 3:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("describe primary running %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayloadWithoutSecondaries()), nil
		case 4:
			if req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
				t.Fatalf("final describe before primary delete %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, describeGlobalClusterPayloadWithoutSecondaries()), nil
		case 5:
			if req.Method != http.MethodDelete || req.URL.Path != "/v2/globalClusters/glo-1/clusters/in01-primary" {
				t.Fatalf("delete primary %s %s", req.Method, req.URL.Path)
			}
			return globalClusterJSONResponse(t, http.StatusOK, map[string]any{
				"code": 0,
				"data": map[string]any{"globalClusterId": "glo-1", "clusterId": "in01-primary", "prompt": "deleted"},
			}), nil
		default:
			return nil, fmt.Errorf("unexpected call %d", call)
		}
	})
	schema := testGlobalClusterResourceSchema(t, resource)
	stateModel := testGlobalClusterBaseModel()
	stateModel.Cluster = stateModel.Cluster[:1]
	state := testGlobalClusterState(t, ctx, schema, stateModel)

	var resp fwresource.DeleteResponse
	resource.Delete(ctx, fwresource.DeleteRequest{State: state}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Delete diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}
}
