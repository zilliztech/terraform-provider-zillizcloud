package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

type ServiceAccountKeyDataModel struct {
	Id               types.String `tfsdk:"id"`
	ServiceAccountId types.String `tfsdk:"service_account_id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	ExpireTime       types.String `tfsdk:"expire_time"`
	CreatedBy        types.String `tfsdk:"created_by"`
	CreatorName      types.String `tfsdk:"creator_name"`
	CreateTime       types.String `tfsdk:"create_time"`
}

func serviceAccountKeyData(key zilliz.ServiceAccountKeyResponse) ServiceAccountKeyDataModel {
	return ServiceAccountKeyDataModel{
		Id:               types.StringValue(key.ApiKeyId),
		ServiceAccountId: types.StringValue(key.ServiceAccountId),
		Name:             types.StringValue(key.Name),
		Description:      types.StringValue(key.Description),
		ExpireTime:       types.StringPointerValue(key.ExpireTime),
		CreatedBy:        types.StringValue(key.CreatedBy),
		CreatorName:      types.StringValue(key.CreatorName),
		CreateTime:       types.StringValue(key.CreateTime),
	}
}

var _ datasource.DataSourceWithConfigure = &ServiceAccountKeyDataSource{}

type ServiceAccountKeyDataSource struct {
	client *zilliz.Client
}

func NewServiceAccountKeyDataSource() datasource.DataSource {
	return &ServiceAccountKeyDataSource{}
}

func (d *ServiceAccountKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account_api_key"
}

func (d *ServiceAccountKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a single Service Account API key metadata. Secrets are never returned.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "API key ID.",
			},
			"service_account_id": schema.StringAttribute{
				Required:            true,
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
	}
}

func (d *ServiceAccountKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	var ok bool
	d.client, ok = req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider type", fmt.Sprintf("Expected *zilliz.Client, got %T", req.ProviderData))
	}
}

func (d *ServiceAccountKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServiceAccountKeyDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := d.client.GetServiceAccountKey(ctx, data.ServiceAccountId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read service_account_api_key", err.Error())
		return
	}
	data = serviceAccountKeyData(*result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
