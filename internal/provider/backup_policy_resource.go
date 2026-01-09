package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ resource.Resource = &BackupPolicyResource{}
var _ resource.ResourceWithConfigure = &BackupPolicyResource{}

func NewBackupPolicyResource() resource.Resource {
	return &BackupPolicyResource{}
}

type BackupPolicyResource struct {
	client *zilliz.Client
}

type CrossRegionCopy struct {
	RegionId      types.String `tfsdk:"region_id"`
	RetentionDays types.Int64  `tfsdk:"retention_days"`
}

type BackupPolicyResourceModel struct {
	Id                types.String      `tfsdk:"id"`
	ClusterId         types.String      `tfsdk:"cluster_id"`
	RegionId          types.String      `tfsdk:"region_id"`
	Frequency         types.String      `tfsdk:"frequency"`
	StartTime         types.String      `tfsdk:"start_time"`
	RetentionDays     types.Int64       `tfsdk:"retention_days"`
	Enabled           types.Bool        `tfsdk:"enabled"`
	CrossRegionCopies []CrossRegionCopy `tfsdk:"cross_region_copies"`
}

func (r *BackupPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup_policy"
}

func (r *BackupPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a backup policy for a Zilliz Cloud cluster.

This resource allows you to configure automated backups for a cluster, including backup frequency, retention period, and cross-region copies.

## Example Usage

` + "```hcl" + `
resource "zillizcloud_backup_policy" "example" {
  cluster_id     = zillizcloud_cluster.example.id
  region_id      = "aws-us-east-1"
  frequency      = "1,2,3,4,5"  # Days of week (1=Monday, 7=Sunday)
  start_time     = "02:00-04:00"
  retention_days = 7
  enabled        = true

  cross_region_copies = [
    {
      region_id      = "aws-us-west-2"
      retention_days = 5
    }
  ]
}
` + "```",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: `The unique identifier for the backup policy resource.

**Format:** ` + "`" + `/clusters/{cluster_id}/backups/policy` + "`",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `The ID of the cluster to configure the backup policy for.

You can obtain this value from the output of the ` + "`zillizcloud_cluster`" + ` resource.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region_id": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `The region ID where the cluster is located.

Example: ` + "`aws-us-east-1`" + `, ` + "`gcp-us-west1`",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"frequency": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `The backup frequency specified as days of the week.

Use a comma-separated list of numbers where 1=Monday through 7=Sunday.

**Examples:**
- ` + "`\"1,2,3,4,5\"`" + ` - Backup on weekdays
- ` + "`\"1,3,5\"`" + ` - Backup on Monday, Wednesday, Friday
- ` + "`\"7\"`" + ` - Backup on Sunday only`,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"start_time": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `The time window during which backups can start.

**Format:** ` + "`HH:MM-HH:MM`" + ` (24-hour format, UTC)

**Example:** ` + "`\"02:00-04:00\"`" + ` - Backups can start between 2:00 AM and 4:00 AM UTC`,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"retention_days": schema.Int64Attribute{
				Required: true,
				MarkdownDescription: `The number of days to retain backups.

**Valid range:** 1-30 days`,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
					int64validator.AtMost(30),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				MarkdownDescription: `Whether the backup policy is enabled.

Defaults to ` + "`true`" + `.`,
			},
			"cross_region_copies": schema.ListNestedAttribute{
				Optional: true,
				MarkdownDescription: `Configuration for cross-region backup copies.

This allows you to replicate backups to other regions for disaster recovery purposes.`,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"region_id": schema.StringAttribute{
							Required: true,
							MarkdownDescription: `The target region ID for the cross-region backup copy.

Example: ` + "`aws-us-west-2`",
						},
						"retention_days": schema.Int64Attribute{
							Required: true,
							MarkdownDescription: `The number of days to retain cross-region backup copies.

**Valid range:** 1-30 days`,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
								int64validator.AtMost(30),
							},
						},
					},
				},
			},
		},
	}
}

func (r *BackupPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
}

func (r *BackupPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BackupPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := r.buildBackupPolicyParams(&data)

	err := r.client.UpsertBackupPolicy(data.ClusterId.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create backup policy",
			fmt.Sprintf("ClusterId: %s, error: %s", data.ClusterId.ValueString(), err.Error()),
		)
		return
	}

	data.Id = types.StringValue(buildBackupPolicyID(data.ClusterId.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BackupPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BackupPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetBackupPolicy(state.ClusterId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read backup policy",
			fmt.Sprintf("ClusterId: %s, error: %s", state.ClusterId.ValueString(), err.Error()),
		)
		return
	}

	enabled := policy.Status == "ENABLED"

	// If policy is disabled and we're reading it, it means it was deleted externally
	if !enabled {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Frequency = types.StringValue(policy.Frequency)
	state.StartTime = types.StringValue(policy.StartTime)
	state.RetentionDays = types.Int64Value(int64(policy.RetentionDays))
	state.Enabled = types.BoolValue(enabled)

	state.CrossRegionCopies = make([]CrossRegionCopy, len(policy.CrossRegionCopies))
	for i, copy := range policy.CrossRegionCopies {
		state.CrossRegionCopies[i] = CrossRegionCopy{
			RegionId:      types.StringValue(copy.RegionId),
			RetentionDays: types.Int64Value(int64(copy.RetentionDays)),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BackupPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state BackupPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := r.buildBackupPolicyParams(&plan)

	err := r.client.UpsertBackupPolicy(plan.ClusterId.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update backup policy",
			fmt.Sprintf("ClusterId: %s, error: %s", plan.ClusterId.ValueString(), err.Error()),
		)
		return
	}

	plan.Id = state.Id
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BackupPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BackupPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteBackupPolicy(state.ClusterId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete backup policy",
			fmt.Sprintf("ClusterId: %s, error: %s", state.ClusterId.ValueString(), err.Error()),
		)
		return
	}
}

func (r *BackupPolicyResource) buildBackupPolicyParams(data *BackupPolicyResourceModel) *zilliz.BackupPolicyParams {
	params := &zilliz.BackupPolicyParams{
		Frequency:     data.Frequency.ValueString(),
		StartTime:     data.StartTime.ValueString(),
		RetentionDays: int(data.RetentionDays.ValueInt64()),
		Enabled:       data.Enabled.ValueBool(),
	}

	// Handle cross-region copies
	if len(data.CrossRegionCopies) > 0 {
		params.CrossRegionCopies = make([]zilliz.CrossRegionCopy, len(data.CrossRegionCopies))
		for i, copy := range data.CrossRegionCopies {
			params.CrossRegionCopies[i] = zilliz.CrossRegionCopy{
				RegionId:      copy.RegionId.ValueString(),
				RetentionDays: int(copy.RetentionDays.ValueInt64()),
			}
		}
	}

	return params
}

func buildBackupPolicyID(clusterId string) string {
	return fmt.Sprintf("/clusters/%s/backups/policy", clusterId)
}
