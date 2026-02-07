package byoc_op

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

const (
	defaultBYOCOpProjectCreateTimeout time.Duration = 120 * time.Minute
	defaultBYOCOpProjectDeleteTimeout time.Duration = 120 * time.Minute
	defaultBYOCOpProjectUpdateTimeout time.Duration = 60 * time.Minute
)

var _ resource.Resource = &BYOCOpProjectResource{}
var _ resource.ResourceWithConfigure = &BYOCOpProjectResource{}
var _ resource.ResourceWithConfigValidators = &BYOCOpProjectResource{}

func NewBYOCOpProjectResource() resource.Resource {
	return &BYOCOpProjectResource{}
}

type BYOCOpProjectResource struct {
	store ByocOpProjectStore
}

func (r *BYOCOpProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_byoc_i_project"
}

func (r *BYOCOpProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "BYOC-I Project resource for managing bring-your-own-cloud operator projects.",
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
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ext_config": schema.StringAttribute{
				MarkdownDescription: "External configuration",
				Optional:            true,
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
					// CSE (Client-Side Encryption) configuration for AWS KMS encryption
					// This enables users to encrypt their data using their own KMS keys
					"cse": schema.SingleNestedAttribute{
						MarkdownDescription: "CSE (Client-Side Encryption) configuration for AWS KMS encryption",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							// The IAM role ARN that has permissions to use the KMS key for encryption/decryption
							"aws_cse_role_arn": schema.StringAttribute{
								MarkdownDescription: "AWS IAM role ARN for client-side encryption operations",
								Optional:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							// The default KMS key ARN used for encrypting data
							"default_aws_cse_key_arn": schema.StringAttribute{
								MarkdownDescription: "Default AWS KMS key ARN for client-side encryption",
								Optional:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							// External ID for cross-account KMS key access (used in IAM role trust policy)
							"external_id": schema.StringAttribute{
								MarkdownDescription: "External ID for cross-account KMS key access",
								Optional:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
			"azure": schema.SingleNestedAttribute{
				MarkdownDescription: "Azure configuration for the BYOC project",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"region": schema.StringAttribute{
						MarkdownDescription: "Azure region",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"network": schema.SingleNestedAttribute{
						MarkdownDescription: "Network configuration",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"vnet_id": schema.StringAttribute{
								MarkdownDescription: "virtual network ID",
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
							"nsg_ids": schema.SetAttribute{
								MarkdownDescription: "List of network security group IDs",
								Required:            true,
								ElementType:         types.StringType,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.RequiresReplace(),
								},
							},
							"private_endpoint_id": schema.StringAttribute{
								MarkdownDescription: "Private endpoint ID",
								Optional:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
					"identity": schema.SingleNestedAttribute{
						MarkdownDescription: "Identity configuration",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"storages": schema.SetNestedAttribute{
								MarkdownDescription: "Storage identity configuration (exactly 10 items required)",
								Required:            true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"client_id": schema.StringAttribute{
											MarkdownDescription: "Client ID of the managed identity",
											Required:            true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.RequiresReplace(),
											},
										},
										"principal_id": schema.StringAttribute{
											MarkdownDescription: "Principal ID of the managed identity",
											Required:            true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.RequiresReplace(),
											},
										},
										"resource_id": schema.StringAttribute{
											MarkdownDescription: "Resource ID of the managed identity",
											Required:            true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.RequiresReplace(),
											},
										},
									},
								},
								Validators: []validator.Set{
									setvalidator.SizeBetween(10, 10),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.RequiresReplace(),
								},
							},
							"kubelet": schema.SingleNestedAttribute{
								MarkdownDescription: "Kubelet identity configuration",
								Required:            true,
								Attributes: map[string]schema.Attribute{
									"client_id": schema.StringAttribute{
										MarkdownDescription: "Client ID",
										Required:            true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"principal_id": schema.StringAttribute{
										MarkdownDescription: "Principal ID",
										Required:            true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"resource_id": schema.StringAttribute{
										MarkdownDescription: "Resource ID",
										Required:            true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
							"maintenance": schema.SingleNestedAttribute{
								MarkdownDescription: "Maintenance identity configuration",
								Required:            true,
								Attributes: map[string]schema.Attribute{
									"client_id": schema.StringAttribute{
										MarkdownDescription: "Client ID",
										Required:            true,
									},
									"principal_id": schema.StringAttribute{
										MarkdownDescription: "Principal ID",
										Required:            true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"resource_id": schema.StringAttribute{
										MarkdownDescription: "Resource ID",
										Required:            true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
					},
					"storage": schema.SingleNestedAttribute{
						MarkdownDescription: "Storage configuration",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"storage_account_name": schema.StringAttribute{
								MarkdownDescription: "Storage account name",
								Required:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"container_name": schema.StringAttribute{
								MarkdownDescription: "Storage container name",
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
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
				Update: true,
			}),
		},
	}
}

func (r *BYOCOpProjectResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("aws"),
			path.MatchRoot("azure"),
		),
	}
}

func (r *BYOCOpProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.store = &byocOpProjectStore{client: client}
}

func (r *BYOCOpProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BYOCOpProjectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating BYOC-I Project...")

	err := r.store.Create(ctx, &data, func(project *BYOCOpProjectResourceModel) error {
		resp.Diagnostics.Append(resp.State.Set(ctx, project)...)
		if resp.Diagnostics.HasError() {
			return fmt.Errorf("failed to set state")
		}
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create BYOC-I project, got error: %s", err))
		return
	}

	tflog.Info(ctx, "Created BYOC-I Project")
}

func (r *BYOCOpProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BYOCOpProjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading BYOC-I Project...")

	project, err := r.store.Describe(ctx, data.ProjectID.ValueString(), data.DataPlaneID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read BYOC-I project, got error: %s", err))
		return
	}

	// Overwrite items with refreshed state
	data.refresh(project)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BYOCOpProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BYOCOpProjectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Most attributes require replacement, so this is mainly for handling timeouts
	var state BYOCOpProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Status = state.Status
	data.ID = state.ID
	data.DataPlaneID = state.DataPlaneID
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BYOCOpProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BYOCOpProjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting BYOC-I Project...")

	err := r.store.Delete(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete BYOC-I project, got error: %s", err))
		return
	}

	tflog.Info(ctx, "Deleted BYOC-I Project")
}
