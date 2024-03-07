// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &CloudProvidersDataSource{}

func NewCloudProvidersDataSource() datasource.DataSource {
	return &CloudProvidersDataSource{}
}

// CloudProvidersDataSource defines the data source implementation.
type CloudProvidersDataSource struct {
	client *zilliz.Client
}

// CloudProviderDataSourceModel describes the data source data model.
type CloudProviderItem struct {
	Description types.String `tfsdk:"description"`
	CloudId     types.String `tfsdk:"cloud_id"`
}

// CloudProvidersDataSourceModel describes the data source data model.
type CloudProvidersDataSourceModel struct {
	CloudProviders []CloudProviderItem `tfsdk:"cloud_providers"`
}

func (d *CloudProvidersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_providers"
}

func (d *CloudProvidersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cloud Providers data source",

		Attributes: map[string]schema.Attribute{
			"cloud_providers": schema.ListNestedAttribute{
				MarkdownDescription: "List of Cloud Providers",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cloud_id": schema.StringAttribute{
							MarkdownDescription: "Cloud Provider Identifier",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Cloud Provider Description",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *CloudProvidersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CloudProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state CloudProvidersDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "sending list cloud providers request...")
	cloudProviders, err := d.client.ListCloudProviders()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to ListCloudProviders, got error: %s", err))
		return
	}

	for _, cp := range cloudProviders {
		state.CloudProviders = append(state.CloudProviders, CloudProviderItem{CloudId: types.StringValue(cp.CloudId), Description: types.StringValue(cp.Description)})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
