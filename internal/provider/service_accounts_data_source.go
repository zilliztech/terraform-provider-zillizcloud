package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSourceWithConfigure = &ServiceAccountsDataSource{}

type ServiceAccountsDataSource struct {
	client *zilliz.Client
}

func NewServiceAccountsDataSource() datasource.DataSource {
	return &ServiceAccountsDataSource{}
}

type ServiceAccountsDataSourceModel struct {
	Name             types.String                  `tfsdk:"name"`
	ServiceAccountId types.String                  `tfsdk:"service_account_id"`
	ProjectId        types.String                  `tfsdk:"project_id"`
	Status           types.String                  `tfsdk:"status"`
	Items            []ServiceAccountResourceModel `tfsdk:"items"`
	Count            types.Int64                   `tfsdk:"total_count"`
}

func (d *ServiceAccountsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_accounts"
}

func (d *ServiceAccountsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all matching Service Accounts, including roles and project assignments. Automatically reads all pages.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Substring filter for the name.",
			},
			"service_account_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Substring filter for the Service Account ID.",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Assigned project ID filter.",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by active or disable status.",
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
				},
			},
		},
	}
}

func (d *ServiceAccountsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	var ok bool
	d.client, ok = req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider type", fmt.Sprintf("Expected *zilliz.Client, got %T", req.ProviderData))
	}
}

func (d *ServiceAccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServiceAccountsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Items = make([]ServiceAccountResourceModel, 0)
	for page := 1; page <= 10000; page++ {
		result, err := d.client.ListServiceAccounts(ctx, zilliz.ServiceAccountListOptions{
			CurrentPage:      page,
			PageSize:         100,
			Name:             data.Name.ValueString(),
			ServiceAccountId: data.ServiceAccountId.ValueString(),
			ProjectId:        data.ProjectId.ValueString(),
			Status:           data.Status.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to list service_accounts", err.Error())
			return
		}
		for _, item := range result.ServiceAccounts {
			var record ServiceAccountResourceModel
			resp.Diagnostics.Append(record.populate(ctx, &item)...)
			data.Items = append(data.Items, record)
		}
		if len(data.Items) >= result.Count {
			break
		}
		if len(result.ServiceAccounts) == 0 || page == 10000 {
			resp.Diagnostics.AddError("Incomplete pagination", "The API did not return all matching records. Narrow the filters and retry.")
			return
		}
	}
	data.Count = types.Int64Value(int64(len(data.Items)))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
