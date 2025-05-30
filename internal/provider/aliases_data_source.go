package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSource = &AliasesDataSource{}

func NewAliasesDataSource() datasource.DataSource {
	return &AliasesDataSource{}
}

type AliasesDataSource struct {
	client *zilliz.Client
}

type AliasItem struct {
	AliasName types.String `tfsdk:"alias_name"`
}

type AliasesDataSourceModel struct {
	ConnectAddress types.String `tfsdk:"connect_address"`
	DbName         types.String `tfsdk:"db_name"`
	CollectionName types.String `tfsdk:"collection_name"`
	Items          []AliasItem  `tfsdk:"items"`
}

func (d *AliasesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aliases"
}

func (d *AliasesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List aliases of a database. If collection_name is specified, only aliases for that collection are returned. If collection_name is omitted, all aliases in the database are returned.",
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
				MarkdownDescription: "Collection name. If specified, only aliases for this collection are returned. If omitted, all aliases in the database are returned.",
				Optional:            true,
			},
			"items": schema.ListNestedAttribute{
				MarkdownDescription: "List of aliases",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"alias_name": schema.StringAttribute{
							MarkdownDescription: "Alias name",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *AliasesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AliasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state AliasesDataSourceModel

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

	// Prepare ListAliases parameters
	params := &zilliz.ListAliasesParams{
		DbName: state.DbName.ValueString(),
	}

	// Only set CollectionName if it's provided
	if !state.CollectionName.IsNull() && !state.CollectionName.IsUnknown() {
		params.CollectionName = state.CollectionName.ValueString()
	}

	aliases, err := clientCollection.ListAliases(params)
	if err != nil {
		if !state.CollectionName.IsNull() && !state.CollectionName.IsUnknown() {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to list aliases for connect_address %q, db_name %q, collection_name %q: %s", state.ConnectAddress.ValueString(), state.DbName.ValueString(), state.CollectionName.ValueString(), err))
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to list aliases for connect_address %q, db_name %q: %s", state.ConnectAddress.ValueString(), state.DbName.ValueString(), err))
		}
		return
	}

	for _, alias := range aliases {
		state.Items = append(state.Items, AliasItem{
			AliasName: types.StringValue(alias),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
