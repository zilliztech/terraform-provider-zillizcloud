// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
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

type CloudRegionItem struct {
	ApiBaseUrl            types.String   `tfsdk:"api_base_url"`
	CloudId               types.String   `tfsdk:"cloud_id"`
	RegionId              types.String   `tfsdk:"region_id"`
	Domain                types.String   `tfsdk:"domain"`
	SupportedClusterTypes []types.String `tfsdk:"supported_cluster_types"`
}

// CloudRegionsDataSourceModel describes the data source data model.
type CloudRegionsDataSourceModel struct {
	CloudRegions []CloudRegionItem `tfsdk:"items"`
	CloudId      types.String      `tfsdk:"cloud_id"`
}

func (d *CloudRegionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

func (d *CloudRegionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cloud Regions data source",

		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{
				MarkdownDescription: "List of Cloud Regions",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"api_base_url": schema.StringAttribute{
							MarkdownDescription: "Cloud Region API Base URL. Deprecated: use `domain` instead.",
							Computed:            true,
							DeprecationMessage:  "api_base_url is deprecated. Use domain instead.",
						},
						"cloud_id": schema.StringAttribute{
							MarkdownDescription: "Cloud Region Identifier",
							Computed:            true,
						},
						"region_id": schema.StringAttribute{
							MarkdownDescription: "Cloud Region Id",
							Computed:            true,
						},
						"domain": schema.StringAttribute{
							MarkdownDescription: "Cloud Region control API domain",
							Computed:            true,
						},
						"supported_cluster_types": schema.ListAttribute{
							MarkdownDescription: "Supported cluster types in this region",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
			"cloud_id": schema.StringAttribute{
				MarkdownDescription: "Cloud ID",
				Optional:            true,
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
	var state CloudRegionsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "sending list cloud regions request...")
	cloudRegions, err := d.client.ListCloudRegions(state.CloudId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to ListCloudRegions, got error: %s", err))
		return
	}

	// Save data into Terraform state

	for _, cr := range cloudRegions {
		supportedClusterTypes := make([]types.String, 0, len(cr.SupportedClusterTypes))
		for _, clusterType := range cr.SupportedClusterTypes {
			supportedClusterTypes = append(supportedClusterTypes, types.StringValue(clusterType))
		}
		state.CloudRegions = append(state.CloudRegions, CloudRegionItem{
			ApiBaseUrl:            types.StringValue(apiBaseUrlFromCloudRegion(cr)),
			CloudId:               types.StringValue(cr.CloudId),
			RegionId:              types.StringValue(cr.RegionId),
			Domain:                types.StringValue(cr.Domain),
			SupportedClusterTypes: supportedClusterTypes})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func apiBaseUrlFromCloudRegion(cr zilliz.CloudRegion) string {
	if cr.ApiBaseUrl != "" {
		return cr.ApiBaseUrl
	}

	domain := strings.TrimSpace(cr.Domain)
	if domain == "" {
		return ""
	}
	domain = strings.TrimRight(domain, "/")
	if strings.HasSuffix(domain, "/v2") {
		return domain
	}
	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		return domain + "/v2"
	}
	return "https://" + domain + "/v2"
}
