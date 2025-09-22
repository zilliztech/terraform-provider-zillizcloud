package cluster

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/util/conv"
)

var _ resource.Resource = &ClusterLoadBalancerSecurityGroupsResource{}
var _ resource.ResourceWithConfigure = &ClusterLoadBalancerSecurityGroupsResource{}
var _ resource.ResourceWithImportState = &ClusterLoadBalancerSecurityGroupsResource{}

func NewClusterLoadBalancerSecurityGroupsResource() resource.Resource {
	return &ClusterLoadBalancerSecurityGroupsResource{}
}

type ClusterLoadBalancerSecurityGroupsResource struct {
	client *zilliz.Client
}

type ClusterLoadBalancerSecurityGroupsResourceModel struct {
	Id               types.String `tfsdk:"id"`
	ClusterId        types.String `tfsdk:"cluster_id"`
	SecurityGroupIds types.Set    `tfsdk:"security_group_ids"`
}

func (r *ClusterLoadBalancerSecurityGroupsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_load_balancer_security_groups"
}

func (r *ClusterLoadBalancerSecurityGroupsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages the load balancer security groups for a Zilliz Cloud cluster.

This resource allows you to associate security groups with the load balancer of a specific cluster. This provides network-level security control for cluster access.

Typical use case: Configure network access controls and security boundaries for cluster load balancers in cloud environments.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: `The ID of the cluster load balancer security groups.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `The ID of the cluster to associate security groups with.

You can obtain this value from the output of the ` + "`zillizcloud_cluster`" + ` resource, for example:
` + "`zillizcloud_cluster.example.id`" + `

**Example:**
` + "`in01-1234567890abcdef`" + `

> **Note:** Changing this value will force recreation of the resource.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"security_group_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				MarkdownDescription: `A set of security group IDs to associate with the load balancer of the cluster.

**Example:**
` + "`" + `["sg-1234567890abcdef0"]` + "`" + `

> **Note:** The security groups must exist in the same VPC as the cluster and be accessible to the Zilliz Cloud service.`,
			},
		},
	}
}

func (r *ClusterLoadBalancerSecurityGroupsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *zilliz.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ClusterLoadBalancerSecurityGroupsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ClusterLoadBalancerSecurityGroupsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform set to Go slice
	var securityGroupIds []string
	if !data.SecurityGroupIds.IsNull() && !data.SecurityGroupIds.IsUnknown() {
		elements := data.SecurityGroupIds.Elements()
		for _, elem := range elements {
			if strValue, ok := elem.(types.String); ok {
				securityGroupIds = append(securityGroupIds, strValue.ValueString())
			}
		}
	}

	// Call the API to upsert security groups
	_, err := r.client.UpsertSecurityGroups(data.ClusterId.ValueString(), &zilliz.UpsertSecurityGroupsParams{
		Ids: securityGroupIds,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cluster load balancer security groups", err.Error())
		return
	}

	// Set ID to cluster_id
	data.Id = data.ClusterId

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterLoadBalancerSecurityGroupsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ClusterLoadBalancerSecurityGroupsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current security groups from API
	securityGroups, err := r.client.GetSecurityGroups(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read cluster load balancer security groups", err.Error())
		return
	}

	// Convert to Terraform set
	securityGroupsSet := conv.SliceToSet(securityGroups)
	state.SecurityGroupIds = securityGroupsSet

	state.ClusterId = state.Id

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ClusterLoadBalancerSecurityGroupsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ClusterLoadBalancerSecurityGroupsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	securityGroupIds, err := conv.SetToSlice[string](ctx, plan.SecurityGroupIds)
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert security group ids", err.Error())
		return
	}

	// Call the API to upsert security groups
	_, err = r.client.UpsertSecurityGroups(plan.ClusterId.ValueString(), &zilliz.UpsertSecurityGroupsParams{
		Ids: securityGroupIds,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update cluster load balancer security groups", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ClusterLoadBalancerSecurityGroupsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ClusterLoadBalancerSecurityGroupsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Clear security groups by passing empty array
	_, err := r.client.UpsertSecurityGroups(data.ClusterId.ValueString(), &zilliz.UpsertSecurityGroupsParams{
		Ids: []string{},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete cluster load balancer security groups", err.Error())
		return
	}
}

func (r *ClusterLoadBalancerSecurityGroupsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

}
