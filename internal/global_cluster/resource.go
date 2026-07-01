package global_cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var (
	globalClusterPostCreateDescribeDelay = 10 * time.Second
	globalClusterSecondaryPollInterval   = 10 * time.Second
	globalClusterSecondaryDeleteTimeout  = 30 * time.Minute
	globalClusterSecondaryRunningTimeout = 30 * time.Minute
)

var (
	_ resource.Resource                = &GlobalClusterResource{}
	_ resource.ResourceWithConfigure   = &GlobalClusterResource{}
	_ resource.ResourceWithImportState = &GlobalClusterResource{}
)

func NewGlobalClusterResource() resource.Resource {
	return newGlobalClusterResource()
}

type GlobalClusterResource struct {
	client                *zilliz.Client
	store                 GlobalClusterStore
	sureInstanceNotExists func(globalClusterID string, clusterID string) condition
	sureGlobalClusterCU   func(globalClusterID string, targetCUSize int64) condition
	sureInstanceIsRunning func(globalClusterID string, member GlobalClusterMemberSpec) condition
	surePrimaryRunning    func(globalClusterID string) condition
}

func newGlobalClusterResource() *GlobalClusterResource {
	r := &GlobalClusterResource{}
	r.sureInstanceNotExists = func(globalClusterID string, clusterID string) condition {
		return func(ctx context.Context) (bool, string, error) {
			globalCluster, err := r.store.Describe(ctx, globalClusterID)
			if err != nil {
				return false, "", err
			}
			return globalCluster.isInstanceNotExists(clusterID)
		}
	}
	r.sureGlobalClusterCU = func(globalClusterID string, targetCUSize int64) condition {
		return func(ctx context.Context) (bool, string, error) {
			globalCluster, err := r.store.Describe(ctx, globalClusterID)
			if err != nil {
				return false, "", err
			}
			return globalCluster.isCUUpdated(targetCUSize)
		}
	}
	r.sureInstanceIsRunning = func(globalClusterID string, member GlobalClusterMemberSpec) condition {
		return func(ctx context.Context) (bool, string, error) {
			globalCluster, err := r.store.Describe(ctx, globalClusterID)
			if err != nil {
				return false, "", err
			}
			return globalCluster.isSecondaryMemberRunning(member)
		}
	}
	r.surePrimaryRunning = func(globalClusterID string) condition {
		return func(ctx context.Context) (bool, string, error) {
			globalCluster, err := r.store.Describe(ctx, globalClusterID)
			if err != nil {
				return false, "", err
			}
			return globalCluster.isPrimaryMemberRunning()
		}
	}

	return r
}

func (r *GlobalClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_cluster"
}

func (r *GlobalClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Zilliz Cloud Global Cluster. The first cluster entry is created as the primary cluster, and every subsequent entry is created as a secondary cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Global cluster identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"global_cluster_name": schema.StringAttribute{
				MarkdownDescription: "Global cluster display name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID where the global cluster is created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cu_type": schema.StringAttribute{
				MarkdownDescription: "CU type shared by primary and secondary clusters.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Performance-optimized"),
				Validators: []validator.String{
					stringvalidator.OneOf("Performance-optimized", "Capacity-optimized", "Tiered-storage", "Extended-capacity"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cu_size": schema.Int64Attribute{
				MarkdownDescription: "CU size shared by primary and secondary clusters.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"cluster": globalClusterMembersNestedAttribute(),
			"connect_address": schema.StringAttribute{
				MarkdownDescription: "Global endpoint address.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"create_time": schema.StringAttribute{
				MarkdownDescription: "Creation time.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region_ids": schema.ListAttribute{
				MarkdownDescription: "Region IDs of member clusters in API member order.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Initial database username returned by create.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Initial database password returned by create.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"create_job_id": schema.StringAttribute{
				MarkdownDescription: "Create operation job identifier returned by the Global Cluster API.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func globalClusterMembersNestedAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Ordered member cluster parameters. The first item is the primary cluster; all remaining items are secondary clusters. Updates may add or remove secondary clusters, but existing members cannot be modified in place.",
		Required:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"cluster_id": schema.StringAttribute{
					MarkdownDescription: "Member cluster identifier assigned by the Global Cluster API.",
					Computed:            true,
				},
				"cluster_name": schema.StringAttribute{
					MarkdownDescription: "Member cluster name.",
					Required:            true,
				},
				"region_id": schema.StringAttribute{
					MarkdownDescription: "Member cluster region ID.",
					Required:            true,
				},
				"role": schema.StringAttribute{
					MarkdownDescription: "Member role. Values are PRIMARY and SECONDARY.",
					Computed:            true,
				},
				"status": schema.StringAttribute{
					MarkdownDescription: "Current member cluster status.",
					Computed:            true,
				},
			},
		},
	}
}

func (r *GlobalClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider type",
			fmt.Sprintf("Expected *zilliz.Client, got: %T. Please check provider configuration.", req.ProviderData),
		)
		return
	}
	r.client = client
	r.store = NewGlobalClusterStore(client)
}

func (r *GlobalClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GlobalClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.store.Create(ctx, CreateGlobalClusterCommand{
		GlobalClusterName: data.GlobalClusterName.ValueString(),
		ProjectID:         data.ProjectID.ValueString(),
		CUType:            data.CUType.ValueString(),
		CUSize:            data.CUSize.ValueInt64(),
		Members:           data.memberSpecs(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create global cluster",
			fmt.Sprintf("global_cluster_name=%s project_id=%s error=%s", data.GlobalClusterName.ValueString(), data.ProjectID.ValueString(), err.Error()),
		)
		return
	}

	data.ID = types.StringValue(created.GlobalClusterID)
	data.Username = types.StringValue(created.Username)
	data.Password = types.StringValue(created.Password)
	data.CreateJobID = types.StringValue(created.JobID)

	time.Sleep(globalClusterPostCreateDescribeDelay)
	globalCluster, err := r.store.Describe(ctx, created.GlobalClusterID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read global cluster after create",
			fmt.Sprintf("global_cluster_id=%s error=%s", created.GlobalClusterID, err.Error()),
		)
		return
	}
	data.applyGlobalCluster(ctx, globalCluster, resp.Diagnostics.Append)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GlobalClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GlobalClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	globalCluster, err := r.store.Describe(ctx, state.ID.ValueString())
	if err != nil {
		if IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read global cluster",
			fmt.Sprintf("global_cluster_id=%s error=%s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	state.applyGlobalCluster(ctx, globalCluster, resp.Diagnostics.Append)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GlobalClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GlobalClusterResourceModel
	var state GlobalClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	globalCluster, err := r.store.Describe(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read global cluster before update",
			fmt.Sprintf("global_cluster_id=%s error=%s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	changePlan, err := globalCluster.PlanSecondaryClusterChange(plan.memberSpecs())
	if err != nil {
		resp.Diagnostics.AddError("Cannot modify global cluster members", err.Error())
		return
	}

	for _, member := range changePlan.Delete {
		if err := r.store.DeleteCluster(ctx, state.ID.ValueString(), member.ClusterID); err != nil {
			resp.Diagnostics.AddError(
				"Failed to delete secondary global cluster member",
				fmt.Sprintf("global_cluster_id=%s cluster_id=%s error=%s", state.ID.ValueString(), member.ClusterID, err.Error()),
			)
			return
		}
		if err := waitFor(
			ctx,
			globalClusterSecondaryDeleteTimeout,
			globalClusterSecondaryPollInterval,
			r.sureInstanceNotExists(state.ID.ValueString(), member.ClusterID),
			func(lastStatus string) error {
				return fmt.Errorf("secondary cluster %s still appears in global cluster members with status %s", member.ClusterID, lastStatus)
			},
		); err != nil {
			resp.Diagnostics.AddError(
				"Timed out waiting for secondary global cluster member deletion",
				fmt.Sprintf("global_cluster_id=%s cluster_id=%s error=%s", state.ID.ValueString(), member.ClusterID, err.Error()),
			)
			return
		}
	}

	if len(changePlan.Add) > 0 {
		if err := r.store.AddSecondaryClusters(ctx, state.ID.ValueString(), changePlan.Add); err != nil {
			resp.Diagnostics.AddError(
				"Failed to add secondary global cluster members",
				fmt.Sprintf("global_cluster_id=%s error=%s", state.ID.ValueString(), err.Error()),
			)
			return
		}
		for _, member := range changePlan.Add {
			if err := waitFor(
				ctx,
				globalClusterSecondaryRunningTimeout,
				globalClusterSecondaryPollInterval,
				r.sureInstanceIsRunning(state.ID.ValueString(), member),
				func(lastStatus string) error {
					return fmt.Errorf("secondary cluster %s in region %s did not reach RUNNING status; last status %s", member.ClusterName, member.RegionID, lastStatus)
				},
			); err != nil {
				resp.Diagnostics.AddError(
					"Timed out waiting for secondary global cluster member creation",
					fmt.Sprintf("global_cluster_id=%s cluster_name=%s region_id=%s error=%s", state.ID.ValueString(), member.ClusterName, member.RegionID, err.Error()),
				)
				return
			}
		}
	}

	if !plan.CUSize.Equal(state.CUSize) {
		if err := r.store.ModifyCU(ctx, state.ID.ValueString(), plan.CUSize.ValueInt64()); err != nil {
			resp.Diagnostics.AddError(
				"Failed to modify global cluster CU",
				fmt.Sprintf("global_cluster_id=%s cu_size=%d error=%s", state.ID.ValueString(), plan.CUSize.ValueInt64(), err.Error()),
			)
			return
		}

		targetCUSize := plan.CUSize.ValueInt64()
		if err := waitFor(
			ctx,
			globalClusterSecondaryRunningTimeout,
			globalClusterSecondaryPollInterval,
			r.sureGlobalClusterCU(state.ID.ValueString(), targetCUSize),
			func(lastStatus string) error {
				return fmt.Errorf("global cluster did not reach cu_size %d with RUNNING members; last status %s", targetCUSize, lastStatus)
			},
		); err != nil {
			resp.Diagnostics.AddError(
				"Timed out waiting for global cluster CU modification",
				fmt.Sprintf("global_cluster_id=%s cu_size=%d error=%s", state.ID.ValueString(), targetCUSize, err.Error()),
			)
			return
		}

	}

	globalCluster, err = r.store.Describe(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read global cluster after update",
			fmt.Sprintf("global_cluster_id=%s error=%s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	plan.ID = state.ID
	plan.Username = state.Username
	plan.Password = state.Password
	plan.CreateJobID = state.CreateJobID
	plan.applyGlobalCluster(ctx, globalCluster, resp.Diagnostics.Append)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

type condition func(context.Context) (done bool, lastStatus string, err error)

func waitFor(
	ctx context.Context,
	timeout time.Duration,
	pollInterval time.Duration,
	condition condition,
	timeoutError func(string) error,
) error {
	deadline := time.Now().Add(timeout)
	lastStatus := "unknown"

	for {
		done, status, err := condition(ctx)
		if err != nil {
			return err
		}
		if status != "" {
			lastStatus = status
		}
		if done {
			return nil
		}

		if time.Now().After(deadline) {
			return timeoutError(lastStatus)
		}

		if pollInterval <= 0 {
			pollInterval = time.Second
		}

		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (r *GlobalClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GlobalClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	globalCluster, err := r.store.Describe(ctx, state.ID.ValueString())
	if err != nil {
		if IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read global cluster before delete",
			fmt.Sprintf("global_cluster_id=%s error=%s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	for _, member := range globalCluster.Clusters {
		if member.Role != GlobalClusterMemberRoleSecondary {
			continue
		}

		if err := waitFor(
			ctx,
			globalClusterSecondaryDeleteTimeout,
			globalClusterSecondaryPollInterval,
			r.surePrimaryRunning(state.ID.ValueString()),
			func(lastStatus string) error {
				return fmt.Errorf("secondary cluster %s still appears in global cluster members with status %s", member.ClusterID, lastStatus)
			},
		); err != nil {
			resp.Diagnostics.AddError(
				"Timed out waiting for secondary global cluster member deletion",
				fmt.Sprintf("global_cluster_id=%s cluster_id=%s error=%s", state.ID.ValueString(), member.ClusterID, err.Error()),
			)
			return
		}
		if err := r.store.DeleteCluster(ctx, state.ID.ValueString(), member.ClusterID); err != nil {
			if IsNotFoundError(err) {
				continue
			}
			resp.Diagnostics.AddError(
				"Failed to delete secondary global cluster member",
				fmt.Sprintf("global_cluster_id=%s cluster_id=%s error=%s", state.ID.ValueString(), member.ClusterID, err.Error()),
			)
			return
		}
	}

	if err := waitFor(
		ctx,
		globalClusterSecondaryDeleteTimeout,
		globalClusterSecondaryPollInterval,
		r.surePrimaryRunning(state.ID.ValueString()),
		func(lastStatus string) error {
			return fmt.Errorf("primary global cluster member did not become RUNNING after secondary member deletion; last observed primary status %s", lastStatus)
		},
	); err != nil {
		resp.Diagnostics.AddError(
			"Timed out waiting for primary global cluster member to become running before deletion",
			fmt.Sprintf("global_cluster_id=%s error=%s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	globalCluster, err = r.store.Describe(ctx, state.ID.ValueString())
	if err != nil {
		if IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read global cluster before primary delete",
			fmt.Sprintf("global_cluster_id=%s error=%s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	var primaryClusterID string
	for _, member := range globalCluster.Clusters {
		if member.Role == GlobalClusterMemberRolePrimary {
			primaryClusterID = member.ClusterID
			break
		}
	}
	if primaryClusterID == "" {
		resp.Diagnostics.AddError(
			"Failed to delete primary global cluster member",
			fmt.Sprintf("global_cluster_id=%s error=primary cluster is missing from global cluster members", state.ID.ValueString()),
		)
		return
	}

	if err := r.store.DeleteCluster(ctx, state.ID.ValueString(), primaryClusterID); err != nil {
		if IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete primary global cluster member",
			fmt.Sprintf("global_cluster_id=%s cluster_id=%s error=%s", state.ID.ValueString(), primaryClusterID, err.Error()),
		)
	}
}

func (r *GlobalClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
