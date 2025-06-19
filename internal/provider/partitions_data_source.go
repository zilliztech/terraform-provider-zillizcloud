package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSource = &PartitionsDataSource{}

func NewPartitionsDataSource() datasource.DataSource {
	return &PartitionsDataSource{}
}

type PartitionsDataSource struct {
	client *zilliz.Client
}

type PartitionItem struct {
	PartitionName types.String `tfsdk:"partition_name"`
}

type PartitionsDataSourceModel struct {
	ConnectAddress types.String    `tfsdk:"connect_address"`
	DbName         types.String    `tfsdk:"db_name"`
	CollectionName types.String    `tfsdk:"collection_name"`
	Items          []PartitionItem `tfsdk:"items"`
}

func (d *PartitionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partitions"
}

func (d *PartitionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List partitions of a given collection by connect_address, db_name, and collection_name",
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
				MarkdownDescription: "List of partitions",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"partition_name": schema.StringAttribute{
							MarkdownDescription: "Partition name",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *PartitionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PartitionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state PartitionsDataSourceModel

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
	partitions, err := clientCollection.ListPartitionses(&zilliz.ListPartitionsesParams{
		DbName:         state.DbName.ValueString(),
		CollectionName: state.CollectionName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to list partitions for connect_address %q, db_name %q, collection_name %q: %s", state.ConnectAddress.ValueString(), state.DbName.ValueString(), state.CollectionName.ValueString(), err))
		return
	}

	for _, partition := range partitions {
		state.Items = append(state.Items, PartitionItem{
			PartitionName: types.StringValue(partition),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
