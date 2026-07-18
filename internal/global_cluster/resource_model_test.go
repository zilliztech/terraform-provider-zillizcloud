package global_cluster

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGlobalClusterMembersFromDomainFillsComputedFieldsAndPreservesSecondaryOrder(t *testing.T) {
	current := []GlobalClusterMemberModel{
		{ClusterName: types.StringValue("secondary-eu"), RegionID: types.StringValue("aws-eu-west-1"), Replica: types.Int64Value(2)},
		{ClusterName: types.StringValue("secondary-ap"), RegionID: types.StringValue("aws-ap-southeast-1"), Replica: types.Int64Value(1)},
		{ClusterName: types.StringValue("primary-a"), RegionID: types.StringValue("aws-us-west-2"), Replica: types.Int64Value(3)},
	}
	apiMembers := []GlobalClusterMember{
		{ClusterID: "in01-primary", ClusterName: "primary-a", RegionID: "aws-us-west-2", Role: GlobalClusterMemberRolePrimary, Status: "RUNNING"},
		{ClusterID: "in01-secondary-ap", ClusterName: "secondary-ap", RegionID: "aws-ap-southeast-1", Role: GlobalClusterMemberRoleSecondary, Status: "RUNNING"},
		{ClusterID: "in01-secondary", ClusterName: "secondary-eu", RegionID: "aws-eu-west-1", Role: GlobalClusterMemberRoleSecondary, Status: "RUNNING"},
		{ClusterID: "in01-secondary-au", ClusterName: "secondary-au", RegionID: "aws-ap-southeast-2", Role: GlobalClusterMemberRoleSecondary, Status: "RUNNING"},
	}

	got := globalClusterMembersFromDomain(apiMembers, current)
	if len(got) != 4 {
		t.Fatalf("unexpected length=%d members=%+v", len(got), got)
	}
	if got[0].ClusterID.ValueString() != "in01-primary" || got[0].Role.ValueString() != "PRIMARY" || got[0].Status.ValueString() != "RUNNING" {
		t.Fatalf("primary was not kept first and hydrated: %+v", got[0])
	}
	if got[0].Replica.ValueInt64() != 3 || got[1].Replica.ValueInt64() != 2 || got[2].Replica.ValueInt64() != 1 || got[3].Replica.ValueInt64() != 1 {
		t.Fatalf("member replicas were not preserved/defaulted: %+v", got)
	}
	if got[1].ClusterID.ValueString() != "in01-secondary" || got[1].Role.ValueString() != "SECONDARY" {
		t.Fatalf("first secondary did not follow configured order: %+v", got[1])
	}
	if got[2].ClusterID.ValueString() != "in01-secondary-ap" || got[2].Role.ValueString() != "SECONDARY" {
		t.Fatalf("second secondary did not follow configured order: %+v", got[2])
	}
	if got[3].ClusterName.ValueString() != "secondary-au" || got[3].ClusterID.ValueString() != "in01-secondary-au" {
		t.Fatalf("api-only secondary was not appended: %+v", got[3])
	}
}

func TestApplyGlobalClusterPopulatesAutoscalingAndMemberReplicas(t *testing.T) {
	ctx := context.Background()
	model := GlobalClusterResourceModel{
		CUSize: types.Int64Unknown(),
		Cluster: []GlobalClusterMemberModel{
			{ClusterName: types.StringValue("primary-a"), RegionID: types.StringValue("aws-us-west-2"), Replica: types.Int64Unknown()},
			{ClusterName: types.StringValue("secondary-eu"), RegionID: types.StringValue("aws-eu-west-1"), Replica: types.Int64Unknown()},
		},
	}
	cuMin, cuMax := int64(4), int64(8)
	replicaMin, replicaMax := int64(1), int64(3)
	var diagnostics diag.Diagnostics
	model.applyGlobalCluster(ctx, &GlobalCluster{
		GlobalClusterID:   "glo-1",
		GlobalClusterName: "global-a",
		ProjectID:         "proj-1",
		RegionIDs:         []string{"aws-us-west-2", "aws-eu-west-1"},
		CUType:            "Performance-optimized",
		Autoscaling: GlobalClusterAutoscaling{
			CU:      &GlobalClusterAutoscalingPolicy{Min: &cuMin, Max: &cuMax},
			Replica: &GlobalClusterAutoscalingPolicy{Min: &replicaMin, Max: &replicaMax},
		},
		Clusters: []GlobalClusterMember{
			{ClusterID: "in01-primary", ClusterName: "primary-a", RegionID: "aws-us-west-2", Role: GlobalClusterMemberRolePrimary, Status: "RUNNING", Replica: 2},
			{ClusterID: "in01-secondary", ClusterName: "secondary-eu", RegionID: "aws-eu-west-1", Role: GlobalClusterMemberRoleSecondary, Status: "RUNNING", Replica: 1},
		},
	}, diagnostics.Append)

	if diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diagnostics.Errors()[0].Summary())
	}
	if !model.CUSize.IsNull() {
		t.Fatalf("cu_size must be null under CU autoscaling, got %s", model.CUSize.String())
	}
	if model.CUSettings == nil || model.CUSettings.DynamicScaling.Min.ValueInt64() != 4 || model.CUSettings.DynamicScaling.Max.ValueInt64() != 8 {
		t.Fatalf("unexpected CU settings: %+v", model.CUSettings)
	}
	if model.ReplicaSettings == nil || model.ReplicaSettings.DynamicScaling.Min.ValueInt64() != 1 || model.ReplicaSettings.DynamicScaling.Max.ValueInt64() != 3 {
		t.Fatalf("unexpected replica settings: %+v", model.ReplicaSettings)
	}
	if model.Cluster[0].Replica.ValueInt64() != 2 || model.Cluster[1].Replica.ValueInt64() != 1 {
		t.Fatalf("unexpected member replicas: %+v", model.Cluster)
	}
}

func TestGlobalClusterResourceModelBuildsCreateAutoscaling(t *testing.T) {
	model := GlobalClusterResourceModel{
		CUSettings: &GlobalClusterScalingModel{DynamicScaling: &GlobalClusterDynamicScalingModel{
			Min: types.Int64Value(4),
			Max: types.Int64Value(8),
		}},
		ReplicaSettings: &GlobalClusterScalingModel{DynamicScaling: &GlobalClusterDynamicScalingModel{
			Min: types.Int64Value(1),
			Max: types.Int64Value(3),
		}},
		Cluster: []GlobalClusterMemberModel{
			{ClusterName: types.StringValue("primary-a"), RegionID: types.StringValue("aws-us-west-2"), Replica: types.Int64Value(2)},
			{ClusterName: types.StringValue("secondary-eu"), RegionID: types.StringValue("aws-eu-west-1"), Replica: types.Int64Value(1)},
		},
	}
	autoscaling := model.autoscaling()
	if autoscaling.CU == nil || *autoscaling.CU.Min != 4 || *autoscaling.CU.Max != 8 || autoscaling.Replica == nil || *autoscaling.Replica.Max != 3 {
		t.Fatalf("unexpected autoscaling: %+v", autoscaling)
	}
	members := model.memberSpecs()
	if len(members) != 2 || members[0].Replica == nil || *members[0].Replica != 2 || members[1].Replica == nil || *members[1].Replica != 1 {
		t.Fatalf("unexpected member specs: %+v", members)
	}
}

func TestApplyGlobalClusterPopulatesComputedFields(t *testing.T) {
	ctx := context.Background()
	model := GlobalClusterResourceModel{}
	var diagnostics diag.Diagnostics
	globalCluster := &GlobalCluster{
		GlobalClusterID:   "glo-1",
		GlobalClusterName: "global-a",
		ProjectID:         "proj-1",
		RegionIDs:         []string{"aws-us-west-2"},
		CUType:            "Performance-optimized",
		CUSize:            4,
		ConnectAddress:    "https://glo-1.global-cluster.vectordb.zillizcloud.com",
		CreateTime:        "2026-06-04T10:00:00Z",
		Clusters: []GlobalClusterMember{
			{ClusterID: "in01-primary", ClusterName: "primary-a", RegionID: "aws-us-west-2", Role: GlobalClusterMemberRolePrimary, Status: "RUNNING"},
		},
	}
	model.applyGlobalCluster(ctx, globalCluster, diagnostics.Append)

	if diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diagnostics.Errors()[0].Summary())
	}
	if model.ID.ValueString() != "glo-1" || model.GlobalClusterName.ValueString() != "global-a" {
		t.Fatalf("identity not populated: %+v", model)
	}
	if len(model.Cluster) != 1 || model.Cluster[0].ClusterID.ValueString() != "in01-primary" || model.Cluster[0].Role.ValueString() != "PRIMARY" {
		t.Fatalf("member computed fields not populated: %+v", model.Cluster)
	}
}
