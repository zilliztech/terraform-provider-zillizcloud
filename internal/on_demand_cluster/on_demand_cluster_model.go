package cluster

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const defaultOnDemandClusterCreateTimeout time.Duration = 45 * time.Minute

type OnDemandClusterResourceModel struct {
	ID                   types.String   `tfsdk:"id"`
	ProjectID            types.String   `tfsdk:"project_id"`
	RegionID             types.String   `tfsdk:"region_id"`
	ClusterName          types.String   `tfsdk:"cluster_name"`
	CUSize               types.Int64    `tfsdk:"cu_size"`
	AutoSuspend          types.Int64    `tfsdk:"auto_suspend"`
	MaxQueryNodeCU       types.Int64    `tfsdk:"max_query_node_cu"`
	MaxQueryNodeReplicas types.Int64    `tfsdk:"max_query_node_replicas"`
	Replicas             types.Int64    `tfsdk:"replicas"`
	ReadyReplicas        types.Int64    `tfsdk:"ready_replicas"`
	Status               types.String   `tfsdk:"status"`
	Endpoint             types.String   `tfsdk:"endpoint"`
	PrivateLink          types.String   `tfsdk:"private_link"`
	CreatedBy            types.String   `tfsdk:"created_by"`
	CreateTime           types.Int64    `tfsdk:"create_time"`
	TTLSeconds           types.Int64    `tfsdk:"ttl_seconds"`
	Prompt               types.String   `tfsdk:"prompt"`
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
}

type autoSuspendValidator struct{}

func (v autoSuspendValidator) Description(ctx context.Context) string {
	return "value must be at least 60 seconds"
}

func (v autoSuspendValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be at least 60 seconds"
}

func (v autoSuspendValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if req.ConfigValue.ValueInt64() < 60 {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid auto_suspend", "auto_suspend must be at least 60 seconds")
	}
}
