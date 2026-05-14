package cluster

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
	util "github.com/zilliztech/terraform-provider-zillizcloud/client/retry"
)

var (
	_ resource.Resource              = &OnDemandClusterResource{}
	_ resource.ResourceWithConfigure = &OnDemandClusterResource{}
)

func NewOnDemandClusterResource() resource.Resource {
	return &OnDemandClusterResource{}
}

type OnDemandClusterResource struct {
	client *zilliz.Client
	store  OnDemandClusterStore
}

func (r *OnDemandClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_on_demand_cluster"
}

func (r *OnDemandClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "On-demand Query Cluster resource. Configurable fields are replacement-only because the public API does not expose an update endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "On-demand Query Cluster identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project where the on-demand Query Cluster is created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the cloud region where the on-demand Query Cluster is created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"cluster_name": schema.StringAttribute{
				MarkdownDescription: "The name of the on-demand Query Cluster.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(64),
				},
			},
			"cu_size": schema.Int64Attribute{
				MarkdownDescription: "The initial CU size. The value must be at least 8.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(8),
				},
			},
			"auto_suspend": schema.Int64Attribute{
				MarkdownDescription: "Auto-suspend duration in seconds for the on-demand Query Cluster. Defaults to 1800 seconds.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1800),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					autoSuspendValidator{},
				},
			},
			"max_query_node_cu": schema.Int64Attribute{
				MarkdownDescription: "Maximum query node CU when set.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"max_query_node_replicas": schema.Int64Attribute{
				MarkdownDescription: "Maximum query node replicas when set.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"replicas": schema.Int64Attribute{
				MarkdownDescription: "Current replica count.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"ready_replicas": schema.Int64Attribute{
				MarkdownDescription: "Current ready replica count.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current on-demand Query Cluster status.",
				Computed:            true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Public endpoint inherited from the parent VectorLake.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"private_link": schema.StringAttribute{
				MarkdownDescription: "Private link endpoint inherited from the parent VectorLake, when available.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "Identity that created the on-demand Query Cluster.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"create_time": schema.Int64Attribute{
				MarkdownDescription: "Creation time in milliseconds since epoch.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"ttl_seconds": schema.Int64Attribute{
				MarkdownDescription: "Session TTL in seconds as reported by the API.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"prompt": schema.StringAttribute{
				MarkdownDescription: "The statement indicating that the latest create or delete operation succeeded.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx,
				timeouts.Opts{
					Create: true,
					CreateDescription: `Timeout defaults to 45 mins. Accepts a string that can be parsed as a duration ` +
						`such as "30s", "45m", or "1h".`,
					Delete:            true,
					DeleteDescription: `Timeout accepts a string that can be parsed as a duration. Delete returns after the API accepts the request.`,
				},
			),
		},
	}
}

func (r *OnDemandClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
	r.store = &OnDemandClusterStoreImpl{client: client}
}

func (r *OnDemandClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Create OnDemandCluster...")

	var plan OnDemandClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, defaultOnDemandClusterCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	created, err := r.store.Create(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create on-demand cluster",
			fmt.Sprintf("project_id=%s region_id=%s cluster_name=%s error=%s",
				plan.ProjectID.ValueString(), plan.RegionID.ValueString(), plan.ClusterName.ValueString(), err),
		)
		return
	}

	state := plan
	state.ID = created.ID
	state.Prompt = created.Prompt
	state.setOnDemandComputedNulls()

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	refreshed, err := r.waitForOnDemandStatus(ctx, createTimeout, state.ID.ValueString(), "SUSPENDED")
	if refreshed != nil {
		state.populateOnDemandComputed(refreshed)
	}
	if err != nil {
		resp.Diagnostics.AddWarning("On-demand cluster created but not in RUNNING state", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *OnDemandClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Read OnDemandCluster...")

	var state OnDemandClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.store.Get(ctx, state.ID.ValueString())
	if err != nil {
		if isOnDemandClusterNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to get on-demand cluster", err.Error())
		return
	}

	state.populateOnDemandComputed(cluster)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *OnDemandClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"On-demand cluster updates are not supported",
		"The public Zilliz Cloud On-demand Query Cluster API does not expose an update endpoint. "+
			"All configurable attributes on zillizcloud_on_demand_cluster require replacement; "+
			"if Terraform reached Update, replace the resource instead.",
	)
}

func (r *OnDemandClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Delete OnDemandCluster...")

	var state OnDemandClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.store.Delete(ctx, state.ID.ValueString()); err != nil {
		if isOnDemandClusterNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete on-demand cluster", err.Error())
		return
	}
}

func (r *OnDemandClusterResource) waitForOnDemandStatus(ctx context.Context, timeout time.Duration, clusterID string, status string) (*OnDemandClusterResourceModel, error) {
	var lastState *OnDemandClusterResourceModel
	_, err := util.NetworkResilientPoll(ctx, timeout, func() (*string, *util.Err) {
		cluster, err := r.store.Get(ctx, clusterID)
		if err != nil {
			if isOnDemandClusterNotFound(err) {
				return nil, &util.Err{Err: err, Halt: true}
			}
			return nil, &util.Err{Err: err, Halt: false}
		}
		lastState = cluster
		if cluster.Status.ValueString() != status {
			return nil, &util.Err{
				Err:  fmt.Errorf("on-demand cluster not yet in the %s state. Current state: %s", status, cluster.Status.ValueString()),
				Halt: false,
			}
		}
		result := status
		return &result, nil
	}, util.DefaultMaxNetworkFailures)

	if err != nil && util.IsNetworkGiveUpError(err) {
		return lastState, err
	}
	return lastState, err
}

func (m *OnDemandClusterResourceModel) setOnDemandComputedNulls() {
	m.Replicas = types.Int64Null()
	m.ReadyReplicas = types.Int64Null()
	m.Status = types.StringNull()
	m.Endpoint = types.StringNull()
	m.PrivateLink = types.StringNull()
	m.CreatedBy = types.StringNull()
	m.CreateTime = types.Int64Null()
	m.TTLSeconds = types.Int64Null()
}

func (m *OnDemandClusterResourceModel) populateOnDemandComputed(input *OnDemandClusterResourceModel) {
	if !input.ID.IsNull() && !input.ID.IsUnknown() && input.ID.ValueString() != "" {
		m.ID = input.ID
	}
	if !input.RegionID.IsNull() && !input.RegionID.IsUnknown() && input.RegionID.ValueString() != "" {
		m.RegionID = input.RegionID
	}
	if !input.ClusterName.IsNull() && !input.ClusterName.IsUnknown() && input.ClusterName.ValueString() != "" {
		m.ClusterName = input.ClusterName
	}
	if !input.CUSize.IsNull() && !input.CUSize.IsUnknown() {
		m.CUSize = input.CUSize
	}
	if !input.AutoSuspend.IsNull() && !input.AutoSuspend.IsUnknown() {
		m.AutoSuspend = input.AutoSuspend
	}
	m.Replicas = input.Replicas
	m.ReadyReplicas = input.ReadyReplicas
	m.Status = input.Status
	m.Endpoint = input.Endpoint
	m.PrivateLink = input.PrivateLink
	m.CreatedBy = input.CreatedBy
	m.CreateTime = input.CreateTime
	m.TTLSeconds = input.TTLSeconds
}

func isOnDemandClusterNotFound(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "http status code: 404") ||
		strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "notfound")
}
