package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSource = &CollectionsDataSource{}

func NewCollectionsDataSource() datasource.DataSource {
	return &CollectionsDataSource{}
}

type CollectionsDataSource struct {
	client *zilliz.Client
}

type CollectionItem struct {
	CollectionName types.String `tfsdk:"collection_name"`
}

type CollectionsDataSourceModel struct {
	ConnectAddress types.String     `tfsdk:"connect_address"`
	DbName         types.String     `tfsdk:"db_name"`
	Items          []CollectionItem `tfsdk:"items"`
}

func (d *CollectionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_collections"
}

func (d *CollectionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List collections of a given database by connect_address and db_name",
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
			"items": schema.ListNestedAttribute{
				MarkdownDescription: "List of collections",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"collection_name": schema.StringAttribute{
							MarkdownDescription: "Collection name",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *CollectionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CollectionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state CollectionsDataSourceModel

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
	collections, err := clientCollection.ListCollections(&zilliz.ListCollectionsParams{
		DbName: state.DbName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to list collections for connect_address %q, db_name %q: %s", state.ConnectAddress.ValueString(), state.DbName.ValueString(), err))
		return
	}

	for _, col := range collections {
		state.Items = append(state.Items, CollectionItem{
			CollectionName: types.StringValue(col),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
