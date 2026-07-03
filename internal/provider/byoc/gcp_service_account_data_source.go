package byoc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

// GCPServiceAccountDataSource defines the data source implementation.
type GCPServiceAccountDataSource struct {
	client *zilliz.Client
}

func NewGCPServiceAccountDataSource() datasource.DataSource {
	return &GCPServiceAccountDataSource{}
}

type GCPServiceAccountDataSourceModel struct {
	Id             types.String `tfsdk:"id"`
	ServiceAccount types.String `tfsdk:"service_account"`
}

func (d *GCPServiceAccountDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcp_service_account"
}

func (d *GCPServiceAccountDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the GCP service account.",
				Computed:            true,
			},
			"service_account": schema.StringAttribute{
				MarkdownDescription: "The GCP service account.",
				Computed:            true,
			},
		},
	}
}

func (d *GCPServiceAccountDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GCPServiceAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state GCPServiceAccountDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceAccount, err := d.client.GetGoogleServiceAccount()
	if err != nil {
		resp.Diagnostics.AddError("Error getting GCP service account", err.Error())
		return
	}

	state.Id = types.StringValue(serviceAccount)
	state.ServiceAccount = types.StringValue(serviceAccount)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
