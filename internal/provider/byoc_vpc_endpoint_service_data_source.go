package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSource = &ByocVpcEndpointServiceDataSource{}

func NewByocVpcEndpointServiceDataSource() datasource.DataSource {
	return &ByocVpcEndpointServiceDataSource{}
}

type ByocVpcEndpointServiceDataSource struct {
	client *zilliz.Client
}

type ByocVpcEndpointServiceDataSourceModel struct {
	CloudId         types.String `tfsdk:"cloud_id"`
	Region          types.String `tfsdk:"region"`
	EndpointService types.String `tfsdk:"endpoint_service"`
}

func (d *ByocVpcEndpointServiceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_byoc_vpc_endpoint_service"
}

func (d *ByocVpcEndpointServiceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get the private endpoint service used to connect a BYOC dataplane to the Zilliz control plane.",
		Attributes: map[string]schema.Attribute{
			"cloud_id": schema.StringAttribute{
				MarkdownDescription: "Cloud provider ID, for example aws.",
				Required:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Cloud-native region, for example ap-southeast-7.",
				Required:            true,
			},
			"endpoint_service": schema.StringAttribute{
				MarkdownDescription: "Full private endpoint service name.",
				Computed:            true,
			},
		},
	}
}

func (d *ByocVpcEndpointServiceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *zilliz.Client, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ByocVpcEndpointServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ByocVpcEndpointServiceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.GetByocVpcEndpointService(state.CloudId.ValueString(), state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to get BYOC VPC endpoint service", err.Error())
		return
	}
	state.EndpointService = types.StringValue(result.EndpointService)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
