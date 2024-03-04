// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ClustersDataSource{}

func NewClustersDataSource() datasource.DataSource {
	return &ClustersDataSource{}
}

// ClusterDataSource defines the data source implementation.
type ClustersDataSource struct {
	client *zilliz.Client
}

// ClustersModel describes the clusters data model.
type ClustersModel struct {
	Clusters types.List   `tfsdk:"clusters"`
	Id       types.String `tfsdk:"id"`
}

func (d *ClustersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clusters"
}

func (d *ClustersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cluster data source",

		Attributes: map[string]schema.Attribute{
			"clusters": schema.ListNestedAttribute{
				MarkdownDescription: "List of Clusters",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The ID of the cluster.",
							Computed:            true,
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
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Clusters identifier",
				Computed:            true,
			},
		},
	}
}

func (d *ClustersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ClustersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClustersModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	clusters, err := d.client.ListClusters()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to ListClusters, got error: %s", err))
		return
	}

	// Save data into Terraform state
	data.Id = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))

	var cs []ClusterModel
	for _, c := range clusters.Clusters {
		cs = append(cs, ClusterModel{
			ClusterId:          types.StringValue(c.ClusterId),
			ClusterName:        types.StringValue(c.ClusterName),
			Description:        types.StringValue(c.Description),
			RegionId:           types.StringValue(c.RegionId),
			ClusterType:        types.StringValue(c.ClusterType),
			CuSize:             types.Int64Value(c.CuSize),
			Status:             types.StringValue(c.Status),
			ConnectAddress:     types.StringValue(c.ConnectAddress),
			PrivateLinkAddress: types.StringValue(c.PrivateLinkAddress),
			CreateTime:         types.StringValue(c.CreateTime)})
	}
	var diag diag.Diagnostics
	data.Clusters, diag = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: ClusterModel{}.AttrTypes()}, cs)
	resp.Diagnostics.Append(diag...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
