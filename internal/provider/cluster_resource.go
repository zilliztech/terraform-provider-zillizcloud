// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

const (
	defaultClusterCreateTimeout time.Duration = 5 * time.Minute
	defaultClusterUpdateTimeout time.Duration = 5 * time.Minute
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ClusterResource{}
var _ resource.ResourceWithConfigure = &ClusterResource{}
var _ resource.ResourceWithImportState = &ClusterResource{}

func NewClusterResource() resource.Resource {
	return &ClusterResource{}
}

// ClusterResource defines the resource implementation.
type ClusterResource struct {
	client *zilliz.Client
}

func (r *ClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *ClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Cluster resource. If 'plan', 'cu_size' and 'cu_type' are not specified, then a free cluster is created.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Cluster identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_name": schema.StringAttribute{
				MarkdownDescription: "The name of the cluster to be created. It is a string of no more than 32 characters.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project where the cluster is to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"plan": schema.StringAttribute{
				MarkdownDescription: "The plan tier of the Zilliz Cloud service. Available options are Serverless, Standard and Enterprise.",
				Required:            true,
			},
			"cu_size": schema.Int64Attribute{
				MarkdownDescription: "The size of the CU to be used for the created cluster. It is an integer from 1 to 256.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("plan"),
						path.MatchRelative().AtParent().AtName("cu_type"),
					),
				},
			},
			"cu_type": schema.StringAttribute{
				MarkdownDescription: "The type of the CU used for the Zilliz Cloud cluster to be created. A compute unit (CU) is the physical resource unit for cluster deployment. Different CU types comprise varying combinations of CPU, memory, and storage. Available options are Performance-optimized, Capacity-optimized, and Cost-optimized. This parameter defaults to Performance-optimized. The value defaults to Performance-optimized.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("cu_size"),
						path.MatchRelative().AtParent().AtName("plan"),
					),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The name of the cluster user generated by default.",
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password of the cluster user generated by default. It will not be displayed again, so note it down and securely store it.",
				Computed:            true,
				Sensitive:           true,
			},
			"prompt": schema.StringAttribute{
				MarkdownDescription: "The statement indicating that this operation succeeds.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "An optional description about the cluster.",
				Computed:            true,
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the region where the cluster exists.",
				Optional:            true,
				Computed:            true,
			},
			"cluster_type": schema.StringAttribute{
				MarkdownDescription: "[Deprecated] The type of CU associated with the cluster. Use 'cu_type' instead. Possible values are Performance-optimized and Capacity-optimized.",
				Computed:            true,
				DeprecationMessage:  "This attribute is deprecated and will be removed in a future version. Please use 'cu_type' instead.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the cluster. Possible values are INITIALIZING, RUNNING, SUSPENDING, and RESUMING.",
				Computed:            true,
			},
			"connect_address": schema.StringAttribute{
				MarkdownDescription: "The public endpoint of the cluster. You can connect to the cluster using this endpoint from the public network.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"private_link_address": schema.StringAttribute{
				MarkdownDescription: "The private endpoint of the cluster. You can set up a private link to allow your VPS in the same cloud region to access your cluster.",
				Computed:            true,
			},
			"create_time": schema.StringAttribute{
				MarkdownDescription: "The time at which the cluster has been created.",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx,
				timeouts.Opts{
					Create: true,
					CreateDescription: `Timeout defaults to 5 mins. Accepts a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) ` +
						`consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are ` +
						`"s" (seconds), "m" (minutes), "h" (hours).`,
					Update: true,
					UpdateDescription: `Timeout defaults to 5 mins. Accepts a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) ` +
						`consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are ` +
						`"s" (seconds), "m" (minutes), "h" (hours).`,
				},
			),
		},
	}
}

func (r *ClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*zilliz.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func CloneClient(ctx context.Context, client *zilliz.Client, data *ClusterResourceModel) (*zilliz.Client, error) {
	var regionId = client.RegionId

	if data.RegionId.ValueString() != "" {
		regionId = data.RegionId.ValueString()
	}

	ctx = tflog.SetField(ctx, "RegionID", regionId)
	tflog.Info(ctx, "Clone Client...")

	return client.Clone(zilliz.WithCloudRegionId(regionId))
}

func (r *ClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Create Cluster...")
	var data ClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkPlan := func(data ClusterResourceModel) (bool, error) {

		if data.Plan.IsNull() {
			return true, nil
		}

		switch zilliz.Plan(data.Plan.ValueString()) {
		case zilliz.StandardPlan, zilliz.EnterprisePlan, zilliz.FreePlan, zilliz.ServerlessPlan:
			return true, nil
		default:
			return false, fmt.Errorf("Invalid plan: %s", data.Plan.ValueString())
		}

	}

	if _, err := checkPlan(data); err != nil {
		resp.Diagnostics.AddError("Invalid plan", err.Error())
		return
	}

	var response *zilliz.CreateClusterResponse
	var err error

	client, err := CloneClient(ctx, r.client, &data)
	if err != nil {
		resp.Diagnostics.AddError("client error", err.Error())
		return
	}

	if (data.Plan.IsNull() || zilliz.Plan(data.Plan.ValueString()) == zilliz.FreePlan || zilliz.Plan(data.Plan.ValueString()) == zilliz.ServerlessPlan) && data.CuSize.IsUnknown() && data.CuType.IsNull() {

		response, err = client.CreateServerlessCluster(zilliz.CreateServerlessClusterParams{
			Plan:        zilliz.Plan(data.Plan.ValueString()),
			ClusterName: data.ClusterName.ValueString(),
			ProjectId:   data.ProjectId.ValueString(),
		})
	} else {
		response, err = client.CreateCluster(zilliz.CreateClusterParams{
			Plan:        zilliz.Plan(data.Plan.ValueString()),
			ClusterName: data.ClusterName.ValueString(),
			CUSize:      int(data.CuSize.ValueInt64()),
			CUType:      data.CuType.ValueString(),
			ProjectId:   data.ProjectId.ValueString(),
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cluster", err.Error())
		return
	}

	data.ClusterId = types.StringValue(response.ClusterId)
	data.Username = types.StringValue(response.Username)
	data.Password = types.StringValue(response.Password)
	data.Prompt = types.StringValue(response.Prompt)

	// Wait for cluster to be RUNNING
	// Create() is passed a default timeout to use if no value
	// has been supplied in the Terraform configuration.
	createTimeout, diags := data.Timeouts.Create(ctx, defaultClusterCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.waitForStatus(ctx, createTimeout, client, "RUNNING")...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.refresh(client)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Read Cluster...")
	var state ClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := CloneClient(ctx, r.client, &state)
	if err != nil {
		resp.Diagnostics.AddError("client error", err.Error())
		return
	}
	resp.Diagnostics.Append(state.refresh(client)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Update Cluster...")

	var plan ClusterResourceModel
	var state ClusterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := CloneClient(ctx, r.client, &state)
	if err != nil {
		resp.Diagnostics.AddError("client error", err.Error())
		return
	}
	// Only support changes of cuSize - all other attributes are set to ForceNew
	_, err = client.ModifyCluster(state.ClusterId.ValueString(), &zilliz.ModifyClusterParams{
		CuSize: int(plan.CuSize.ValueInt64()),
	})

	if err != nil {
		resp.Diagnostics.AddError("Failed to modify cluster", err.Error())
		return
	}

	// Wait for cluster to be RUNNING
	// Update() is passed a default timeout to use if no value
	// has been supplied in the Terraform configuration.
	updateTimeout, diags := plan.Timeouts.Update(ctx, defaultClusterUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(state.waitForStatus(ctx, updateTimeout, client, "RUNNING")...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(state.refresh(client)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Delete Cluster...")
	var data ClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := CloneClient(ctx, r.client, &data)
	if err != nil {
		resp.Diagnostics.AddError("client error", err.Error())
		return
	}
	_, err = client.DropCluster(data.ClusterId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to drop cluster", err.Error())
		return
	}
}

func (r *ClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: clusterId,regionId. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region_id"), idParts[1])...)
}

// ClusterResourceModel describes the resource data model.
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
	ClusterType        types.String   `tfsdk:"cluster_type"`
	Status             types.String   `tfsdk:"status"`
	ConnectAddress     types.String   `tfsdk:"connect_address"`
	PrivateLinkAddress types.String   `tfsdk:"private_link_address"`
	CreateTime         types.String   `tfsdk:"create_time"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

func (data *ClusterResourceModel) refresh(client *zilliz.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	c, err := client.DescribeCluster(data.ClusterId.ValueString())
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to DescribeCluster, got error: %s", err))
		return diags
	}

	// Save data into Terraform state
	data.ClusterId = types.StringValue(c.ClusterId)
	data.ClusterName = types.StringValue(c.ClusterName)
	data.CuSize = types.Int64Value(c.CuSize)

	data.Description = types.StringValue(c.Description)
	data.RegionId = types.StringValue(c.RegionId)
	data.ClusterType = types.StringValue(c.ClusterType)
	data.Status = types.StringValue(c.Status)
	data.ConnectAddress = types.StringValue(c.ConnectAddress)
	data.PrivateLinkAddress = types.StringValue(c.PrivateLinkAddress)
	data.CreateTime = types.StringValue(c.CreateTime)
	data.ProjectId = types.StringValue(c.ProjectId)
	data.Plan = types.StringValue(string(c.Plan))
	data.CuType = types.StringValue(c.ClusterType)

	return diags
}

func (data *ClusterResourceModel) waitForStatus(ctx context.Context, timeout time.Duration, client *zilliz.Client, status string) diag.Diagnostics {
	var diags diag.Diagnostics

	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		cluster, err := client.DescribeCluster(data.ClusterId.ValueString())
		if err != nil {
			return retry.NonRetryableError(err)
		}
		if cluster.Status != status {
			return retry.RetryableError(fmt.Errorf("cluster not yet in the %s state. Current state: %s", status, cluster.Status))
		}
		return nil
	})
	if err != nil {
		diags.AddError("Failed to wait for cluster to enter the RUNNING state.", err.Error())
	}

	return diags
}
