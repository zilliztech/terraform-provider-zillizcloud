package cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

type ClusterVersionsDataSource struct {
	client *zilliz.Client
}

type clusterVersionsModel struct {
	RegionID types.String          `tfsdk:"region_id"`
	BackupID types.String          `tfsdk:"backup_id"`
	Regions  []regionVersionsModel `tfsdk:"regions"`
}

type regionVersionsModel struct {
	RegionID types.String            `tfsdk:"region_id"`
	Versions []versionSelectionModel `tfsdk:"versions"`
}

func NewClusterVersionsDataSource() datasource.DataSource {
	return &ClusterVersionsDataSource{}
}

func (d *ClusterVersionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_versions"
}

func (d *ClusterVersionsDataSource) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{datasourcevalidator.AtLeastOneOf(path.MatchRoot("region_id"), path.MatchRoot("backup_id"))}
}

func (d *ClusterVersionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists eligible Milvus version series and channels for Dedicated cluster creation or backup restoration.",
		Attributes: map[string]schema.Attribute{
			"region_id": schema.StringAttribute{
				MarkdownDescription: "Region to query. Required for creation. During restore, a region other than the backup's source region yields an empty list.",
				Optional:            true,
				Validators:          []validator.String{stringvalidator.RegexMatches(nonBlank, "Provide a non-blank region ID.")},
			},
			"backup_id": schema.StringAttribute{
				MarkdownDescription: "Backup ID to query for restoration. The API resolves the source region when region_id is omitted.",
				Optional:            true,
				Validators:          []validator.String{stringvalidator.RegexMatches(nonBlank, "Provide a non-blank backup ID.")},
			},
			"regions": schema.ListNestedAttribute{
				MarkdownDescription: "Eligible versions grouped by region. The API returns at most one region. List order does not imply a default selection.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"region_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Region ID."},
					"versions": schema.ListNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Eligible Milvus versions.",
						NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
							"milvus_version": schema.StringAttribute{Computed: true, MarkdownDescription: "Milvus version series."},
							"channel":        schema.StringAttribute{Computed: true, MarkdownDescription: "Version channel, such as GA, PUBLIC_PREVIEW, or ORIGIN."},
						}},
					},
				}},
			},
		},
	}
}

func (d *ClusterVersionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected data source configuration", "Expected a Zilliz Cloud client.")
		return
	}
	d.client = client
}

func (d *ClusterVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clusterVersionsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	groups, err := d.client.ListClusterVersions(ctx, state.RegionID.ValueString(), state.BackupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to list cluster versions", err.Error())
		return
	}
	state.Regions = make([]regionVersionsModel, 0, len(groups))
	for _, group := range groups {
		versions := make([]versionSelectionModel, 0, len(group.Versions))
		for _, candidate := range group.Versions {
			versions = append(versions, versionSelectionModel{
				MilvusVersion: types.StringValue(candidate.MilvusVersion),
				Channel:       types.StringValue(candidate.Channel),
			})
		}
		state.Regions = append(state.Regions, regionVersionsModel{RegionID: types.StringValue(group.RegionID), Versions: versions})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
