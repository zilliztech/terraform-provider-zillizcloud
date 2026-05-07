package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var (
	_ resource.Resource              = &VolumeResource{}
	_ resource.ResourceWithConfigure = &VolumeResource{}
)

var (
	volumeDeleteTimeout      = 2 * time.Minute
	volumeDeletePollInterval = 5 * time.Second
)

func NewVolumeResource() resource.Resource {
	return &VolumeResource{}
}

type VolumeResource struct {
	client *zilliz.Client
}

type VolumeResourceModel struct {
	Id                   types.String `tfsdk:"id"`
	ProjectId            types.String `tfsdk:"project_id"`
	RegionId             types.String `tfsdk:"region_id"`
	VolumeName           types.String `tfsdk:"volume_name"`
	Type                 types.String `tfsdk:"type"`
	StorageIntegrationId types.String `tfsdk:"storage_integration_id"`
	Path                 types.String `tfsdk:"path"`
	Status               types.String `tfsdk:"status"`
	CreateTime           types.String `tfsdk:"create_time"`
}

func (r *VolumeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *VolumeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a volume in a Zilliz Cloud project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Volume name used as the Terraform resource ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "Cloud region ID.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"volume_name": schema.StringAttribute{
				MarkdownDescription: "Volume name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Volume type. Valid values are MANAGED and EXTERNAL.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("MANAGED"),
				Validators: []validator.String{
					stringvalidator.OneOf("MANAGED", "EXTERNAL"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"storage_integration_id": schema.StringAttribute{
				MarkdownDescription: "Storage integration ID for external volumes.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "Storage path. If specified, it must be empty or end with '/'.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^$|.*/$`), "must be empty or end with '/'"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current volume status.",
				Computed:            true,
			},
			"create_time": schema.StringAttribute{
				MarkdownDescription: "Volume creation time.",
				Computed:            true,
			},
		},
	}
}

func (r *VolumeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VolumeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Type.ValueString() == "EXTERNAL" && (data.StorageIntegrationId.IsNull() || data.StorageIntegrationId.IsUnknown() || data.StorageIntegrationId.ValueString() == "") {
		resp.Diagnostics.AddError(
			"Missing storage integration ID",
			"storage_integration_id is required when type is EXTERNAL.",
		)
		return
	}

	body := &zilliz.CreateVolumeRequest{
		ProjectID:  data.ProjectId.ValueString(),
		RegionID:   data.RegionId.ValueString(),
		VolumeName: data.VolumeName.ValueString(),
		Type:       data.Type.ValueString(),
	}
	if !data.StorageIntegrationId.IsNull() && !data.StorageIntegrationId.IsUnknown() {
		body.StorageIntegrationID = data.StorageIntegrationId.ValueString()
	}
	if !data.Path.IsNull() && !data.Path.IsUnknown() {
		body.Path = data.Path.ValueString()
	}

	created, err := r.client.CreateVolume(body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create volume",
			fmt.Sprintf("project_id=%s volume_name=%s error=%s", data.ProjectId.ValueString(), data.VolumeName.ValueString(), err.Error()),
		)
		return
	}

	volumeName := data.VolumeName.ValueString()
	if created.VolumeName != "" {
		volumeName = created.VolumeName
	}
	data.Id = types.StringValue(volumeName)

	described, err := r.client.DescribeVolume(volumeName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read volume after create",
			fmt.Sprintf("volume_name=%s error=%s", volumeName, err.Error()),
		)
		return
	}
	data.applyDescribeVolume(described)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (data *VolumeResourceModel) applyDescribeVolume(volume *zilliz.DescribeVolumeData) {
	if volume.VolumeName != "" {
		data.VolumeName = types.StringValue(volume.VolumeName)
		data.Id = types.StringValue(volume.VolumeName)
	}
	if volume.Type != "" {
		data.Type = types.StringValue(volume.Type)
	}
	if volume.RegionID != "" {
		data.RegionId = types.StringValue(volume.RegionID)
	}
	data.StorageIntegrationId = volumeOptionalString(data.StorageIntegrationId, volume.StorageIntegrationID)
	data.Path = volumeOptionalString(data.Path, volume.Path)
	data.Status = types.StringValue(volume.Status)
	data.CreateTime = types.StringValue(volume.CreateTime)
}

func volumeOptionalString(current types.String, value string) types.String {
	if value != "" {
		return types.StringValue(value)
	}
	if current.IsNull() || current.IsUnknown() {
		return types.StringNull()
	}
	return current
}

func (r *VolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VolumeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volumeName := state.Id.ValueString()
	if volumeName == "" {
		volumeName = state.VolumeName.ValueString()
	}

	described, err := r.client.DescribeVolume(volumeName)
	if err != nil {
		if isVolumeNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read volume",
			fmt.Sprintf("volume_name=%s error=%s", volumeName, err.Error()),
		)
		return
	}

	state.applyDescribeVolume(described)
	if state.Id.IsNull() || state.Id.IsUnknown() || state.Id.ValueString() == "" {
		state.Id = types.StringValue(volumeName)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func isVolumeNotFoundError(err error) bool {
	var apiErr zilliz.Error
	if errors.As(err, &apiErr) && apiErr.Code == http.StatusNotFound {
		return true
	}

	var apiErrPtr *zilliz.Error
	if errors.As(err, &apiErrPtr) && apiErrPtr != nil && apiErrPtr.Code == http.StatusNotFound {
		return true
	}

	return strings.Contains(err.Error(), "http status code: 404")
}

func (r *VolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VolumeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VolumeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volumeName := state.Id.ValueString()
	if volumeName == "" {
		volumeName = state.VolumeName.ValueString()
	}
	if volumeName == "" {
		resp.Diagnostics.AddError(
			"Missing volume name",
			"Cannot delete volume because both id and volume_name are empty in Terraform state.",
		)
		return
	}

	_, err := r.client.DeleteVolume(volumeName)
	if err != nil {
		if isVolumeNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete volume",
			fmt.Sprintf("volume_name=%s error=%s", volumeName, err.Error()),
		)
		return
	}
}

func (r *VolumeResource) waitUntilVolumeDeleted(ctx context.Context, volumeName string) error {
	waitCtx, cancel := context.WithTimeout(ctx, volumeDeleteTimeout)
	defer cancel()

	for {
		_, err := r.client.DescribeVolume(volumeName)
		if err != nil {
			if isVolumeNotFoundError(err) {
				return nil
			}
			return fmt.Errorf("failed to check deletion status: %w", err)
		}

		timer := time.NewTimer(volumeDeletePollInterval)
		select {
		case <-waitCtx.Done():
			timer.Stop()
			return fmt.Errorf("timed out waiting for volume to be removed")
		case <-timer.C:
		}
	}
}
