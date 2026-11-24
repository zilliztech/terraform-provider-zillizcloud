package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSourceWithConfigure = &ServiceAccountDataSource{}

type ServiceAccountDataSource struct {
	client *zilliz.Client
}

func NewServiceAccountDataSource() datasource.DataSource {
	return &ServiceAccountDataSource{}
}

func (d *ServiceAccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (d *ServiceAccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a single Service Account, including roles and project assignments.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Service Account ID.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Account name.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Account status.",
			},
			"project_scope": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Owning project ID, or null for an organization account.",
			},
			"org_role_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Organization role ID, or null when no organization role is granted.",
			},
			"project_role_ids": schema.SetAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Roles in the owning project for a project-local account.",
			},
			"project_assignments": schema.MapAttribute{
				Computed:            true,
				ElementType:         serviceAccountRoleSetType,
				MarkdownDescription: "Project IDs mapped to granted role ID sets.",
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

func (d *ServiceAccountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	var ok bool
	d.client, ok = req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider type", fmt.Sprintf("Expected *zilliz.Client, got %T", req.ProviderData))
	}
}

func (d *ServiceAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServiceAccountResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := d.client.GetServiceAccount(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read service_account", err.Error())
		return
	}
	resp.Diagnostics.Append(data.populate(ctx, result)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
