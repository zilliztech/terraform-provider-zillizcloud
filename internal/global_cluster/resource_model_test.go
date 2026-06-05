package global_cluster

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGlobalClusterMembersFromDomainFillsComputedFieldsAndPreservesSecondaryOrder(t *testing.T) {
	current := []GlobalClusterMemberModel{
		{ClusterName: types.StringValue("secondary-eu"), RegionID: types.StringValue("aws-eu-west-1")},
		{ClusterName: types.StringValue("secondary-ap"), RegionID: types.StringValue("aws-ap-southeast-1")},
		{ClusterName: types.StringValue("primary-a"), RegionID: types.StringValue("aws-us-west-2")},
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
