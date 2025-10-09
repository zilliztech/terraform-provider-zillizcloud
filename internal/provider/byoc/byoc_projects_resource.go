// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package byoc

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
	zschema "github.com/zilliztech/terraform-provider-zillizcloud/internal/provider/schema"
)

const (
	TERRAFORM_DEPLOY_TYPE = 5
)

const (
	defaultBYOCProjectCreateTimeout time.Duration = 120 * time.Minute
	defaultBYOCProjectDeleteTimeout time.Duration = 120 * time.Minute
	defaultBYOCProjectUpdateTimeout time.Duration = 60 * time.Minute
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &BYOCProjectResource{}
var _ resource.ResourceWithConfigure = &BYOCProjectResource{}
var _ resource.ResourceWithValidateConfig = &BYOCProjectResource{}

func NewBYOCProjectResource() resource.Resource {
	return &BYOCProjectResource{}
}

// BYOCProjectResource defines the resource implementation.
type BYOCProjectResource struct {
	store ByocProjectStore
}

func (r *BYOCProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_byoc_project"
}

func (r *BYOCProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "BYOC Project resource for managing bring-your-own-cloud projects.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Project identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"data_plane_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the data plane",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the BYOC project",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the BYOC project, possible values are RUNNING, STOPPED",
				Required:            true,
			},
			"aws": schema.SingleNestedAttribute{
				MarkdownDescription: "AWS configuration for the BYOC project",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"region": schema.StringAttribute{
						MarkdownDescription: "AWS region",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},

					"network": schema.SingleNestedAttribute{
						MarkdownDescription: "Network configuration",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"vpc_id": schema.StringAttribute{
								MarkdownDescription: "VPC ID",
								Required:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"subnet_ids": schema.SetAttribute{
								MarkdownDescription: "List of subnet IDs",
								Required:            true,
								ElementType:         types.StringType,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.RequiresReplace(),
								},
							},
							"security_group_ids": schema.SetAttribute{
								MarkdownDescription: "List of security group IDs",
								Required:            true,
								ElementType:         types.StringType,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.RequiresReplace(),
								},
							},
							"vpc_endpoint_id": schema.StringAttribute{
								MarkdownDescription: "VPC endpoint ID",
								Optional:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
					"role_arn": schema.SingleNestedAttribute{
						MarkdownDescription: "Role ARN configuration",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"storage": schema.StringAttribute{
								MarkdownDescription: "Storage role ARN",
								Required:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"eks": schema.StringAttribute{
								MarkdownDescription: "EKS role ARN",
								Required:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"cross_account": schema.StringAttribute{
								MarkdownDescription: "Cross account role ARN",
								Required:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
					"storage": schema.SingleNestedAttribute{
						MarkdownDescription: "Storage configuration",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"bucket_id": schema.StringAttribute{
								MarkdownDescription: "Storage bucket ID",
								Required:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
			"instances": zschema.Instances,
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx,
				timeouts.Opts{
					Create: true,
					CreateDescription: `Timeout defaults to 120 mins. Accepts a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) ` +
						`consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are ` +
						`"s" (seconds), "m" (minutes), "h" (hours).`,
					Update: true,
					UpdateDescription: `Timeout defaults to 60 mins. Accepts a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) ` +
						`consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are ` +
						`"s" (seconds), "m" (minutes), "h" (hours).`,
					Delete: true,
					DeleteDescription: `Timeout defaults to 120 mins. Accepts a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) ` +
						`consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are ` +
						`"s" (seconds), "m" (minutes), "h" (hours).`,
				},
			),
		},
	}
}

func (r *BYOCProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.store = &byocProjectStore{client: client}
}

func (r *BYOCProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var data BYOCProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, fmt.Sprintf("Create BYOC Project request: %+v", data))

	err := r.store.Create(ctx, &data, func(project *BYOCProjectResourceModel) error {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		if resp.Diagnostics.HasError() {
			return fmt.Errorf("failed to set state")
		}
		return nil
	})

	if err != nil {
		resp.Diagnostics.AddError("Failed to create BYOC project", err.Error())
		return
	}

}

func (r *BYOCProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Read BYOC Project...")
	var data BYOCProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, err := r.store.Describe(ctx, data.ID.ValueString(), data.DataPlaneID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read BYOC project", err.Error())
		return
	}

	//BUG: get the auto scaling and arch from the local state, since they are not returned by the API
	model.Instances.AutoScaling = data.Instances.AutoScaling
	model.Instances.Arch = data.Instances.Arch

	data.AWS = model.AWS
	data.Instances = model.Instances
	data.Status = model.Status
	data.DataPlaneID = model.DataPlaneID

	tflog.Info(ctx, fmt.Sprintf("Read BYOC Project response: %+v", data))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Convert Go slice to Terraform Set.
func sliceToTerraformSet(input []string) (types.Set, diag.Diagnostics) {
	// Create a slice of `attr.Value` for each string
	elements := make([]attr.Value, len(input))
	for i, v := range input {
		elements[i] = types.StringValue(v) // Convert each string to Terraform's `types.String`
	}

	// Create the SetValue from the element slice
	set, diags := types.SetValue(types.StringType, elements)
	return set, diags
}

func (r *BYOCProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Update BYOC Project...")

	var currentData BYOCProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &currentData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plannedData BYOCProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plannedData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("Update BYOC Project current state: %+v", currentData))
	tflog.Info(ctx, fmt.Sprintf("Update BYOC Project planned state: %+v", plannedData))

	currentStatus := currentData.Status.ValueString()
	plannedStatus := plannedData.Status.ValueString()

	// Check if only status changed
	if currentStatus != plannedStatus {
		// Status changed from STOPPED to RUNNING - Resume
		if currentStatus == BYOCProjectStatusStopped.String() && plannedStatus == BYOCProjectStatusRunning.String() {
			tflog.Info(ctx, "Resume BYOC Project...")
			err := r.store.Resume(ctx, &plannedData)
			if err != nil {
				resp.Diagnostics.AddError("Failed to resume BYOC project", err.Error())
				return
			}
		} else if currentStatus == BYOCProjectStatusRunning.String() && plannedStatus == BYOCProjectStatusStopped.String() {
			// Status changed from RUNNING to STOPPED - Suspend
			tflog.Info(ctx, "Suspend BYOC Project...")
			err := r.store.Suspend(ctx, &plannedData)
			if err != nil {
				resp.Diagnostics.AddError("Failed to suspend BYOC project", err.Error())
				return
			}
		} else {
			resp.Diagnostics.AddError(
				"Invalid BYOC Project Status Change",
				fmt.Sprintf("Invalid status change from %s to %s. Expected transitions are %s to %s or %s to %s",
					currentStatus, plannedStatus,
					BYOCProjectStatusStopped.String(), BYOCProjectStatusRunning.String(),
					BYOCProjectStatusRunning.String(), BYOCProjectStatusStopped.String()),
			)
			return
		}
	} else {
		// Any other changes - invoke create method
		tflog.Info(ctx, "Non-status changes detected, invoking create method...")
		// print the planned data
		// tflog.Info(ctx, fmt.Sprintf("Planned data: %+v", plannedData))
		err := r.store.Create(ctx, &plannedData, func(project *BYOCProjectResourceModel) error {
			resp.Diagnostics.Append(resp.State.Set(ctx, &plannedData)...)
			if resp.Diagnostics.HasError() {
				return fmt.Errorf("failed to set state")
			}
			return nil
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to update BYOC project", err.Error())
			return
		}
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plannedData)...)
}

func (r *BYOCProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Delete BYOC Project...")
	var data BYOCProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.store.Delete(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete BYOC project", err.Error())
		return
	}
}

func (r *BYOCProjectResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data BYOCProjectResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate fundamental VM configuration
	if !data.Instances.Fundamental.MinCount.IsNull() && !data.Instances.Fundamental.MaxCount.IsNull() {
		minCount := data.Instances.Fundamental.MinCount.ValueInt64()
		maxCount := data.Instances.Fundamental.MaxCount.ValueInt64()
		if minCount > maxCount {
			resp.Diagnostics.AddAttributeError(
				path.Root("instances").AtName("fundamental").AtName("min_count"),
				"Invalid min_count",
				fmt.Sprintf("min_count (%d) must be less than or equal to max_count (%d)", minCount, maxCount),
			)
		}
	}

	// Validate search VM configuration
	if !data.Instances.Search.MinCount.IsNull() && !data.Instances.Search.MaxCount.IsNull() {
		minCount := data.Instances.Search.MinCount.ValueInt64()
		maxCount := data.Instances.Search.MaxCount.ValueInt64()
		if minCount > maxCount {
			resp.Diagnostics.AddAttributeError(
				path.Root("instances").AtName("search").AtName("min_count"),
				"Invalid min_count",
				fmt.Sprintf("min_count (%d) must be less than or equal to max_count (%d)", minCount, maxCount),
			)
		}
	}

	// Validate index VM configuration
	if !data.Instances.Index.MinCount.IsNull() && !data.Instances.Index.MaxCount.IsNull() {
		minCount := data.Instances.Index.MinCount.ValueInt64()
		maxCount := data.Instances.Index.MaxCount.ValueInt64()
		if minCount > maxCount {
			resp.Diagnostics.AddAttributeError(
				path.Root("instances").AtName("index").AtName("min_count"),
				"Invalid min_count",
				fmt.Sprintf("min_count (%d) must be less than or equal to max_count (%d)", minCount, maxCount),
			)
		}
	}
}
