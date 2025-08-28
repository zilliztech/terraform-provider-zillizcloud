package cluster

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

type StatusAction int

const (
	StatusActionNone StatusAction = iota
	StatusActionSuspend
	StatusActionResume
)

type ClusterResourceModel struct {
	ClusterId          types.String   `tfsdk:"id"`
	Plan               types.String   `tfsdk:"plan"`
	ClusterName        types.String   `tfsdk:"cluster_name"`
	CuSize             types.Int64    `tfsdk:"cu_size"`
	CuType             types.String   `tfsdk:"cu_type"`
	ProjectId          types.String   `tfsdk:"project_id"`
	Username           types.String   `tfsdk:"username"`
	Password           types.String   `tfsdk:"password"`
	Prompt             types.String   `tfsdk:"prompt"`
	Description        types.String   `tfsdk:"description"`
	RegionId           types.String   `tfsdk:"region_id"`
	Status             types.String   `tfsdk:"status"`
	DesiredStatus      types.String   `tfsdk:"desired_status"`
	ConnectAddress     types.String   `tfsdk:"connect_address"`
	PrivateLinkAddress types.String   `tfsdk:"private_link_address"`
	CreateTime         types.String   `tfsdk:"create_time"`
	Labels             types.Map      `tfsdk:"labels"`
	Replica            types.Int64    `tfsdk:"replica"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

// populate the ClusterResourceModel with the input which is the response from the API.
func (c *ClusterResourceModel) populate(input *ClusterResourceModel) {

	c.ClusterId = input.ClusterId
	c.ClusterName = input.ClusterName
	c.ProjectId = input.ProjectId
	c.Description = input.Description
	c.Status = input.Status
	c.DesiredStatus = input.Status
	c.ConnectAddress = input.ConnectAddress
	c.PrivateLinkAddress = input.PrivateLinkAddress
	c.CreateTime = input.CreateTime
	c.Replica = input.Replica
	c.CuSize = input.CuSize
	c.CuType = input.CuType

	// only for free or serverless plan, set default value
	plan := input.Plan.ValueString()
	isFreeOrServerless := plan == string(zilliz.FreePlan) || plan == string(zilliz.ServerlessPlan)
	if isFreeOrServerless {
		c.CuSize = types.Int64Value(1)
		c.CuType = types.StringValue("Performance-optimized")
		c.Replica = types.Int64Value(1)
	}

}

// Comparison methods for ClusterResourceModel
func (c *ClusterResourceModel) isCuSizeChanged(other ClusterResourceModel) bool {
	return c.CuSize.ValueInt64() != other.CuSize.ValueInt64()
}

func (c *ClusterResourceModel) isClusterPlanChanged(other ClusterResourceModel) bool {
	return c.Plan.ValueString() != other.Plan.ValueString()
}

func (c *ClusterResourceModel) isReplicaChanged(other ClusterResourceModel) bool {
	return c.Replica.ValueInt64() != other.Replica.ValueInt64()
}

func (c *ClusterResourceModel) isLabelsChanged(other ClusterResourceModel) bool {
	return !c.Labels.Equal(other.Labels)
}

func (c *ClusterResourceModel) isStatusChangeRequired(other ClusterResourceModel) bool {
	if !c.DesiredStatus.IsNull() && c.DesiredStatus.ValueString() != "" {
		desiredStatus := c.DesiredStatus.ValueString()
		currentStatus := other.Status.ValueString()

		if (desiredStatus == "SUSPENDED" && currentStatus == "RUNNING") ||
			(desiredStatus == "RUNNING" && currentStatus == "SUSPENDED") {
			return true
		}
	}

	if c.DesiredStatus.IsNull() && other.Status.ValueString() == "SUSPENDED" {
		return true
	}

	return false
}

func (c *ClusterResourceModel) getStatusAction(other ClusterResourceModel) StatusAction {
	if !c.DesiredStatus.IsNull() && c.DesiredStatus.ValueString() != "" {
		desiredStatus := c.DesiredStatus.ValueString()
		currentStatus := other.Status.ValueString()

		if desiredStatus == "SUSPENDED" && currentStatus == "RUNNING" {
			return StatusActionSuspend
		}
		if desiredStatus == "RUNNING" && currentStatus == "SUSPENDED" {
			return StatusActionResume
		}
	}

	if c.DesiredStatus.IsNull() && other.Status.ValueString() == "SUSPENDED" {
		return StatusActionResume
	}

	return StatusActionNone
}
