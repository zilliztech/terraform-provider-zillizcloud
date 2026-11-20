package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSource = &ApiKeysDataSource{}

func NewApiKeysDataSource() datasource.DataSource {
	return &ApiKeysDataSource{}
}

type ApiKeysDataSource struct {
	client *zilliz.Client
}

type ApiKeysDataSourceProjectModel struct {
	ProjectId   types.String `tfsdk:"project_id"`
	ProjectName types.String `tfsdk:"project_name"`
	Role        types.String `tfsdk:"role"`
	AllCluster  types.Bool   `tfsdk:"all_cluster"`
	ClusterIds  types.List   `tfsdk:"cluster_ids"`
	AllVolume   types.Bool   `tfsdk:"all_volume"`
	VolumeIds   types.List   `tfsdk:"volume_ids"`
}

type ApiKeysDataSourceItemModel struct {
	ApiKeyId    types.String                    `tfsdk:"api_key_id"`
	Name        types.String                    `tfsdk:"name"`
	Description types.String                    `tfsdk:"description"`
	CreatorName types.String                    `tfsdk:"creator_name"`
	CreatedBy   types.String                    `tfsdk:"created_by"`
	OrgRole     types.String                    `tfsdk:"org_role"`
	CreateTime  types.String                    `tfsdk:"create_time"`
	Projects    []ApiKeysDataSourceProjectModel `tfsdk:"projects"`
}

type ApiKeysDataSourceModel struct {
	ApiKeys []ApiKeysDataSourceItemModel `tfsdk:"api_keys"`
}

func (d *ApiKeysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_keys"
}

func (d *ApiKeysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Lists all Customized API keys in the organization.

**Important:** This data source requires an Org Owner API key. A project-scoped key
cannot list API keys (error 96041).`,
		Attributes: map[string]schema.Attribute{
			"api_keys": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of all Customized API keys in the organization.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"api_key_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the API key.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the API key.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Description of the API key.",
						},
						"creator_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The display name of the API key creator.",
						},
						"created_by": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The creator identifier (email or key ID).",
						},
						"org_role": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The organization role.",
						},
						"create_time": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Creation time in ISO 8601 format.",
						},
						"projects": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Project access configuration.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"project_id": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The project ID.",
									},
									"project_name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The project name.",
									},
									"role": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The project role.",
									},
									"all_cluster": schema.BoolAttribute{
										Computed:            true,
										MarkdownDescription: "Whether all clusters are included.",
									},
									"cluster_ids": schema.ListAttribute{
										Computed:            true,
										ElementType:         types.StringType,
										MarkdownDescription: "Specific cluster IDs if not all clusters.",
									},
									"all_volume": schema.BoolAttribute{
										Computed:            true,
										MarkdownDescription: "Whether all volumes are included.",
									},
									"volume_ids": schema.ListAttribute{
										Computed:            true,
										ElementType:         types.StringType,
										MarkdownDescription: "Specific volume IDs if not all volumes.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *ApiKeysDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ApiKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	apiKeys, err := d.client.ListApiKeys()
	if err != nil {
		resp.Diagnostics.AddError("Failed to list API keys", err.Error())
		return
	}

	var state ApiKeysDataSourceModel
	for _, k := range apiKeys {
		item := ApiKeysDataSourceItemModel{
			ApiKeyId:    types.StringValue(k.ApiKeyId),
			Name:        types.StringValue(k.Name),
			Description: types.StringValue(k.Description),
			CreatorName: types.StringValue(k.CreatorName),
			CreatedBy:   types.StringValue(k.CreatedBy),
			OrgRole:     types.StringValue(k.OrgRole),
			CreateTime:  types.StringValue(k.CreateTime),
		}

		for _, p := range k.Projects {
			proj := ApiKeysDataSourceProjectModel{
				ProjectId:   types.StringValue(p.ProjectId),
				ProjectName: types.StringValue(p.ProjectName),
				Role:        types.StringValue(p.Role),
				AllCluster:  types.BoolValue(p.AllCluster),
				AllVolume:   types.BoolValue(p.AllVolume),
			}

			if len(p.Clusters) > 0 {
				vals := make([]attr.Value, 0, len(p.Clusters))
				for _, c := range p.Clusters {
					vals = append(vals, types.StringValue(c.ClusterId))
				}
				proj.ClusterIds, _ = types.ListValue(types.StringType, vals)
			} else {
				proj.ClusterIds = types.ListValueMust(types.StringType, []attr.Value{})
			}

			if len(p.Volumes) > 0 {
				vals := make([]attr.Value, 0, len(p.Volumes))
				for _, v := range p.Volumes {
					vals = append(vals, types.StringValue(v.VolumeId))
				}
				proj.VolumeIds, _ = types.ListValue(types.StringType, vals)
			} else {
				proj.VolumeIds = types.ListValueMust(types.StringType, []attr.Value{})
			}

			item.Projects = append(item.Projects, proj)
		}

		// Ensure nil projects becomes empty slice for consistent state
		if item.Projects == nil {
			item.Projects = []ApiKeysDataSourceProjectModel{}
		}

		state.ApiKeys = append(state.ApiKeys, item)
	}

	// Ensure nil api_keys becomes empty slice
	if state.ApiKeys == nil {
		state.ApiKeys = []ApiKeysDataSourceItemModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
