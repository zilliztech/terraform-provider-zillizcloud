package byoc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

// ExternalIdDataSource defines the data source implementation.
type ExternalIdDataSource struct {
	client *zilliz.Client
}

func NewExternalIdDataSource() datasource.DataSource {
	return &ExternalIdDataSource{}
}

type ExternalIdDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	ExternalId types.String `tfsdk:"external_id"`
}

func (d *ExternalIdDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_id"
}

func (d *ExternalIdDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the external ID.",
				Computed:            true,
			},
			"external_id": schema.StringAttribute{
				MarkdownDescription: "The external ID.",
				Computed:            true,
			},
		},
	}
}

func (d *ExternalIdDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ExternalIdDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ExternalIdDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	externalId, err := d.client.GetExternalId()
	if err != nil {
		resp.Diagnostics.AddError("Error getting external ID", err.Error())
		return
	}

	state.Id = types.StringValue(externalId)
	state.ExternalId = types.StringValue(externalId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
