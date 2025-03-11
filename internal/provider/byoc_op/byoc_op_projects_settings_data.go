package byoc_op

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSource = &BYOCOpProjectSettingsData{}
var _ datasource.DataSourceWithConfigure = &BYOCOpProjectSettingsData{}

func NewBYOCOpProjectSettingsData() datasource.DataSource {
	return &BYOCOpProjectSettingsData{}
}

type BYOCOpProjectSettingsData struct {
	store ByocOpProjectSettingsDataStore
}

func (r *BYOCOpProjectSettingsData) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_byoc_op_project_settings"
}

func (r *BYOCOpProjectSettingsData) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	nodeSchema := schema.SingleNestedAttribute{
		MarkdownDescription: "Fundamental VM configuration",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"disk_size": schema.Int64Attribute{
				MarkdownDescription: "Disk size in GB",
				Computed:            true,
			},
			"min_size": schema.Int64Attribute{
				MarkdownDescription: "Minimum number of instances",
				Computed:            true,
			},
			"max_size": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of instances",
				Computed:            true,
			},
			"desired_size": schema.Int64Attribute{
				MarkdownDescription: "Desired number of instances",
				Computed:            true,
			},
			"instance_types": schema.StringAttribute{
				MarkdownDescription: "Instance type",
				Computed:            true,
			},
			"capacity_type": schema.StringAttribute{
				MarkdownDescription: "Capacity type (ON_DEMAND or SPOT)",
				Computed:            true,
			},
		},
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "BYOC Op Project Settings resource for managing project configurations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Settings identifier",
				Computed:            true,
			},
			"project_name": schema.StringAttribute{
				MarkdownDescription: "The name of the project",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project",
				Required:            true,
			},
			"data_plane_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the data plane",
				Required:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "AWS region",
				Computed:            true,
			},
			"cloud_provider": schema.StringAttribute{
				MarkdownDescription: "Cloud provider",
				Computed:            true,
			},

			"op_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Operation configuration settings",
				Computed:            true,

				// Optional:            true,
				Attributes: map[string]schema.Attribute{
					"token": schema.StringAttribute{
						MarkdownDescription: "Operation token",
						Computed:            true,
					},
					"agent_image_url": schema.StringAttribute{
						MarkdownDescription: "Agent image URL",
						Computed:            true,
					},
				},
			},
			"node_quotas": schema.SingleNestedAttribute{
				Computed: true,
				// Optional:            true,
				Attributes: map[string]schema.Attribute{
					"core":        nodeSchema,
					"index":       nodeSchema,
					"search":      nodeSchema,
					"fundamental": nodeSchema,
				},
			},
		},
	}
}

func (r *BYOCOpProjectSettingsData) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *zilliz.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	// r.client = client
	r.store = &byocOpProjectSettingsDataStore{
		client: client,
	}
}

func (r *BYOCOpProjectSettingsData) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BYOCOpProjectSettingsDataModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading BYOC Op Project Settings...")
	model, err := r.store.Describe(ctx, data.ProjectID.ValueString(), data.DataPlaneID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to describe BYOC Op Project Settings after creation", err.Error())
		return
	}
	// Implement read logic here
	// Update the model with the latest data from the API
	data.refresh(model)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

type ByocOpProjectSettingsDataStore interface {
	Describe(ctx context.Context, projectID string, dataPlaneID string) (model BYOCOpProjectSettingsDataModel, err error)
}

type byocOpProjectSettingsDataStore struct {
	client *zilliz.Client
}

var _ ByocOpProjectSettingsDataStore = &byocOpProjectSettingsDataStore{}

func (s *byocOpProjectSettingsDataStore) Describe(ctx context.Context, projectID string, dataPlaneID string) (data BYOCOpProjectSettingsDataModel, err error) {

	{
		response, err := s.client.DescribeByocOpProject(&zilliz.DescribeByocOpProjectRequest{
			ProjectId:   projectID,
			DataPlaneID: dataPlaneID,
		})
		if err != nil {
			return data, fmt.Errorf("failed to describe BYOC Op project: %w", err)
		}
		data.ID = types.StringValue(response.ProjectID)
		data.DataPlaneID = types.StringValue(response.DataPlaneID)
		data.ProjectID = types.StringValue(response.ProjectID)
		data.ProjectName = types.StringValue(response.ProjectName)
		data.CloudProvider = types.StringValue(response.CloudID)
		data.Region = types.StringValue(response.RegionID)
	}

	{
		response, err := s.client.DescribeByocOpProjectSettings(&zilliz.DescribeByocOpProjectSettingsRequest{
			ProjectId:   projectID,
			DataPlaneId: dataPlaneID,
		})
		if err != nil {
			return data, fmt.Errorf("failed to describe BYOC Op project settings: %w", err)
		}

		OpConfig, diag := types.ObjectValue(map[string]attr.Type{
			"token":           types.StringType,
			"agent_image_url": types.StringType,
		}, map[string]attr.Value{
			"token":           types.StringValue(response.OpConfig.Token),
			"agent_image_url": types.StringValue(response.OpConfig.AgentImageUrl),
		})
		if diag.HasError() {
			return data, fmt.Errorf("failed to abstract OpConfig from response")
		}
		data.OpConfig = OpConfig

		core, err := buildNodeQuotas("core", response.NodeQuotas)
		if err != nil {
			return data, err
		}

		index, err := buildNodeQuotas("index", response.NodeQuotas)
		if err != nil {
			return data, err
		}

		search, err := buildNodeQuotas("search", response.NodeQuotas)
		if err != nil {
			return data, err
		}

		fundamental, err := buildNodeQuotas("fundamental", response.NodeQuotas)
		if err != nil {
			return data, err
		}

		NodeQuotas, diag := types.ObjectValue(nodeQuotasGenerateAttrTypes, map[string]attr.Value{
			"core":        core,
			"index":       index,
			"search":      search,
			"fundamental": fundamental,
		})
		if diag.HasError() {
			return data, fmt.Errorf("failed to abstract NodeQuotas from response")
		}
		data.NodeQuotas = NodeQuotas
	}

	return data, nil
}
