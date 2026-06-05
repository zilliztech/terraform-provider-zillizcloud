package global_cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GlobalClusterResourceModel struct {
	ID                types.String               `tfsdk:"id"`
	GlobalClusterName types.String               `tfsdk:"global_cluster_name"`
	ProjectID         types.String               `tfsdk:"project_id"`
	CUType            types.String               `tfsdk:"cu_type"`
	CUSize            types.Int64                `tfsdk:"cu_size"`
	Cluster           []GlobalClusterMemberModel `tfsdk:"cluster"`
	ConnectAddress    types.String               `tfsdk:"connect_address"`
	CreateTime        types.String               `tfsdk:"create_time"`
	RegionIDs         types.List                 `tfsdk:"region_ids"`
	Username          types.String               `tfsdk:"username"`
	Password          types.String               `tfsdk:"password"`
	CreateJobID       types.String               `tfsdk:"create_job_id"`
}

type GlobalClusterMemberModel struct {
	ClusterID   types.String `tfsdk:"cluster_id"`
	ClusterName types.String `tfsdk:"cluster_name"`
	RegionID    types.String `tfsdk:"region_id"`
	Role        types.String `tfsdk:"role"`
	Status      types.String `tfsdk:"status"`
}

func (data *GlobalClusterResourceModel) applyGlobalCluster(ctx context.Context, globalCluster *GlobalCluster, appendDiags func(...diag.Diagnostic)) {
	if globalCluster == nil {
		return
	}

	data.ID = types.StringValue(globalCluster.GlobalClusterID)
	data.GlobalClusterName = types.StringValue(globalCluster.GlobalClusterName)
	data.ProjectID = types.StringValue(globalCluster.ProjectID)
	data.CUType = types.StringValue(globalCluster.CUType)
	data.CUSize = types.Int64Value(globalCluster.CUSize)
	data.ConnectAddress = types.StringValue(globalCluster.ConnectAddress)
	data.CreateTime = types.StringValue(globalCluster.CreateTime)

	regionIDs, regionDiags := types.ListValueFrom(ctx, types.StringType, globalCluster.RegionIDs)
	appendDiags(regionDiags...)
	if !regionDiags.HasError() {
		data.RegionIDs = regionIDs
	}

	data.Cluster = globalClusterMembersFromDomain(globalCluster.Clusters, data.Cluster)
}

func (data GlobalClusterResourceModel) memberSpecs() []GlobalClusterMemberSpec {
	specs := make([]GlobalClusterMemberSpec, 0, len(data.Cluster))
	for _, member := range data.Cluster {
		specs = append(specs, GlobalClusterMemberSpec{ClusterName: member.ClusterName.ValueString(), RegionID: member.RegionID.ValueString()})
	}
	return specs
}

func globalClusterMembersFromDomain(apiMembers []GlobalClusterMember, current []GlobalClusterMemberModel) []GlobalClusterMemberModel {
	primaryMembers, secondaryMembers := splitGlobalClusterMembers(apiMembers)
	if len(primaryMembers) == 0 && len(secondaryMembers) == 0 {
		return current
	}

	result := make([]GlobalClusterMemberModel, 0, len(primaryMembers)+len(secondaryMembers))
	for _, primary := range primaryMembers {
		result = append(result, globalClusterMemberModel(primary))
	}

	usedSecondaries := make([]bool, len(secondaryMembers))
	for _, configuredMember := range current {
		matchedIndex := findGlobalClusterMemberIndex(secondaryMembers, usedSecondaries, configuredMember)
		if matchedIndex == -1 {
			continue
		}
		usedSecondaries[matchedIndex] = true
		result = append(result, globalClusterMemberModel(secondaryMembers[matchedIndex]))
	}

	for i, member := range secondaryMembers {
		if !usedSecondaries[i] {
			result = append(result, globalClusterMemberModel(member))
		}
	}
	return result
}

func splitGlobalClusterMembers(members []GlobalClusterMember) ([]GlobalClusterMember, []GlobalClusterMember) {
	primaryMembers := make([]GlobalClusterMember, 0, 1)
	secondaryMembers := make([]GlobalClusterMember, 0)
	for _, member := range members {
		switch member.Role {
		case GlobalClusterMemberRolePrimary:
			primaryMembers = append(primaryMembers, member)
		case GlobalClusterMemberRoleSecondary:
			secondaryMembers = append(secondaryMembers, member)
		}
	}
	return primaryMembers, secondaryMembers
}

func findGlobalClusterMemberIndex(members []GlobalClusterMember, used []bool, configuredMember GlobalClusterMemberModel) int {
	for i, member := range members {
		if used[i] {
			continue
		}
		if member.ClusterName == configuredMember.ClusterName.ValueString() && member.RegionID == configuredMember.RegionID.ValueString() {
			return i
		}
	}
	return -1
}

func globalClusterMemberModel(member GlobalClusterMember) GlobalClusterMemberModel {
	return GlobalClusterMemberModel{
		ClusterID:   types.StringValue(member.ClusterID),
		ClusterName: types.StringValue(member.ClusterName),
		RegionID:    types.StringValue(member.RegionID),
		Role:        types.StringValue(string(member.Role)),
		Status:      types.StringValue(member.Status),
	}
}
