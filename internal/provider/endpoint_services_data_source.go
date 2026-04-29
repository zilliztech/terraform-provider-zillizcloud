package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSource = &EndpointServicesDataSource{}

func NewEndpointServicesDataSource() datasource.DataSource {
	return &EndpointServicesDataSource{}
}

type EndpointServicesDataSource struct {
	client *zilliz.Client
}

type EndpointServiceItem struct {
	RegionId          types.String `tfsdk:"region_id"`
	CloudId           types.String `tfsdk:"cloud_id"`
	EndpointService   types.String `tfsdk:"endpoint_service"`
	WhitelistRequired types.Bool   `tfsdk:"whitelist_required"`
}

type EndpointServicesDataSourceModel struct {
	RegionId         types.String          `tfsdk:"region_id"`
	CurrentPage      types.Int64           `tfsdk:"current_page"`
	PageSize         types.Int64           `tfsdk:"page_size"`
	EndpointServices []EndpointServiceItem `tfsdk:"endpoint_services"`
	TotalCount       types.Int64           `tfsdk:"total_count"`
}

func (d *EndpointServicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_endpoint_services"
}

func (d *EndpointServicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List available private link endpoint services for a region.",
		Attributes: map[string]schema.Attribute{
			"region_id": schema.StringAttribute{
				MarkdownDescription: "Cloud region ID.",
				Required:            true,
			},
			"current_page": schema.Int64Attribute{
				MarkdownDescription: "Page number (defaults to 1).",
				Optional:            true,
			},
			"page_size": schema.Int64Attribute{
				MarkdownDescription: "Page size (1-100, defaults to 10).",
				Optional:            true,
			},
			"total_count": schema.Int64Attribute{
				MarkdownDescription: "Total count of endpoint services.",
				Computed:            true,
			},
			"endpoint_services": schema.ListNestedAttribute{
				MarkdownDescription: "List of endpoint services.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"region_id":          schema.StringAttribute{Computed: true},
						"cloud_id":           schema.StringAttribute{Computed: true},
						"endpoint_service":   schema.StringAttribute{Computed: true},
						"whitelist_required": schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *EndpointServicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *zilliz.Client, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *EndpointServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EndpointServicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentPage := 1
	if !state.CurrentPage.IsNull() {
		currentPage = int(state.CurrentPage.ValueInt64())
	}
	pageSize := 10
	if !state.PageSize.IsNull() {
		pageSize = int(state.PageSize.ValueInt64())
	}

	svcs, page, err := d.client.ListEndpointServices(state.RegionId.ValueString(), currentPage, pageSize)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to ListEndpointServices, got error: %s", err))
		return
	}

	state.EndpointServices = nil
	for _, s := range svcs {
		state.EndpointServices = append(state.EndpointServices, EndpointServiceItem{
			RegionId:          types.StringValue(s.RegionId),
			CloudId:           types.StringValue(s.CloudId),
			EndpointService:   types.StringValue(s.EndpointService),
			WhitelistRequired: types.BoolValue(s.WhitelistRequired),
		})
	}
	state.TotalCount = types.Int64Value(int64(page.Count))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
