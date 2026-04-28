package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSource = &EndpointsDataSource{}

func NewEndpointsDataSource() datasource.DataSource {
	return &EndpointsDataSource{}
}

type EndpointsDataSource struct {
	client *zilliz.Client
}

type EndpointItem struct {
	RegionId              types.String `tfsdk:"region_id"`
	CloudId               types.String `tfsdk:"cloud_id"`
	EndpointService       types.String `tfsdk:"endpoint_service"`
	EndpointServiceStatus types.String `tfsdk:"endpoint_service_status"`
	EndpointId            types.String `tfsdk:"endpoint_id"`
	EndpointStatus        types.String `tfsdk:"endpoint_status"`
	GcpProjectId          types.String `tfsdk:"gcp_project_id"`
}

type EndpointsDataSourceModel struct {
	ProjectId   types.String   `tfsdk:"project_id"`
	CurrentPage types.Int64    `tfsdk:"current_page"`
	PageSize    types.Int64    `tfsdk:"page_size"`
	Endpoints   []EndpointItem `tfsdk:"endpoints"`
	TotalCount  types.Int64    `tfsdk:"total_count"`
}

func (d *EndpointsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_endpoints"
}

func (d *EndpointsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List private link endpoints under a project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID.",
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
				MarkdownDescription: "Total count of endpoints.",
				Computed:            true,
			},
			"endpoints": schema.ListNestedAttribute{
				MarkdownDescription: "List of endpoints.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"region_id":               schema.StringAttribute{Computed: true},
						"cloud_id":                schema.StringAttribute{Computed: true},
						"endpoint_service":        schema.StringAttribute{Computed: true},
						"endpoint_service_status": schema.StringAttribute{Computed: true},
						"endpoint_id":             schema.StringAttribute{Computed: true},
						"endpoint_status":         schema.StringAttribute{Computed: true},
						"gcp_project_id":          schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *EndpointsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EndpointsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EndpointsDataSourceModel
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

	eps, page, err := d.client.ListEndpoints(state.ProjectId.ValueString(), currentPage, pageSize)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to ListEndpoints, got error: %s", err))
		return
	}

	state.Endpoints = nil
	for _, e := range eps {
		gcp := types.StringNull()
		if e.GcpProjectId != nil {
			gcp = types.StringValue(*e.GcpProjectId)
		}
		state.Endpoints = append(state.Endpoints, EndpointItem{
			RegionId:              types.StringValue(e.RegionId),
			CloudId:               types.StringValue(e.CloudId),
			EndpointService:       types.StringValue(e.EndpointService),
			EndpointServiceStatus: types.StringValue(e.EndpointServiceStatus),
			EndpointId:            types.StringValue(e.EndpointId),
			EndpointStatus:        types.StringValue(e.EndpointStatus),
			GcpProjectId:          gcp,
		})
	}
	state.TotalCount = types.Int64Value(int64(page.Count))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
