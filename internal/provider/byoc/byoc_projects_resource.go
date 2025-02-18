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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
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
			"status": schema.Int64Attribute{
				MarkdownDescription: "The status of the BYOC project",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
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
							"fundamental_vm": schema.StringAttribute{
								MarkdownDescription: "Fundamental VM instance type",
								Required:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"search_vm": schema.StringAttribute{
								MarkdownDescription: "Search VM instance type",
								Required:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
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
	createTimeout, diags := data.Timeouts.Create(ctx, defaultBYOCProjectCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.store.Create(ctx, createTimeout, &data, func(project *BYOCProjectResourceModel) error {
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

	data.refresh(model)

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
	var data BYOCProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BYOCProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Delete BYOC Project...")
	var data BYOCProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, defaultBYOCProjectDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// throw error for now
	err := r.store.Delete(ctx, deleteTimeout, data.ID.ValueString(), data.DataPlaneID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete BYOC project", err.Error())
		return
	}
}
