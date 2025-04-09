package byoc_op

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ resource.Resource = &BYOCOpProjectSettingsResource{}
var _ resource.ResourceWithConfigure = &BYOCOpProjectSettingsResource{}

func NewBYOCOpProjectSettingsResource() resource.Resource {
	return &BYOCOpProjectSettingsResource{}
}

type BYOCOpProjectSettingsResource struct {
	store ByocOpProjectSettingsStore
}

func (r *BYOCOpProjectSettingsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_byoc_op_project_settings"
}

func (r *BYOCOpProjectSettingsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	nodeSchema := schema.SingleNestedAttribute{
		MarkdownDescription: "VM configuration",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"disk_size": schema.Int64Attribute{
				MarkdownDescription: "Disk size in GB",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"min_size": schema.Int64Attribute{
				MarkdownDescription: "Minimum number of instances",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_size": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of instances",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"desired_size": schema.Int64Attribute{
				MarkdownDescription: "Desired number of instances",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"instance_types": schema.StringAttribute{
				MarkdownDescription: "Instance type",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"capacity_type": schema.StringAttribute{
				MarkdownDescription: "Capacity type (ON_DEMAND or SPOT)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "BYOC Op Project Settings resource for managing project configurations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Settings identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_name": schema.StringAttribute{
				MarkdownDescription: "The name of the project",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data_plane_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the data plane",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "AWS region",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cloud_provider": schema.StringAttribute{
				MarkdownDescription: "Cloud provider",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_link_enabled": schema.BoolAttribute{
				MarkdownDescription: "Private link enabled",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"instances": schema.SingleNestedAttribute{
				MarkdownDescription: "Instance type configuration",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"core_vm": schema.StringAttribute{
						MarkdownDescription: "Core VM instance type",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"core_vm_min_count": schema.Int64Attribute{
						MarkdownDescription: "Core VM instance type",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(1),
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.RequiresReplace(),
						},
					},
					"fundamental_vm": schema.StringAttribute{
						MarkdownDescription: "Fundamental VM instance type",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"fundamental_vm_min_count": schema.Int64Attribute{
						MarkdownDescription: "Fundamental VM instance type",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(0),
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.RequiresReplace(),
						},
					},
					"search_vm": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"search_vm_min_count": schema.Int64Attribute{
						MarkdownDescription: "Search VM instance type",
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(0),
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.RequiresReplace(),
						},
					},
				},
			},
			"op_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Operation configuration settings",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"token": schema.StringAttribute{
						MarkdownDescription: "Operation token",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"agent_image_url": schema.StringAttribute{
						MarkdownDescription: "Agent image URL",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"node_quotas": schema.SingleNestedAttribute{
				Computed: true,
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

func (r *BYOCOpProjectSettingsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.store = &byocOpProjectSettingsStore{
		client: client,
	}
}

func (r *BYOCOpProjectSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BYOCOpProjectSettingsResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating BYOC Op Project Settings...")

	err := r.store.Create(ctx, &data, func(input *BYOCOpProjectSettingsResourceModel) error {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		if resp.Diagnostics.HasError() {
			return fmt.Errorf("failed to set state")
		}
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create BYOC Op Project Settings", err.Error())
		return
	}

	model, err := r.store.Describe(ctx, data.ProjectID.ValueString(), data.DataPlaneID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to describe BYOC Op Project Settings after creation", err.Error())
		return
	}

	// data.refresh(model)

	data.OpConfig = model.OpConfig
	data.NodeQuotas = model.NodeQuotas

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BYOCOpProjectSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BYOCOpProjectSettingsResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading BYOC Op Project Settings...")
	model, err := r.store.Describe(ctx, data.ProjectID.ValueString(), data.DataPlaneID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to describe BYOC Op Project Settings", err.Error())
		return
	}

	// Update the model with the latest data from the API
	// data.refresh(model)
	data.OpConfig = model.OpConfig
	data.NodeQuotas = model.NodeQuotas

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BYOCOpProjectSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BYOCOpProjectSettingsResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Updating BYOC Op Project Settings...")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BYOCOpProjectSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BYOCOpProjectSettingsResourceModel

	tflog.Info(ctx, "Deleting BYOC Op Project Settings...")
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.store.Delete(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete BYOC Op Project Settings", err.Error())
		return
	}
}
