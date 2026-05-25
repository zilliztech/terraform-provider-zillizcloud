package cluster

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	fwtimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func TestClusterResourceSchemaReplicaHasNoStaticDefault(t *testing.T) {
	_, schema := testClusterResourceWithSchema(t, nil, nil)

	replicaAttr, ok := schema.Attributes["replica"].(fwschema.Int64Attribute)
	if !ok {
		t.Fatalf("replica attribute = %T, want schema.Int64Attribute", schema.Attributes["replica"])
	}
	if !replicaAttr.Optional || !replicaAttr.Computed {
		t.Fatalf("replica Optional=%t Computed=%t, want both true", replicaAttr.Optional, replicaAttr.Computed)
	}
	if replicaAttr.Default != nil {
		t.Fatal("replica should not have a schema-level default")
	}
}

func TestClusterResourceModifyPlanDefaultsReplicaOnlyForFixedMode(t *testing.T) {
	ctx := context.Background()
	r, schema := testClusterResourceWithSchema(t, nil, nil)

	configModel := testClusterModel(0, nil)
	configModel.Replica = types.Int64Null()
	planModel := configModel
	planModel.Replica = types.Int64Unknown()

	plan := testClusterPlan(t, ctx, schema, planModel)
	resp := fwresource.ModifyPlanResponse{Plan: plan}
	r.ModifyPlan(ctx, fwresource.ModifyPlanRequest{
		Config: testClusterConfig(t, ctx, schema, configModel),
		State:  tfsdk.State{Schema: schema},
		Plan:   plan,
	}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("modify plan diagnostics: %s", resp.Diagnostics)
	}

	var got ClusterResourceModel
	diags := resp.Plan.Get(ctx, &got)
	if diags.HasError() {
		t.Fatalf("plan get diagnostics: %s", diags)
	}
	if got.Replica.ValueInt64() != 1 {
		t.Fatalf("replica = %d, want 1", got.Replica.ValueInt64())
	}
}

func TestClusterResourceModifyPlanPreservesLiveReplicaForDynamicMode(t *testing.T) {
	ctx := context.Background()
	r, schema := testClusterResourceWithSchema(t, nil, nil)
	replicaSettings := testReplicaSettings(1, 2)

	configModel := testClusterModel(0, replicaSettings)
	configModel.Replica = types.Int64Null()
	planModel := testClusterModel(1, replicaSettings)
	stateModel := testClusterModel(2, replicaSettings)

	plan := testClusterPlan(t, ctx, schema, planModel)
	resp := fwresource.ModifyPlanResponse{Plan: plan}
	r.ModifyPlan(ctx, fwresource.ModifyPlanRequest{
		Config: testClusterConfig(t, ctx, schema, configModel),
		State:  testClusterState(t, ctx, schema, stateModel),
		Plan:   plan,
	}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("modify plan diagnostics: %s", resp.Diagnostics)
	}

	var got ClusterResourceModel
	diags := resp.Plan.Get(ctx, &got)
	if diags.HasError() {
		t.Fatalf("plan get diagnostics: %s", diags)
	}
	if got.Replica.ValueInt64() != 2 {
		t.Fatalf("replica = %d, want live state replica 2", got.Replica.ValueInt64())
	}
}

func TestClusterResourceUpdateDoesNotModifyFixedReplicaInDynamicMode(t *testing.T) {
	ctx := context.Background()
	store := &fakeClusterStore{current: testClusterModel(2, testReplicaSettings(1, 3))}
	r, schema := testClusterResourceWithSchema(t, store, nil)

	state := testClusterState(t, ctx, schema, testClusterModel(2, testReplicaSettings(1, 2)))
	plan := testClusterPlan(t, ctx, schema, testClusterModel(1, testReplicaSettings(1, 3)))

	resp := fwresource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(ctx, fwresource.UpdateRequest{State: state, Plan: plan}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("update diagnostics: %s", resp.Diagnostics)
	}
	if len(store.modifyReplicaCalls) != 0 {
		t.Fatalf("ModifyReplica calls = %#v, want none", store.modifyReplicaCalls)
	}
	if store.modifyAutoscalingCalls != 1 {
		t.Fatalf("ModifyAutoscaling calls = %d, want 1", store.modifyAutoscalingCalls)
	}

	var got ClusterResourceModel
	diags := resp.State.Get(ctx, &got)
	if diags.HasError() {
		t.Fatalf("state get diagnostics: %s", diags)
	}
	if got.Replica.ValueInt64() != 2 {
		t.Fatalf("replica = %d, want live replica 2", got.Replica.ValueInt64())
	}
}

func TestClusterResourceCreateDynamicModeUsesLiveReplicaState(t *testing.T) {
	ctx := context.Background()
	store := &fakeClusterStore{current: testClusterModel(2, testReplicaSettings(1, 2))}
	r, schema := testClusterResourceWithSchema(t, store, nil)

	planModel := testClusterModel(0, testReplicaSettings(1, 2))
	planModel.ClusterId = types.StringNull()
	planModel.Replica = types.Int64Unknown()
	plan := testClusterPlan(t, ctx, schema, planModel)

	resp := fwresource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(ctx, fwresource.CreateRequest{Plan: plan}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("create diagnostics: %s", resp.Diagnostics)
	}
	if len(store.modifyReplicaCalls) != 0 {
		t.Fatalf("ModifyReplica calls = %#v, want none", store.modifyReplicaCalls)
	}
	if store.modifyAutoscalingCalls != 1 {
		t.Fatalf("ModifyAutoscaling calls = %d, want 1", store.modifyAutoscalingCalls)
	}

	var got ClusterResourceModel
	diags := resp.State.Get(ctx, &got)
	if diags.HasError() {
		t.Fatalf("state get diagnostics: %s", diags)
	}
	if got.Replica.ValueInt64() != 2 {
		t.Fatalf("replica = %d, want live replica 2", got.Replica.ValueInt64())
	}
}

func TestClusterResourceUpdateModifiesExplicitFixedReplica(t *testing.T) {
	ctx := context.Background()
	store := &fakeClusterStore{current: testClusterModel(2, nil)}
	client := testClusterClient(t)
	r, schema := testClusterResourceWithSchema(t, store, client)

	state := testClusterState(t, ctx, schema, testClusterModel(1, nil))
	plan := testClusterPlan(t, ctx, schema, testClusterModel(2, nil))

	resp := fwresource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(ctx, fwresource.UpdateRequest{State: state, Plan: plan}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("update diagnostics: %s", resp.Diagnostics)
	}
	if len(store.modifyReplicaCalls) != 1 || store.modifyReplicaCalls[0] != 2 {
		t.Fatalf("ModifyReplica calls = %#v, want [2]", store.modifyReplicaCalls)
	}
}

func testClusterResourceWithSchema(t *testing.T, store ClusterStore, client *zilliz.Client) (*ClusterResource, fwschema.Schema) {
	t.Helper()

	r := &ClusterResource{store: store, client: client}
	var schemaResp fwresource.SchemaResponse
	r.Schema(context.Background(), fwresource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %s", schemaResp.Diagnostics)
	}
	return r, schemaResp.Schema
}

func testClusterModel(replica int64, replicaSettings *ReplicaSettings) ClusterResourceModel {
	return ClusterResourceModel{
		ClusterId:          types.StringValue("cluster-1"),
		Plan:               types.StringValue(EnterprisePlan),
		ClusterName:        types.StringValue("test-cluster"),
		CuSize:             types.Int64Value(8),
		CuType:             types.StringValue("Performance-optimized"),
		ProjectId:          types.StringValue("project-1"),
		Username:           types.StringUnknown(),
		Password:           types.StringUnknown(),
		Prompt:             types.StringUnknown(),
		Description:        types.StringValue(""),
		RegionId:           types.StringValue("aws-us-west-2"),
		Status:             types.StringValue("RUNNING"),
		DesiredStatus:      types.StringValue("RUNNING"),
		ConnectAddress:     types.StringValue("public.example.com"),
		PrivateLinkAddress: types.StringValue("private.example.com"),
		CreateTime:         types.StringValue("2026-01-01T00:00:00Z"),
		Labels:             types.MapValueMust(types.StringType, map[string]attr.Value{}),
		SecurityGroups:     types.SetNull(types.StringType),
		Replica:            types.Int64Value(replica),
		ReplicaSettings:    replicaSettings,
		AwsCseKeyArn:       types.StringNull(),
		Timeouts: fwtimeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"update": types.StringType,
		})},
	}
}

func testReplicaSettings(min, max int64) *ReplicaSettings {
	return &ReplicaSettings{
		DynamicScaling: &DynamicScaling{
			Min: types.Int64Value(min),
			Max: types.Int64Value(max),
		},
	}
}

func testClusterConfig(t *testing.T, ctx context.Context, schema fwschema.Schema, model ClusterResourceModel) tfsdk.Config {
	t.Helper()
	plan := testClusterPlan(t, ctx, schema, model)
	return tfsdk.Config{Schema: schema, Raw: plan.Raw}
}

func testClusterPlan(t *testing.T, ctx context.Context, schema fwschema.Schema, model ClusterResourceModel) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: schema}
	diags := plan.Set(ctx, &model)
	if diags.HasError() {
		t.Fatalf("plan set diagnostics: %s", diags)
	}
	return plan
}

func testClusterState(t *testing.T, ctx context.Context, schema fwschema.Schema, model ClusterResourceModel) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: schema}
	diags := state.Set(ctx, &model)
	if diags.HasError() {
		t.Fatalf("state set diagnostics: %s", diags)
	}
	return state
}

func testClusterClient(t *testing.T) *zilliz.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/clusters/cluster-1" {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": zilliz.Cluster{
				ClusterId:   "cluster-1",
				ClusterName: "test-cluster",
				Status:      "RUNNING",
			},
		}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client, err := zilliz.NewClient(
		zilliz.WithApiKey("test-api-key"),
		zilliz.WithBaseUrl(server.URL),
	)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	return client
}

type fakeClusterStore struct {
	current                ClusterResourceModel
	modifyReplicaCalls     []int
	modifyAutoscalingCalls int
}

func (s *fakeClusterStore) Get(ctx context.Context, clusterId string) (*ClusterResourceModel, error) {
	return &s.current, nil
}

func (s *fakeClusterStore) GetLabels(ctx context.Context, clusterId string) (types.Map, error) {
	return types.MapValueMust(types.StringType, map[string]attr.Value{}), nil
}

func (s *fakeClusterStore) Create(ctx context.Context, cluster *ClusterResourceModel) (*ClusterResourceModel, error) {
	return &ClusterResourceModel{ClusterId: types.StringValue("cluster-1")}, nil
}

func (s *fakeClusterStore) Delete(ctx context.Context, clusterId string) error {
	return nil
}

func (s *fakeClusterStore) UpgradeCuSize(ctx context.Context, clusterId string, cuSize int) error {
	s.current.CuSize = types.Int64Value(int64(cuSize))
	return nil
}

func (s *fakeClusterStore) ModifyReplica(ctx context.Context, clusterId string, replica int) error {
	s.modifyReplicaCalls = append(s.modifyReplicaCalls, replica)
	s.current.Replica = types.Int64Value(int64(replica))
	return nil
}

func (s *fakeClusterStore) SuspendCluster(ctx context.Context, clusterId string) error {
	s.current.Status = types.StringValue("SUSPENDED")
	return nil
}

func (s *fakeClusterStore) ResumeCluster(ctx context.Context, clusterId string) error {
	s.current.Status = types.StringValue("RUNNING")
	return nil
}

func (s *fakeClusterStore) UpdateLabels(ctx context.Context, clusterId string, labels map[string]string) error {
	return nil
}

func (s *fakeClusterStore) ModifyClusterProperties(ctx context.Context, clusterId string, clusterName string) error {
	s.current.ClusterName = types.StringValue(clusterName)
	return nil
}

func (s *fakeClusterStore) UpsertSecurityGroups(ctx context.Context, clusterId string, securityGroupIds []string) error {
	return nil
}

func (s *fakeClusterStore) GetSecurityGroups(ctx context.Context, clusterId string) ([]string, error) {
	return nil, nil
}

func (s *fakeClusterStore) ModifyAutoscaling(ctx context.Context, clusterId string, params *zilliz.ModifyAutoscalingCombinedParams) error {
	s.modifyAutoscalingCalls++
	return nil
}
