package cluster

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

const (
	restorePollInterval            = 5 * time.Second
	defaultRestoreDeleteTimeout    = 2 * time.Minute
	defaultRestoreCollectionStatus = "RELEASE"
	// Control API also uses this code when the cluster is being deleted.
	clusterNotFoundCode = 40021
)

var restoreImportID = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

var _ resource.ResourceWithImportState = (*ClusterRestoreResource)(nil)

type ClusterRestoreResource struct {
	client *zilliz.Client
}

type clusterRestoreModel struct {
	ID                   types.String   `tfsdk:"id"`
	SourceClusterID      types.String   `tfsdk:"source_cluster_id"`
	BackupID             types.String   `tfsdk:"backup_id"`
	ProjectID            types.String   `tfsdk:"project_id"`
	ClusterName          types.String   `tfsdk:"cluster_name"`
	CUSize               types.Int64    `tfsdk:"cu_size"`
	Description          types.String   `tfsdk:"description"`
	CollectionStatus     types.String   `tfsdk:"collection_status"`
	Version              types.Object   `tfsdk:"version"`
	RestoreVersionPolicy types.String   `tfsdk:"restore_version_policy"`
	JobID                types.String   `tfsdk:"job_id"`
	RestoreStatus        types.String   `tfsdk:"restore_status"`
	RegionID             types.String   `tfsdk:"region_id"`
	MilvusVersion        types.String   `tfsdk:"milvus_version"`
	ConnectAddress       types.String   `tfsdk:"connect_address"`
	Username             types.String   `tfsdk:"username"`
	Password             types.String   `tfsdk:"password"`
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
}

func NewClusterRestoreResource() resource.Resource {
	return &ClusterRestoreResource{}
}

func (r *ClusterRestoreResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_restore"
}

func (r *ClusterRestoreResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{resourcevalidator.Conflicting(path.MatchRoot("version"), path.MatchRoot("restore_version_policy"))}
}

func (r *ClusterRestoreResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	version := versionSelectionSchema("Explicit version selected from the candidates for this backup. Conflicts with restore_version_policy.", "GA", "PUBLIC_PREVIEW", "ORIGIN")
	version.PlanModifiers = []planmodifier.Object{objectplanmodifier.RequiresReplace()}
	keepKnownValue := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Restores an existing backup into a new Dedicated cluster in the source cluster's region. " +
			"Name, description, and CU size are updated in place. Changing the source, backup, target project, collection status, or version selection replaces the cluster and can delete data added after the restore. " +
			"Tracking the restore job requires observability read permission on the target project. " +
			"A failed or canceled job fails the apply and protects the recorded target from automatic deletion. If the target disappears, a subsequent plan can retry the restore. An observation timeout records the target and job with a warning so refresh can resume tracking. Until success has been observed, a job-read failure blocks refresh and apply.",
		Attributes: map[string]schema.Attribute{
			"source_cluster_id": requiredRestoreInput("ID of the cluster that owns the backup."),
			"backup_id":         requiredRestoreInput("ID of the backup to restore."),
			"project_id":        requiredRestoreInput("ID of the target project."),
			"cluster_name": schema.StringAttribute{
				MarkdownDescription: "Name of the restored cluster.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.RegexMatches(nonBlank, "Provide a non-blank name.")},
			},
			"cu_size": schema.Int64Attribute{
				MarkdownDescription: "Number of compute units for the restored cluster.",
				Required:            true,
				Validators:          []validator.Int64{int64validator.Between(1, 1024)},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the restored cluster. When omitted, track the current description.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{stringvalidator.LengthAtMost(255)},
			},
			"collection_status": schema.StringAttribute{
				MarkdownDescription: "Whether to KEEP or RELEASE the restored collections. Defaults to RELEASE.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(defaultRestoreCollectionStatus),
				Validators:          []validator.String{stringvalidator.OneOf("KEEP", defaultRestoreCollectionStatus)},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"version": version,
			"restore_version_policy": schema.StringAttribute{
				MarkdownDescription: "Restore using the LATEST or ORIGINAL policy. Conflicts with version. Omit both to use the API default.",
				Optional:            true,
				Validators:          []validator.String{stringvalidator.OneOf("LATEST", "ORIGINAL")},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"id":              schema.StringAttribute{Computed: true, PlanModifiers: keepKnownValue, MarkdownDescription: "ID of the new cluster."},
			"job_id":          schema.StringAttribute{Computed: true, PlanModifiers: keepKnownValue, MarkdownDescription: "ID of the asynchronous restore job."},
			"restore_status":  schema.StringAttribute{Computed: true, MarkdownDescription: "Latest known restore job status. If waiting times out, Terraform records the accepted cluster and job so refresh can continue tracking them."},
			"region_id":       schema.StringAttribute{Computed: true, PlanModifiers: keepKnownValue, MarkdownDescription: "Region of the restored cluster."},
			"milvus_version":  schema.StringAttribute{Computed: true, MarkdownDescription: "Milvus version series currently reported by the restored cluster."},
			"connect_address": schema.StringAttribute{Computed: true, MarkdownDescription: "Endpoint of the restored cluster."},
			"username":        schema.StringAttribute{Computed: true, PlanModifiers: keepKnownValue, MarkdownDescription: "Default database username returned when the restore is submitted."},
			"password":        schema.StringAttribute{Computed: true, PlanModifiers: keepKnownValue, Sensitive: true, MarkdownDescription: "Default database password returned when the restore is submitted."},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create:            true,
				Update:            true,
				UpdateDescription: "Time to apply cluster changes. Defaults to 30 minutes.",
				Delete:            true,
				DeleteDescription: "Time allowed for the delete request. Defaults to 2 minutes.",
				CreateDescription: "Time to wait for the restore to finish. Defaults to 45 minutes. " +
					"After submission, a timeout produces a warning and keeps the cluster and job IDs in state.",
			}),
		},
	}
}

func requiredRestoreInput(description string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: description,
		Required:            true,
		Validators:          []validator.String{stringvalidator.RegexMatches(nonBlank, "Provide a non-blank value.")},
		PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
}

func (r *ClusterRestoreResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configuration", "Expected a Zilliz Cloud client.")
		return
	}
	r.client = client
}

func (r *ClusterRestoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state clusterRestoreModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	wait, diags := state.Timeouts.Create(ctx, defaultClusterCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	version, err := versionSelection(ctx, state.Version)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("version"), "Invalid version selection", err.Error())
		return
	}
	if version != nil && !state.RestoreVersionPolicy.IsNull() {
		resp.Diagnostics.AddError("Conflicting restore options", "Set either version or restore_version_policy.")
		return
	}

	waitCtx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()
	result, err := r.client.RestoreCluster(waitCtx, state.SourceClusterID.ValueString(), state.BackupID.ValueString(), zilliz.RestoreClusterParams{
		ClusterName:          state.ClusterName.ValueString(),
		ProjectID:            state.ProjectID.ValueString(),
		CUSize:               state.CUSize.ValueInt64(),
		Description:          state.Description.ValueString(),
		CollectionStatus:     state.CollectionStatus.ValueString(),
		Version:              version,
		RestoreVersionPolicy: state.RestoreVersionPolicy.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to submit cluster restore", err.Error())
		return
	}
	if result.ClusterID == "" {
		resp.Diagnostics.AddError("Restore response omitted the target cluster ID",
			fmt.Sprintf("The API returned job ID %q. Check the job and target project before submitting another restore.", result.JobID))
		return
	}

	state.ID = types.StringValue(result.ClusterID)
	state.JobID = types.StringNull()
	state.RestoreStatus = types.StringValue("PENDING")
	state.RegionID = types.StringNull()
	state.MilvusVersion = types.StringNull()
	state.ConnectAddress = types.StringNull()
	if state.Description.IsUnknown() {
		state.Description = types.StringNull()
	}
	state.Username = types.StringValue(result.Username)
	state.Password = types.StringValue(result.Password)
	if result.JobID != "" {
		state.JobID = types.StringValue(result.JobID)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// A missing observation does not establish failure after the restore was accepted.
	if result.JobID == "" {
		resp.Diagnostics.AddWarning("Restore response omitted the job ID",
			fmt.Sprintf("Target cluster %s is in state. Check its restore status in the Zilliz Cloud console.", result.ClusterID))
		return
	}
	job, err := r.waitForRestore(waitCtx, result.JobID)
	if job.Status != "" {
		state.RestoreStatus = types.StringValue(job.Status)
	}
	switch {
	case err != nil:
		resp.Diagnostics.AddWarning("Cluster restore was accepted", fmt.Sprintf(
			"Terraform retained target cluster %s and restore job %s. Waiting ended with: %s", result.ClusterID, result.JobID, err))
	case restoreFailed(job.Status):
		resp.Diagnostics.AddError("Cluster restore did not succeed", restoreFailureDetail(state, job.Reason))
	default:
		cluster, err := r.client.DescribeClusterContext(waitCtx, result.ClusterID)
		if err != nil {
			resp.Diagnostics.AddWarning("Unable to read restored cluster", err.Error())
		} else {
			state.syncTargetOutputs(cluster)
			if state.Description.IsNull() {
				state.Description = types.StringValue(cluster.Description)
			}
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ClusterRestoreResource) waitForRestore(ctx context.Context, jobID string) (zilliz.RestoreJob, error) {
	var last zilliz.RestoreJob
	ticker := time.NewTicker(restorePollInterval)
	defer ticker.Stop()
	for {
		job, err := r.client.DescribeRestoreJob(ctx, jobID)
		if err != nil {
			return last, err
		}
		last = job
		if restoreFinished(job.Status) {
			return job, nil
		}
		select {
		case <-ctx.Done():
			return last, ctx.Err()
		case <-ticker.C:
		}
	}
}

func restoreFinished(status string) bool {
	return status == "SUCCESSFUL" || restoreFailed(status)
}

func restoreFailed(status string) bool {
	return status == "FAILED" || status == "CANCELED"
}

func restoreFailureDetail(state clusterRestoreModel, reason string) string {
	return fmt.Sprintf("Restore job %s for cluster %s reached %s. Inspect the job and target cluster. Terraform will retry on a later apply once the target no longer exists. Job detail: %s",
		state.JobID.ValueString(), state.ID.ValueString(), state.RestoreStatus.ValueString(), reason)
}

func (state *clusterRestoreModel) syncTargetOutputs(cluster zilliz.Cluster) {
	state.RegionID = types.StringValue(cluster.RegionId)
	state.MilvusVersion = types.StringPointerValue(cluster.MilvusVersion)
	state.ConnectAddress = types.StringValue(cluster.ConnectAddress)
}

func (r *ClusterRestoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterRestoreModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.CollectionStatus.IsNull() {
		state.CollectionStatus = types.StringValue(defaultRestoreCollectionStatus)
	}
	var reason string
	if !state.JobID.IsNull() && !restoreFinished(state.RestoreStatus.ValueString()) {
		job, err := r.client.DescribeRestoreJob(ctx, state.JobID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to read cluster restore job", fmt.Sprintf(
				"Cannot confirm restore job %s for cluster %s. Retry after the job service is available: %s",
				state.JobID.ValueString(), state.ID.ValueString(), err))
			return
		}
		if job.Status != "" {
			state.RestoreStatus = types.StringValue(job.Status)
			reason = job.Reason
		}
	}
	status := state.RestoreStatus.ValueString()
	cluster, err := r.client.DescribeClusterContext(ctx, state.ID.ValueString())
	if err != nil {
		if !errors.Is(err, zilliz.Error{Code: clusterNotFoundCode}) {
			resp.Diagnostics.AddError("Unable to read restored cluster", err.Error())
			return
		}
		// The target may be temporarily unavailable while a restore is running.
		if restoreFinished(status) || (status == "" && state.JobID.IsNull()) {
			resp.State.RemoveResource(ctx)
			return
		}
	} else {
		state.syncTargetOutputs(cluster)
		if status == "" || status == "SUCCESSFUL" {
			state.ClusterName = types.StringValue(cluster.ClusterName)
			state.CUSize = types.Int64Value(cluster.CuSize)
			state.Description = types.StringValue(cluster.Description)
			state.ProjectID = types.StringValue(cluster.ProjectId)
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if restoreFailed(status) {
		resp.Diagnostics.AddError("Cluster restore did not succeed", restoreFailureDetail(state, reason))
	}
}

func (r *ClusterRestoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state clusterRestoreModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	wait, diags := plan.Timeouts.Update(ctx, defaultClusterUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateCtx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()
	defer func() { resp.Diagnostics.Append(resp.State.Set(ctx, &state)...) }()

	var properties zilliz.ModifyPropertiesParams
	if !plan.ClusterName.Equal(state.ClusterName) {
		properties.ClusterName = plan.ClusterName.ValueString()
	}
	if !plan.Description.IsUnknown() && !plan.Description.IsNull() && !plan.Description.Equal(state.Description) {
		properties.Description = plan.Description.ValueStringPointer()
	}
	if properties.ClusterName != "" || properties.Description != nil {
		if _, err := r.client.ModifyClusterPropertiesContext(updateCtx, state.ID.ValueString(), &properties); err != nil {
			resp.Diagnostics.AddError("Unable to update restored cluster properties", err.Error())
			return
		}
		state.ClusterName = plan.ClusterName
		if properties.Description != nil {
			state.Description = plan.Description
		}
	}
	if !plan.CUSize.Equal(state.CUSize) {
		params := &zilliz.ModifyClusterParams{CuSize: int(plan.CUSize.ValueInt64())}
		if _, err := r.client.ModifyClusterContext(updateCtx, state.ID.ValueString(), params); err != nil {
			resp.Diagnostics.AddError("Unable to resize restored cluster", err.Error())
			return
		}
		ticker := time.NewTicker(restorePollInterval)
		defer ticker.Stop()
		for {
			cluster, err := r.client.DescribeClusterContext(updateCtx, state.ID.ValueString())
			if err == nil && cluster.Status == "RUNNING" && cluster.CuSize == plan.CUSize.ValueInt64() {
				state.CUSize = plan.CUSize
				break
			}
			select {
			case <-updateCtx.Done():
				resp.Diagnostics.AddError("Unable to confirm restored cluster resize", updateCtx.Err().Error())
				return
			case <-ticker.C:
			}
		}
	}
	state.Timeouts = plan.Timeouts
}

func (r *ClusterRestoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clusterRestoreModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	wait, diags := state.Timeouts.Delete(ctx, defaultRestoreDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	deleteCtx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()

	// Terraform can attempt a replacement after Create returns an error with state.
	// Protect an existing failed target even when refresh is disabled.
	if restoreFailed(state.RestoreStatus.ValueString()) {
		_, err := r.client.DescribeClusterContext(deleteCtx, state.ID.ValueString())
		switch {
		case err == nil:
			resp.Diagnostics.AddError("Refusing to delete the failed restore target", restoreFailureDetail(state, "Confirm its disposition outside Terraform."))
		case !errors.Is(err, zilliz.Error{Code: clusterNotFoundCode}):
			resp.Diagnostics.AddError("Unable to verify whether the failed restore target still exists", err.Error())
		}
		return
	}
	_, err := r.client.DropClusterContext(deleteCtx, state.ID.ValueString())
	if err != nil && !errors.Is(err, zilliz.Error{Code: clusterNotFoundCode}) {
		resp.Diagnostics.AddError("Unable to delete restored cluster", err.Error())
	}
}

func (r *ClusterRestoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ",")
	if len(parts) < 3 || len(parts) > 6 {
		resp.Diagnostics.AddError("Invalid restored cluster import ID",
			"Use target_cluster_id,source_cluster_id,backup_id[,job_id[,collection_status[,version_or_policy]]].")
		return
	}
	for index, id := range parts[:min(len(parts), 4)] {
		if index == 3 && id == "" {
			continue
		}
		if !restoreImportID.MatchString(id) {
			resp.Diagnostics.AddError("Invalid restored cluster import ID", "Provide valid target, source, backup, and optional job IDs.")
			return
		}
	}

	collectionStatus := defaultRestoreCollectionStatus
	if len(parts) >= 5 && parts[4] != "" {
		collectionStatus = parts[4]
		if collectionStatus != "KEEP" && collectionStatus != defaultRestoreCollectionStatus {
			resp.Diagnostics.AddError("Invalid imported collection status", "Use KEEP or RELEASE.")
			return
		}
	}
	if len(parts) == 6 && parts[5] != "" {
		switch selection := parts[5]; selection {
		case "LATEST", "ORIGINAL":
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("restore_version_policy"), selection)...)
		default:
			series, channel, ok := strings.Cut(selection, ":")
			if !ok || !milvusVersionPattern.MatchString(series) ||
				(channel != "GA" && channel != "PUBLIC_PREVIEW" && channel != "ORIGIN") {
				resp.Diagnostics.AddError("Invalid imported version selection", "Use LATEST, ORIGINAL, or a selection such as 2.6.x:ORIGIN.")
				return
			}
			value, diags := types.ObjectValue(
				map[string]attr.Type{"milvus_version": types.StringType, "channel": types.StringType},
				map[string]attr.Value{"milvus_version": types.StringValue(series), "channel": types.StringValue(channel)},
			)
			resp.Diagnostics.Append(diags...)
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("version"), value)...)
		}
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source_cluster_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("backup_id"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("collection_status"), collectionStatus)...)
	if len(parts) >= 4 && parts[3] != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_id"), parts[3])...)
	}
}
