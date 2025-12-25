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

type DynamicScaling struct {
	Min types.Int64 `tfsdk:"min"`
	Max types.Int64 `tfsdk:"max"`
}

func (d *DynamicScaling) Equal(other *DynamicScaling) bool {
	if d == nil && other == nil {
		return true
	}
	if d == nil || other == nil {
		return false
	}
	return d.Min.Equal(other.Min) && d.Max.Equal(other.Max)
}

type CuSettings struct {
	DynamicScaling *DynamicScaling `tfsdk:"dynamic_scaling"`
}

func (c *CuSettings) IsdynamicScalingNull() bool {
	return c.DynamicScaling == nil || c.DynamicScaling.Min.IsNull() || c.DynamicScaling.Max.IsNull()
}

func (c *CuSettings) Equal(other *CuSettings) bool {
	if c == nil && other == nil {
		return true
	}
	if c == nil || other == nil {
		return false
	}
	return c.DynamicScaling.Equal(other.DynamicScaling)
}

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
	SecurityGroups     types.Set      `tfsdk:"load_balancer_security_groups"`
	Replica            types.Int64    `tfsdk:"replica"`
	CuSettings         *CuSettings    `tfsdk:"cu_settings"`
	BucketInfo         *BucketInfo    `tfsdk:"bucket_info"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

type BucketInfo struct {
	BucketName types.String `tfsdk:"bucket_name"`
	Prefix     types.String `tfsdk:"prefix"`
}

func (b *BucketInfo) Equal(other *BucketInfo) bool {
	if b == nil && other == nil {
		return true
	}
	if b == nil || other == nil {
		return false
	}
	return b.BucketName.Equal(other.BucketName) && b.Prefix.Equal(other.Prefix)
}

func (c *ClusterResourceModel) isCuSettingsDisabled() bool {
	return c.CuSettings == nil || c.CuSettings.DynamicScaling == nil
}

func (c *ClusterResourceModel) setUnknown() {
	unknown := types.StringValue("UNKNOWN")

	c.ConnectAddress = unknown
	c.PrivateLinkAddress = unknown
	c.CreateTime = unknown
	c.Status = unknown
	c.Description = unknown
	if c.RegionId.IsNull() {
		c.RegionId = unknown
	}
	c.SecurityGroups = types.SetNull(types.StringType)
}

// populate the ClusterResourceModel with the input which is the response from the API.
func (c *ClusterResourceModel) populate(input *ClusterResourceModel) {

	c.ClusterId = input.ClusterId
	c.ClusterName = input.ClusterName
	c.ProjectId = input.ProjectId
	if !input.RegionId.IsNull() && input.RegionId.ValueString() != "" {
		c.RegionId = input.RegionId
	}
	c.Description = input.Description
	c.Status = input.Status
	c.DesiredStatus = input.Status
	c.ConnectAddress = input.ConnectAddress
	c.PrivateLinkAddress = input.PrivateLinkAddress
	c.CreateTime = input.CreateTime
	c.Plan = input.Plan
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

// only for free or serverless plan, set default value
func (c *ClusterResourceModel) completeForFreeOrServerless(input *ClusterResourceModel) {
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

func (c *ClusterResourceModel) isReplicaChanged(other ClusterResourceModel) bool {
	return c.Replica.ValueInt64() != other.Replica.ValueInt64()
}

func (c *ClusterResourceModel) isLabelsChanged(other ClusterResourceModel) bool {
	return !c.Labels.Equal(other.Labels)
}

func (c *ClusterResourceModel) isClusterNameChanged(other ClusterResourceModel) bool {
	return c.ClusterName.ValueString() != other.ClusterName.ValueString()
}

func (plan *ClusterResourceModel) isSecurityGroupsChanged(state ClusterResourceModel) bool {
	return !plan.SecurityGroups.Equal(state.SecurityGroups)
}

func (c *ClusterResourceModel) isBucketInfoChanged(other ClusterResourceModel) bool {
	return !c.BucketInfo.Equal(other.BucketInfo)
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

func (c *ClusterResourceModel) isCuSettingsChanged(other ClusterResourceModel) bool {
	return !c.CuSettings.Equal(other.CuSettings)
}
