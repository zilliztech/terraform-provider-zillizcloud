package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSource = &DatabasesDataSource{}

func NewDatabasesDataSource() datasource.DataSource {
	return &DatabasesDataSource{}
}

type DatabasesDataSource struct {
	client *zilliz.Client
}

type DatabaseItem struct {
	DbName types.String `tfsdk:"db_name"`
}

type DatabasesDataSourceModel struct {
	ConnectAddress types.String   `tfsdk:"connect_address"`
	Items          []DatabaseItem `tfsdk:"items"`
}

func (d *DatabasesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databases"
}

func (d *DatabasesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List databases of a given cluster by connect_address",
		Attributes: map[string]schema.Attribute{
			"connect_address": schema.StringAttribute{
				MarkdownDescription: "Cluster's connection address",
				Required:            true,
			},
			"items": schema.ListNestedAttribute{
				MarkdownDescription: "List of databases",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"db_name": schema.StringAttribute{
							MarkdownDescription: "Database name",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *DatabasesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatabasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DatabasesDataSourceModel

	// Parse config input
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientCluster, err := d.client.Cluster(state.ConnectAddress.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to create cluster client for connect_address %q: %s", state.ConnectAddress.ValueString(), err))
		return
	}
	dbs, err := clientCluster.ListDatabases()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to list databases for connect_address %q: %s", state.ConnectAddress.ValueString(), err))
		return
	}

	for _, db := range dbs {
		state.Items = append(state.Items, DatabaseItem{
			DbName: types.StringValue(db),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
