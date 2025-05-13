// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSource = &ListUserDataSource{}

func NewListUserDataSource() datasource.DataSource {
	return &ListUserDataSource{}
}

type ListUserDataSource struct {
	client *zilliz.Client
}

type ListUserItem struct {
	UserId types.String `tfsdk:"user_id"`
}

type ListUserDataSourceModel struct {
	ConnectAddress types.String   `tfsdk:"connect_address"`
	Items          []ListUserItem `tfsdk:"items"`
}

func (d *ListUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *ListUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List users of a given cluster by connect_address",

		Attributes: map[string]schema.Attribute{
			"connect_address": schema.StringAttribute{
				MarkdownDescription: "Cluster's connection address",
				Required:            true,
			},
			"items": schema.ListNestedAttribute{
				MarkdownDescription: "List of users",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_id": schema.StringAttribute{
							MarkdownDescription: "User ID",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *ListUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *zilliz.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ListUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ListUserDataSourceModel

	// Parse config input
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.User(state.ConnectAddress.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to create user client for connect_address %q: %s", state.ConnectAddress.ValueString(), err))
		return
	}

	users, err := client.ListUsers()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to list users for connect_address %q: %s", state.ConnectAddress.ValueString(), err))
		return
	}

	for _, u := range users {
		state.Items = append(state.Items, ListUserItem{
			UserId: types.StringValue(string(u)),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
