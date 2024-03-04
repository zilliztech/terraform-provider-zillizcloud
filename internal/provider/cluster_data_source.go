// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ClusterDataSource{}

func NewClusterDataSource() datasource.DataSource {
	return &ClusterDataSource{}
}

// ClusterDataSource defines the data source implementation.
type ClusterDataSource struct {
	client *zilliz.Client
}

// ClusterModel describes the cluster data model.
type ClusterModel struct {
	ClusterId          types.String `tfsdk:"id"`
	ClusterName        types.String `tfsdk:"cluster_name"`
	Description        types.String `tfsdk:"description"`
	RegionId           types.String `tfsdk:"region_id"`
	ClusterType        types.String `tfsdk:"cluster_type"`
	CuSize             types.Int64  `tfsdk:"cu_size"`
	Status             types.String `tfsdk:"status"`
	ConnectAddress     types.String `tfsdk:"connect_address"`
	PrivateLinkAddress types.String `tfsdk:"private_link_address"`
	CreateTime         types.String `tfsdk:"create_time"`
}

func (p ClusterModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                   types.StringType,
		"cluster_name":         types.StringType,
		"description":          types.StringType,
		"region_id":            types.StringType,
		"cluster_type":         types.StringType,
		"cu_size":              types.Int64Type,
		"status":               types.StringType,
		"connect_address":      types.StringType,
		"private_link_address": types.StringType,
		"create_time":          types.StringType,
	}
}

func (d *ClusterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (d *ClusterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cluster data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the cluster.",
				Required:            true,
			},
			"cluster_name": schema.StringAttribute{
				MarkdownDescription: "The name of the cluster.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "An optional description about the cluster.",
				Computed:            true,
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the region where the cluster exists.",
				Computed:            true,
			},
			"cluster_type": schema.StringAttribute{
				MarkdownDescription: "The type of CU associated with the cluster. Possible values are Performance-optimized and Capacity-optimized.",
				Computed:            true,
			},
			"cu_size": schema.Int64Attribute{
				MarkdownDescription: "The size of the CU associated with the cluster.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the cluster. Possible values are INITIALIZING, RUNNING, SUSPENDING, and RESUMING.",
				Computed:            true,
			},
			"connect_address": schema.StringAttribute{
				MarkdownDescription: "The public endpoint of the cluster. You can connect to the cluster using this endpoint from the public network.",
				Computed:            true,
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
	}
}

func (d *ClusterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

	d.client = client
}

func (d *ClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	c, err := d.client.DescribeCluster(data.ClusterId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to DescribeCluster, got error: %s", err))
		return
	}

	// Save data into Terraform state
	data.ClusterId = types.StringValue(c.ClusterId)
	data.ClusterName = types.StringValue(c.ClusterName)
	data.Description = types.StringValue(c.Description)
	data.RegionId = types.StringValue(c.RegionId)
	data.ClusterType = types.StringValue(c.ClusterType)
	data.CuSize = types.Int64Value(c.CuSize)
	data.Status = types.StringValue(c.Status)
	data.ConnectAddress = types.StringValue(c.ConnectAddress)
	data.PrivateLinkAddress = types.StringValue(c.PrivateLinkAddress)
	data.CreateTime = types.StringValue(c.CreateTime)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
