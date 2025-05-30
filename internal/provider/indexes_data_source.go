package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSource = &IndexesDataSource{}

func NewIndexesDataSource() datasource.DataSource {
	return &IndexesDataSource{}
}

type IndexesDataSource struct {
	client *zilliz.Client
}

type IndexItem struct {
	IndexName types.String `tfsdk:"index_name"`
}

type IndexesDataSourceModel struct {
	ConnectAddress types.String `tfsdk:"connect_address"`
	DbName         types.String `tfsdk:"db_name"`
	CollectionName types.String `tfsdk:"collection_name"`
	Items          []IndexItem  `tfsdk:"items"`
}

func (d *IndexesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_indexes"
}

func (d *IndexesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List indexes of a given collection by connect_address, db_name, and collection_name",
		Attributes: map[string]schema.Attribute{
			"connect_address": schema.StringAttribute{
				MarkdownDescription: `The connection address of the target Zilliz Cloud cluster.
You can obtain this value from the output of the ` + "`zillizcloud_cluster`" + ` resource, for example:
` + "`zillizcloud_cluster.example.connect_address`" + `

**Example:**
` + "`https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534`" + `

> **Note:** The address must include the protocol (e.g., ` + "`https://`" + `).`,
				Required: true,
			},
			"db_name": schema.StringAttribute{
				MarkdownDescription: "Database name",
				Required:            true,
			},
			"collection_name": schema.StringAttribute{
				MarkdownDescription: "Collection name",
				Required:            true,
			},
			"items": schema.ListNestedAttribute{
				MarkdownDescription: "List of indexes",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index_name": schema.StringAttribute{
							MarkdownDescription: "Index name",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *IndexesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IndexesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state IndexesDataSourceModel

	// Parse config input
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientCollection, err := d.client.Collection(state.ConnectAddress.ValueString(), state.DbName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to create collection client for connect_address %q, db_name %q: %s", state.ConnectAddress.ValueString(), state.DbName.ValueString(), err))
		return
	}
	indexes, err := clientCollection.ListIndex(&zilliz.ListIndexParams{
		DbName:         state.DbName.ValueString(),
		CollectionName: state.CollectionName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to list indexes for connect_address %q, db_name %q, collection_name %q: %s", state.ConnectAddress.ValueString(), state.DbName.ValueString(), state.CollectionName.ValueString(), err))
		return
	}

	for _, idx := range indexes {
		state.Items = append(state.Items, IndexItem{
			IndexName: types.StringValue(idx),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
