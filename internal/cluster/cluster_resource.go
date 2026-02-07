package cluster

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
	util "github.com/zilliztech/terraform-provider-zillizcloud/client/retry"
	customvalidator "github.com/zilliztech/terraform-provider-zillizcloud/internal/validator"
)

const (
	defaultClusterCreateTimeout time.Duration = 45 * time.Minute
	defaultClusterUpdateTimeout time.Duration = 30 * time.Minute

	FreePlan             string = "Free"
	ServerlessPlan       string = "Serverless"
	StandardPlan         string = "Standard"
	EnterprisePlan       string = "Enterprise"
	BusinessCriticalPlan string = "BusinessCritical"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ClusterResource{}
var _ resource.ResourceWithConfigure = &ClusterResource{}
var _ resource.ResourceWithImportState = &ClusterResource{}

func NewClusterResource() resource.Resource {
	return &ClusterResource{}
}

// ClusterResource defines the resource implementation.
type ClusterResource struct {
	client  *zilliz.Client
	store   ClusterStore
	timeout func() time.Duration
}

func (r *ClusterResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("cu_size"),
			path.MatchRoot("cu_settings"),
		),
		resourcevalidator.Conflicting(
			path.MatchRoot("replica"),
			path.MatchRoot("replica_settings"),
		),
	}
}

func (r *ClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *ClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Cluster resource. If 'plan', 'cu_size' and 'cu_type' are not specified, then a free cluster is created.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Cluster identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_name": schema.StringAttribute{
				MarkdownDescription: "The name of the cluster to be created. It is a string of no more than 32 characters.",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project where the cluster is to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"plan": schema.StringAttribute{
				MarkdownDescription: "The plan tier of the Zilliz Cloud service. Available options are Serverless, Standard and Enterprise.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(FreePlan, ServerlessPlan, StandardPlan, EnterprisePlan, BusinessCriticalPlan),
				},
			},
			"cu_size": schema.Int64Attribute{
				MarkdownDescription: "The size of the CU to be used for the created cluster. It is an integer from 1 to 256.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"cu_type": schema.StringAttribute{
				MarkdownDescription: `The type of the CU used for the Zilliz Cloud cluster to be created. Available options are Performance-optimized, Capacity-optimized and Tiered-storage.`,
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Performance-optimized"),
				Validators: []validator.String{
					stringvalidator.OneOf("Performance-optimized", "Capacity-optimized", "Tiered-storage", "Extended-capacity"),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The name of the cluster user generated by default.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password of the cluster user generated by default. It will not be displayed again, so note it down and securely store it.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"prompt": schema.StringAttribute{
				MarkdownDescription: "The statement indicating that this operation succeeds.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "An optional description about the cluster.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the region where the cluster exists.",
				Optional:            true,
				Computed:            true,
			},
			"desired_status": schema.StringAttribute{
				MarkdownDescription: "The desired status of the cluster. Possible values are RUNNING and SUSPENDED. Defaults to RUNNING.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("RUNNING"),
				Validators: []validator.String{
					stringvalidator.OneOf("RUNNING", "SUSPENDED"),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the cluster. Possible values are RUNNING, SUSPENDING, SUSPENDED, and RESUMING.",
				Computed:            true,
				Optional:            true,
			},
			"connect_address": schema.StringAttribute{
				MarkdownDescription: "The public endpoint of the cluster. You can connect to the cluster using this endpoint from the public network.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"private_link_address": schema.StringAttribute{
				MarkdownDescription: "The private endpoint of the cluster. You can set up a private link to allow your VPS in the same cloud region to access your cluster.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"create_time": schema.StringAttribute{
				MarkdownDescription: "The time at which the cluster has been created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: "A map of labels to assign to the cluster. Labels are key-value pairs that can be used to organize and categorize clusters.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default: mapdefault.StaticValue(
					types.MapValueMust(types.StringType, map[string]attr.Value{}),
				),
				Validators: []validator.Map{
					customvalidator.K8sLabelMapValidator{},
				},
			},
			"replica": schema.Int64Attribute{
				MarkdownDescription: "The number of replicas for the cluster. Defaults to 1.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
					customvalidator.ReplicaCuSizeValidator{},
				},
			},
			"load_balancer_security_groups": schema.SetAttribute{
				MarkdownDescription: "A set of security group IDs to associate with the load balancer of the cluster.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             setdefault.StaticValue(types.SetNull(types.StringType)),
				DeprecationMessage:  "This field is deprecated. Use the zillizcloud_cluster_load_balancer_security_groups resource instead.",
			},
			"cu_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Query compute unit configuration for the cluster. The cu_settings and cu_size cannot be set simultaneously.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"dynamic_scaling": schema.SingleNestedAttribute{
						MarkdownDescription: "Dynamic scaling configuration for query CUs.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"min": schema.Int64Attribute{
								MarkdownDescription: "Minimum number of compute units (CU) for autoscaling. Must be at least 1.",
								Optional:            true,
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
							"max": schema.Int64Attribute{
								MarkdownDescription: "Maximum number of compute units (CU) for autoscaling. Must be greater than or equal to min.",
								Optional:            true,
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
						},
					},
					"schedule_scaling": schema.ListNestedAttribute{
						MarkdownDescription: "Scheduled scaling configuration for query CUs. Allows you to schedule CU scaling at specific times using cron expressions.",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"timezone": schema.StringAttribute{
									MarkdownDescription: "The timezone for the cron expression. Defaults to Etc/UTC.",
									Optional:            true,
									Computed:            true,
									Default:             stringdefault.StaticString("Etc/UTC"),
								},
								"cron": schema.StringAttribute{
									MarkdownDescription: "Cron expression defining when the scheduled scaling should occur.",
									Required:            true,
								},
								"target": schema.Int64Attribute{
									MarkdownDescription: "Target number of compute units (CU) for the scheduled scaling. Must be at least 1.",
									Required:            true,
									Validators: []validator.Int64{
										int64validator.AtLeast(1),
									},
								},
							},
						},
					},
				},
			},
			"replica_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Query compute unit configuration for the cluster. The cu_settings and cu_size cannot be set simultaneously.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"dynamic_scaling": schema.SingleNestedAttribute{
						MarkdownDescription: "Dynamic scaling configuration for query CUs.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"min": schema.Int64Attribute{
								MarkdownDescription: "Minimum number of compute units (CU) for autoscaling. Must be at least 1.",
								Optional:            true,
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
							"max": schema.Int64Attribute{
								MarkdownDescription: "Maximum number of compute units (CU) for autoscaling. Must be greater than or equal to min.",
								Optional:            true,
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
						},
					},
					"schedule_scaling": schema.ListNestedAttribute{
						MarkdownDescription: "Scheduled scaling configuration for query CUs. Allows you to schedule CU scaling at specific times using cron expressions.",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"timezone": schema.StringAttribute{
									MarkdownDescription: "The timezone for the cron expression. Defaults to Etc/UTC.",
									Optional:            true,
									Computed:            true,
									Default:             stringdefault.StaticString("Etc/UTC"),
								},
								"cron": schema.StringAttribute{
									MarkdownDescription: "Cron expression defining when the scheduled scaling should occur.",
									Required:            true,
								},
								"target": schema.Int64Attribute{
									MarkdownDescription: "Target number of compute units (CU) for the scheduled scaling. Must be at least 1.",
									Required:            true,
									Validators: []validator.Int64{
										int64validator.AtLeast(1),
									},
								},
							},
						},
					},
				},
			},
			"aws_cse_key_arn": schema.StringAttribute{
				MarkdownDescription: "The ARN of the AWS KMS key used for client-side encryption (CSE). Only used for BYOC clusters. Immutable after creation.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bucket_info": schema.SingleNestedAttribute{
				MarkdownDescription: "Bucket information for the cluster. Only used for BYOC clusters.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"bucket_name": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The name of the bucket to be used for the cluster.",
					},
					"prefix": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The prefix within the bucket to use for the cluster. If not provided, the cluster will use the bucket's root directory.",
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx,
				timeouts.Opts{
					Create: true,
					CreateDescription: `Timeout defaults to 45 mins. Accepts a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) ` +
						`consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are ` +
						`"s" (seconds), "m" (minutes), "h" (hours).`,
					Update: true,
					UpdateDescription: `Timeout defaults to 30 mins. Accepts a string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) ` +
						`consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are ` +
						`"s" (seconds), "m" (minutes), "h" (hours).`,
				},
			),
		},
	}
}

func (r *ClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*zilliz.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
	r.store = &ClusterStoreImpl{client: client}
}

func CloneClient(ctx context.Context, client *zilliz.Client, data *ClusterResourceModel) (*zilliz.Client, error) {
	var regionId = client.RegionId

	if data.RegionId.ValueString() != "" {
		regionId = data.RegionId.ValueString()
	}

	ctx = tflog.SetField(ctx, "RegionID", regionId)
	tflog.Info(ctx, "Clone Client...")

	return client.Clone(zilliz.WithCloudRegionId(regionId))
}

func convertTerraformMapToStringMap(terraformMap types.Map) map[string]string {
	labels := make(map[string]string)
	if !terraformMap.IsNull() && !terraformMap.IsUnknown() {
		elements := terraformMap.Elements()
		for k, v := range elements {
			if strValue, ok := v.(types.String); ok {
				labels[k] = strValue.ValueString()
			}
		}
	}
	return labels
}

func (r *ClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Create Cluster...")
	tfPlan := ClusterResourceModel{}

	createTimeout, diags := tfPlan.Timeouts.Create(ctx, defaultClusterCreateTimeout)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	diags = req.Plan.Get(ctx, &tfPlan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Validate replica value during create - must be 1
	if !tfPlan.Replica.IsNull() && !tfPlan.Replica.IsUnknown() && tfPlan.Replica.ValueInt64() != 1 {
		resp.Diagnostics.AddError("Invalid replica value for cluster creation", "Replica value must be 1 during cluster creation")
		return
	}

	// Validate cu_settings is not set during create
	if tfPlan.CuSettings != nil {
		resp.Diagnostics.AddError("Invalid cu_settings for cluster creation", "cu_settings cannot be set during cluster creation. Please create the cluster first, then update it to configure cu_settings.")
		return
	}

	// Validate replica_settings is not set during create
	if tfPlan.ReplicaSettings != nil {
		resp.Diagnostics.AddError("Invalid replica_settings for cluster creation", "replica_settings cannot be set during cluster creation. Please create the cluster first, then update it to configure replica_settings.")
		return
	}

	tfState := tfPlan

	tfState.completeForFreeOrServerless(&tfPlan)
	tfState.setUnknown()

	newState, err := r.store.Create(ctx, &tfPlan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cluster", err.Error())
		return
	}
	tfState.ClusterId = newState.ClusterId
	tfState.Username = newState.Username
	tfState.Password = newState.Password
	tfState.Prompt = newState.Prompt
	tfState.CuSize = types.Int64Value(1)
	tfState.Plan = types.StringValue("unknown")

	resp.Diagnostics.Append(resp.State.Set(ctx, tfState)...)

	r.tryBestUpdateStatesAfterCreation(ctx, &tfPlan, &tfState, resp)
}

// tryBestUpdateStatesAfterCreation is called after resource already created, so there're just warnings if anything wrong.
func (r *ClusterResource) tryBestUpdateStatesAfterCreation(ctx context.Context, plan, state *ClusterResourceModel, resp *resource.CreateResponse) {
	newState, isRunning := r.getStateAndWaitForRunning(ctx, state.ClusterId.ValueString())
	if !isRunning {
		resp.Diagnostics.AddWarning("Cluster created but not in RUNNING state", "The cluster was created successfully, but it is not in the RUNNING state after waiting for the specified timeout. Please check the Zilliz Cloud console for more details or contact support.")
	}
	if newState != nil {
		state.Status = newState.Status
		state.ConnectAddress = newState.ConnectAddress
		state.PrivateLinkAddress = newState.PrivateLinkAddress
		state.CreateTime = newState.CreateTime
		state.Description = newState.Description
		state.RegionId = newState.RegionId
		state.Plan = newState.Plan
	}

	diags := resp.State.Set(ctx, state)
	if diags.HasError() {
		errorToWarning(resp, diags)
		return
	}

	if len(plan.SecurityGroups.Elements()) != 0 {
		err := r.handleSecurityGroupsUpdate(ctx, *plan, *state)
		if err != nil {
			resp.Diagnostics.AddWarning("Failed to update cluster security groups", err.Error())
			return
		}
		state.SecurityGroups = plan.SecurityGroups
	}

	diags = resp.State.Set(ctx, state)
	if diags.HasError() {
		errorToWarning(resp, diags)
	}
}

func errorToWarning(resp *resource.CreateResponse, diagnostics diag.Diagnostics) {
	for _, diag := range diagnostics.Errors() {
		resp.Diagnostics.AddWarning(diag.Summary(), diag.Detail())
	}
}

func (r *ClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Read Cluster...")
	var state ClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.store.Get(ctx, state.ClusterId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get cluster", err.Error())
		return
	}

	state.ClusterId = cluster.ClusterId
	state.ClusterName = cluster.ClusterName
	state.ProjectId = cluster.ProjectId
	state.RegionId = cluster.RegionId
	state.Description = cluster.Description
	state.Status = cluster.Status
	state.ConnectAddress = cluster.ConnectAddress
	state.PrivateLinkAddress = cluster.PrivateLinkAddress
	state.CreateTime = cluster.CreateTime
	state.Plan = cluster.Plan
	state.Replica = cluster.Replica
	state.CuSize = cluster.CuSize
	state.CuType = cluster.CuType

	if state.DesiredStatus.IsNull() {
		state.DesiredStatus = cluster.Status
	}

	state.completeForFreeOrServerless(cluster)

	labels, err := r.store.GetLabels(ctx, state.ClusterId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get cluster labels", err.Error())
		return
	}
	state.Labels = labels

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ClusterResource) handleCuSizeUpdate(ctx context.Context, plan, state ClusterResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	err := r.store.UpgradeCuSize(ctx, state.ClusterId.ValueString(), int(plan.CuSize.ValueInt64()))
	if err != nil {
		diags.AddError("Failed to modify cluster", err.Error())
		return diags
	}

	err = r.waitForStatus(ctx, r.timeout(), state.ClusterId.ValueString(), "RUNNING")
	if err != nil && !util.IsNetworkGiveUpError(err) {
		diags.AddError("Failed to wait for cluster to enter RUNNING state", err.Error())
	}
	return diags
}

func (r *ClusterResource) handleReplicaUpdate(ctx context.Context, plan, state ClusterResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	err := r.store.ModifyReplica(ctx, state.ClusterId.ValueString(), int(plan.Replica.ValueInt64()))
	if err != nil {
		diags.AddError("Failed to modify cluster replica", err.Error())
		return diags
	}

	err = r.waitForStatus(ctx, r.timeout(), state.ClusterId.ValueString(), "RUNNING")
	if err != nil && !util.IsNetworkGiveUpError(err) {
		diags.AddError("Failed to wait for cluster to enter RUNNING state", err.Error())
	}
	return diags
}

func (r *ClusterResource) handleStatusUpdate(ctx context.Context, plan, state ClusterResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if !plan.isStatusChangeRequired(state) {
		return diags
	}

	action := plan.getStatusAction(state)
	var err error
	var targetStatus string

	switch action {
	case StatusActionSuspend:
		err = r.store.SuspendCluster(ctx, state.ClusterId.ValueString())
		targetStatus = "SUSPENDED"
		if err != nil {
			diags.AddError("Failed to suspend cluster", err.Error())
			return diags
		}
	case StatusActionResume:
		err = r.store.ResumeCluster(ctx, state.ClusterId.ValueString())
		targetStatus = "RUNNING"
		if err != nil {
			diags.AddError("Failed to resume cluster", err.Error())
			return diags
		}
	case StatusActionNone:
		return diags
	}

	err = r.waitForStatus(ctx, r.timeout(), state.ClusterId.ValueString(), targetStatus)
	if err != nil && !util.IsNetworkGiveUpError(err) {
		diags.AddError("Failed to wait for cluster to enter desired state", err.Error())
	}
	return diags
}

func (r *ClusterResource) handleLabelsUpdate(ctx context.Context, plan, state ClusterResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	labels := convertTerraformMapToStringMap(plan.Labels)

	err := r.store.UpdateLabels(ctx, state.ClusterId.ValueString(), labels)
	if err != nil {
		diags.AddError("Failed to update cluster labels", err.Error())
		return diags
	}

	return diags
}

func (r *ClusterResource) handleClusterNameUpdate(ctx context.Context, plan, state ClusterResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	err := r.store.ModifyClusterProperties(ctx, state.ClusterId.ValueString(), plan.ClusterName.ValueString())
	if err != nil {
		diags.AddError("Failed to modify cluster name", err.Error())
		return diags
	}

	return diags
}

func (r *ClusterResource) handleSecurityGroupsUpdate(ctx context.Context, plan, state ClusterResourceModel) error {
	// Convert Terraform set to Go slice
	var securityGroupIds []string
	elements := plan.SecurityGroups.Elements()
	for _, elem := range elements {
		if strValue, ok := elem.(types.String); ok {
			securityGroupIds = append(securityGroupIds, strValue.ValueString())
		}
	}
	err := r.store.UpsertSecurityGroups(ctx, state.ClusterId.ValueString(), securityGroupIds)
	if err != nil {
		return fmt.Errorf("failed to update cluster security groups: %w", err)
	}

	return nil
}

func (r *ClusterResource) handleCuSettingsUpdate(ctx context.Context, plan, state ClusterResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// can not set both dynamic scaling and schedule scaling at the same time
	if plan.CuSettings != nil && !plan.CuSettings.IsdynamicScalingNull() && !plan.CuSettings.IsSchedulesNull() {
		diags.AddError("Invalid configuration", "Cannot set both dynamic scaling and schedule scaling at the same time")
		return diags
	}

	if plan.CuSettings != nil && !plan.CuSettings.IsdynamicScalingNull() {
		minCU := int(plan.CuSettings.DynamicScaling.Min.ValueInt64())
		maxCU := int(plan.CuSettings.DynamicScaling.Max.ValueInt64())

		if minCU > maxCU {
			diags.AddError("Invalid autoscaling configuration", fmt.Sprintf("Minimum CU (%d) must be less than or equal to maximum CU (%d)", minCU, maxCU))
			return diags
		}

		err := r.store.ModifyAutoscaling(ctx, state.ClusterId.ValueString(), minCU, maxCU)
		if err != nil {
			diags.AddError("Failed to modify cluster autoscaling", err.Error())
			return diags
		}
	}

	if plan.CuSettings != nil && !plan.CuSettings.IsSchedulesNull() {
		schedules := make([]zilliz.ScheduleConfig, len(plan.CuSettings.ScheduleScaling))
		for i, s := range plan.CuSettings.ScheduleScaling {
			schedules[i] = zilliz.ScheduleConfig{
				Timezone: s.Timezone.ValueString(),
				Cron:     s.Cron.ValueString(),
				Target:   int(s.Target.ValueInt64()),
			}
		}

		err := r.store.ModifySchedules(ctx, state.ClusterId.ValueString(), schedules)
		if err != nil {
			diags.AddError("Failed to modify cluster schedules", err.Error())
			return diags
		}
	}

	return diags
}

func (r *ClusterResource) handleReplicaSettingsUpdate(ctx context.Context, plan, state ClusterResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// can not set both dynamic scaling and schedule scaling at the same time
	if plan.ReplicaSettings != nil && !plan.ReplicaSettings.IsdynamicScalingNull() && !plan.ReplicaSettings.IsSchedulesNull() {
		diags.AddError("Invalid configuration", "Cannot set both dynamic scaling and schedule scaling at the same time")
		return diags
	}

	if plan.ReplicaSettings != nil && !plan.ReplicaSettings.IsdynamicScalingNull() {
		minReplica := int(plan.ReplicaSettings.DynamicScaling.Min.ValueInt64())
		maxReplica := int(plan.ReplicaSettings.DynamicScaling.Max.ValueInt64())

		if minReplica > maxReplica {
			diags.AddError("Invalid autoscaling configuration", fmt.Sprintf("Minimum replica (%d) must be less than or equal to maximum replica (%d)", minReplica, maxReplica))
			return diags
		}

		err := r.store.ModifyReplicaAutoscaling(ctx, state.ClusterId.ValueString(), minReplica, maxReplica)
		if err != nil {
			diags.AddError("Failed to modify cluster autoscaling", err.Error())
		}
	}

	if plan.ReplicaSettings != nil && !plan.ReplicaSettings.IsSchedulesNull() {
		schedules := make([]zilliz.ScheduleConfig, len(plan.ReplicaSettings.ScheduleScaling))
		for i, s := range plan.ReplicaSettings.ScheduleScaling {
			schedules[i] = zilliz.ScheduleConfig{
				Timezone: s.Timezone.ValueString(),
				Cron:     s.Cron.ValueString(),
				Target:   int(s.Target.ValueInt64()),
			}
		}

		err := r.store.ModifyReplicaSchedules(ctx, state.ClusterId.ValueString(), schedules)
		if err != nil {
			diags.AddError("Failed to modify cluster schedules", err.Error())
			return diags
		}
	}

	return diags
}

func (r *ClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Update Cluster...")

	var plan ClusterResourceModel
	var state ClusterResourceModel

	updateTimeout, timeoutDiags := plan.Timeouts.Update(ctx, defaultClusterUpdateTimeout)
	resp.Diagnostics.Append(timeoutDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.timeout = func() time.Duration {
		return updateTimeout
	}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that cu_size and replica are not changed at the same time
	if plan.isCuSizeChanged(state) && plan.isReplicaChanged(state) {
		resp.Diagnostics.AddError("Invalid configuration change", "Cannot change cu_size and replica at the same time. Please update them in separate operations.")
		return
	}

	if plan.isCuSizeChanged(state) && plan.isCuSettingsDisabled() {
		resp.Diagnostics.Append(r.handleCuSizeUpdate(ctx, plan, state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if plan.isReplicaChanged(state) && plan.isReplicaSettingsDisabled() {
		resp.Diagnostics.Append(r.handleReplicaUpdate(ctx, plan, state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(r.handleStatusUpdate(ctx, plan, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.isLabelsChanged(state) {
		resp.Diagnostics.Append(r.handleLabelsUpdate(ctx, plan, state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Labels = plan.Labels
	}

	if plan.isClusterNameChanged(state) {
		resp.Diagnostics.Append(r.handleClusterNameUpdate(ctx, plan, state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if plan.isCuSettingsChanged(state) {
		resp.Diagnostics.Append(r.handleCuSettingsUpdate(ctx, plan, state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if plan.isReplicaSettingsChanged(state) {
		resp.Diagnostics.Append(r.handleReplicaSettingsUpdate(ctx, plan, state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if plan.isBucketInfoChanged(state) {
		resp.Diagnostics.AddError("Invalid configuration change", "Cannot change bucket info after cluster is created")
		return
	}

	if plan.isAwsCseKeyArnChanged(state) {
		resp.Diagnostics.AddError("Invalid configuration change", "Cannot change AWS CSE key ARN after cluster is created")
		return
	}

	cluster, err := r.store.Get(ctx, state.ClusterId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get cluster", err.Error())
		return
	}

	state.populate(cluster)

	state.CuSettings = plan.CuSettings
	state.ReplicaSettings = plan.ReplicaSettings
	state.Timeouts = plan.Timeouts

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Delete Cluster...")
	var data ClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.store.Delete(ctx, data.ClusterId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to drop cluster", err.Error())
		return
	}
}

func (r *ClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: clusterId,regionId. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region_id"), idParts[1])...)
}

func (r *ClusterResource) waitForStatus(ctx context.Context, timeout time.Duration, clusterId string, status string) error {
	_, err := util.NetworkResilientPoll(ctx, timeout, func() (*string, *util.Err) {
		cluster, err := r.client.DescribeCluster(clusterId)
		if err != nil {
			// Allow network errors to be retried, other errors are non-retryable
			return nil, &util.Err{Err: err, Halt: false}
		}
		if cluster.Status != status {
			// This is a retryable error
			return nil, &util.Err{
				Err:  fmt.Errorf("cluster not yet in the %s state. Current state: %s", status, cluster.Status),
				Halt: false,
			}
		}
		// Success, no error
		return &cluster.Status, nil
	}, util.DefaultMaxNetworkFailures)

	return err
}

func (r *ClusterResource) getStateAndWaitForRunning(ctx context.Context, clusterId string) (lastState *ClusterResourceModel, isRunning bool) {
	const retryInterval = 10 * time.Second
	for {
		lastRet, isRunning := r.getStateAndCheckRunningOnce(ctx, clusterId)
		if isRunning {
			return lastRet, true
		}
		select {
		case <-ctx.Done():
			return lastRet, false
		case <-time.After(retryInterval):
		}
	}
}

func (r *ClusterResource) getStateAndCheckRunningOnce(ctx context.Context, clusterId string) (lastState *ClusterResourceModel, isRunning bool) {
	ret, err := r.store.Get(ctx, clusterId)
	if err != nil {
		tflog.Warn(ctx, "Failed to get cluster state", map[string]interface{}{
			"cluster_id": clusterId,
			"error":      err.Error(),
		})
		return nil, false
	}
	return ret, ret.Status.ValueString() == "RUNNING"
}
