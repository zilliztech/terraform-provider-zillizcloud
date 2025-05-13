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

var _ datasource.DataSource = &ListRoleDataSource{}

func NewListRoleDataSource() datasource.DataSource {
	return &ListRoleDataSource{}
}

type ListRoleDataSource struct {
	client *zilliz.Client
}

type ListRoleItem struct {
	RoleId types.String `tfsdk:"role_id"`
}

type ListRoleDataSourceModel struct {
	ConnectAddress types.String   `tfsdk:"connect_address"`
	Items          []ListRoleItem `tfsdk:"items"`
}

func (d *ListRoleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

func (d *ListRoleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List roles of a given cluster by connect_address",

		Attributes: map[string]schema.Attribute{
			"connect_address": schema.StringAttribute{
				MarkdownDescription: "Cluster's connection address",
				Required:            true,
			},
			"items": schema.ListNestedAttribute{
				MarkdownDescription: "List of roles",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role_id": schema.StringAttribute{
							MarkdownDescription: "Role ID",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *ListRoleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ListRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ListRoleDataSourceModel

	// Parse config input
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.Role(state.ConnectAddress.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to get role client for connect_address %q: %s", state.ConnectAddress.ValueString(), err))
		return
	}

	roles, err := client.ListRoles()
	if err != nil {
		resp.Diagnostics.AddError("List Roles Error", fmt.Sprintf("Failed to list roles for connect_address %q: %s", state.ConnectAddress.ValueString(), err))
		return
	}

	for _, u := range roles {
		state.Items = append(state.Items, ListRoleItem{
			RoleId: types.StringValue(string(u)),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
