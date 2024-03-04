// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure Region defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &CloudRegionsDataSource{}

func NewCloudRegionsDataSource() datasource.DataSource {
	return &CloudRegionsDataSource{}
}

// CloudRegionsDataSource defines the data source implementation.
type CloudRegionsDataSource struct {
	client *zilliz.Client
}

// CloudRegionDataSourceModel describes the data source data model.
type CloudRegionModel struct {
	ApiBaseUrl types.String `tfsdk:"api_base_url"`
	CloudId    types.String `tfsdk:"cloud_id"`
	RegionId   types.String `tfsdk:"region_id"`
}

func (p CloudRegionModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"api_base_url": types.StringType,
		"cloud_id":     types.StringType,
		"region_id":    types.StringType,
	}
}

// CloudRegionsDataSourceModel describes the data source data model.
type CloudRegionsDataSourceModel struct {
	CloudRegions types.List   `tfsdk:"cloud_regions"`
	CloudId      types.String `tfsdk:"cloud_id"`
	Id           types.String `tfsdk:"id"`
}

func (d *CloudRegionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_regions"
}

func (d *CloudRegionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cloud Regions data source",

		Attributes: map[string]schema.Attribute{
			"cloud_regions": schema.ListNestedAttribute{
				MarkdownDescription: "List of Cloud Regions",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"api_base_url": schema.StringAttribute{
							MarkdownDescription: "Cloud Region API Base URL",
							Computed:            true,
						},
						"cloud_id": schema.StringAttribute{
							MarkdownDescription: "Cloud Region Identifier",
							Computed:            true,
						},
						"region_id": schema.StringAttribute{
							MarkdownDescription: "Cloud Region Id",
							Computed:            true,
						},
					},
				},
			},
			"cloud_id": schema.StringAttribute{
				MarkdownDescription: "Cloud ID",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Cloud Regions identifier",
				Computed:            true,
			},
		},
	}
}

func (d *CloudRegionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CloudRegionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CloudRegionsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	cloudRegions, err := d.client.ListCloudRegions(data.CloudId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to ListCloudRegions, got error: %s", err))
		return
	}

	// Save data into Terraform state
	data.Id = data.CloudId

	var crs []CloudRegionModel
	for _, cr := range cloudRegions {
		crs = append(crs, CloudRegionModel{
			ApiBaseUrl: types.StringValue(cr.ApiBaseUrl),
			CloudId:    types.StringValue(cr.CloudId),
			RegionId:   types.StringValue(cr.RegionId)})
	}
	var diag diag.Diagnostics
	data.CloudRegions, diag = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: CloudRegionModel{}.AttrTypes()}, crs)
	resp.Diagnostics.Append(diag...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
