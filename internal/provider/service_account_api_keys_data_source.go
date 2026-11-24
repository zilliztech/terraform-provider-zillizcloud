package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSourceWithConfigure = &ServiceAccountKeysDataSource{}

type ServiceAccountKeysDataSource struct {
	client *zilliz.Client
}

func NewServiceAccountKeysDataSource() datasource.DataSource {
	return &ServiceAccountKeysDataSource{}
}

type ServiceAccountKeysDataSourceModel struct {
	Name             types.String                 `tfsdk:"name"`
	ServiceAccountId types.String                 `tfsdk:"service_account_id"`
	ApiKeyId         types.String                 `tfsdk:"api_key_id"`
	Items            []ServiceAccountKeyDataModel `tfsdk:"items"`
	Count            types.Int64                  `tfsdk:"total_count"`
}

func (d *ServiceAccountKeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account_api_keys"
}

func (d *ServiceAccountKeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all matching Service Account API key metadata. Secrets are never returned. Automatically reads all pages.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Substring filter for the name.",
			},
			"service_account_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Owning Service Account ID.",
			},
			"api_key_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Substring filter for the API key ID.",
			},
			"total_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of returned items.",
			},
			"items": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Matching records.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "API key ID.",
						},
						"service_account_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Owning Service Account ID.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Key name.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Description.",
						},
						"expire_time": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Expiration time in UTC ISO 8601 format, or null.",
						},
						"created_by": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Creator identifier.",
						},
						"creator_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Creator display name.",
						},
						"create_time": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Creation time in UTC ISO 8601 format.",
						},
					},
				},
			},
		},
	}
}

func (d *ServiceAccountKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	var ok bool
	d.client, ok = req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider type", fmt.Sprintf("Expected *zilliz.Client, got %T", req.ProviderData))
	}
}

func (d *ServiceAccountKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServiceAccountKeysDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Items = make([]ServiceAccountKeyDataModel, 0)
	for page := 1; page <= 10000; page++ {
		result, err := d.client.ListServiceAccountKeys(ctx, data.ServiceAccountId.ValueString(), zilliz.ServiceAccountKeyListOptions{
			CurrentPage: page,
			PageSize:    100,
			Name:        data.Name.ValueString(),
			ApiKeyId:    data.ApiKeyId.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to list service_account_api_keys", err.Error())
			return
		}
		for _, item := range result.ApiKeys {
			data.Items = append(data.Items, serviceAccountKeyData(item))
		}
		if len(data.Items) >= result.Count {
			break
		}
		if len(result.ApiKeys) == 0 || page == 10000 {
			resp.Diagnostics.AddError("Incomplete pagination", "The API did not return all matching records. Narrow the filters and retry.")
			return
		}
	}
	data.Count = types.Int64Value(int64(len(data.Items)))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
