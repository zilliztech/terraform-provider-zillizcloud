package provider

import (
	"context"
	"fmt"
	"strings"
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

var _ resource.ResourceWithConfigure = &ServiceAccountKeyResource{}
var _ resource.ResourceWithImportState = &ServiceAccountKeyResource{}

type ServiceAccountKeyResource struct {
	client *zilliz.Client
}

func NewServiceAccountKeyResource() resource.Resource {
	return &ServiceAccountKeyResource{}
}

type ServiceAccountKeyResourceModel struct {
	Id               types.String `tfsdk:"id"`
	ServiceAccountId types.String `tfsdk:"service_account_id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	ExpireDays       types.Int64  `tfsdk:"expire_days"`
	ExpireTime       types.String `tfsdk:"expire_time"`
	Secret           types.String `tfsdk:"secret"`
	CreatedBy        types.String `tfsdk:"created_by"`
	CreatorName      types.String `tfsdk:"creator_name"`
	CreateTime       types.String `tfsdk:"create_time"`
}

func (r *ServiceAccountKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account_api_key"
}

func (r *ServiceAccountKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an API key belonging to a Service Account. " +
			"The key inherits all account permissions. " +
			"The secret is returned only on creation and stored in Terraform state; protect your state backend. " +
			"Import cannot recover the secret. " +
			"The owning account must be active when creating or editing keys. " +
			"Changing expire_days replaces the key; use create_before_destroy when rotating credentials.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "API key ID.",
			},
			"service_account_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "Owning Service Account ID. Changes replace the key.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Validators:          serviceAccountNameValidators(),
				MarkdownDescription: "Key name, 1–128 characters, without leading or trailing whitespace.",
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
				MarkdownDescription: "Description, up to 255 characters. An empty string clears it.",
			},
			"expire_days": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 730),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				MarkdownDescription: "Validity in 24-hour days from creation, from 1 to 730. " +
					"Changes replace the key. " +
					"Imported keys with a non-integer-day or unlimited lifetime require replacement.",
			},
			"expire_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Expiration time in UTC ISO 8601 format; null for a legacy key without expiration.",
			},
			"secret": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Secret returned only when creating the key. Unavailable for imported keys.",
			},
			"created_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creator email or API key ID.",
			},
			"creator_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creator display name.",
			},
			"create_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation time in UTC ISO 8601 format.",
			},
		},
	}
}

func (r *ServiceAccountKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	var ok bool
	r.client, ok = req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider type", fmt.Sprintf("Expected *zilliz.Client, got %T", req.ProviderData))
	}
}

func (data *ServiceAccountKeyResourceModel) populate(key *zilliz.ServiceAccountKeyResponse) {
	data.Id = types.StringValue(key.ApiKeyId)
	data.ServiceAccountId = types.StringValue(key.ServiceAccountId)
	data.Name = types.StringValue(key.Name)
	data.Description = types.StringValue(key.Description)
	data.ExpireTime = types.StringPointerValue(key.ExpireTime)
	data.CreatedBy = types.StringValue(key.CreatedBy)
	data.CreatorName = types.StringValue(key.CreatorName)
	data.CreateTime = types.StringValue(key.CreateTime)
	// Metadata does not contain the secret. Preserve the existing state value.
	if data.ExpireDays.IsNull() && key.ExpireTime != nil {
		created, err1 := time.Parse(time.RFC3339Nano, key.CreateTime)
		expires, err2 := time.Parse(time.RFC3339Nano, *key.ExpireTime)
		duration := expires.Sub(created)
		if err1 == nil && err2 == nil && duration > 0 && duration%(24*time.Hour) == 0 {
			data.ExpireDays = types.Int64Value(int64(duration / (24 * time.Hour)))
		}
	}
}

func (r *ServiceAccountKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServiceAccountKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateServiceAccountKey(ctx, data.ServiceAccountId.ValueString(), &zilliz.CreateServiceAccountKeyRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		ExpireDays:  data.ExpireDays.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Service Account API key", err.Error())
		return
	}
	data.Secret = types.StringValue(created.Secret)
	data.populate(&created.ApiKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceAccountKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceAccountKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	key, err := r.client.GetServiceAccountKey(ctx, data.ServiceAccountId.ValueString(), data.Id.ValueString())
	if zilliz.IsServiceAccountKeyNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Service Account API key", err.Error())
		return
	}
	data.populate(key)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceAccountKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ServiceAccountKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	update := &zilliz.UpdateServiceAccountKeyRequest{}
	if !plan.Name.Equal(state.Name) {
		value := plan.Name.ValueString()
		update.Name = &value
	}
	if !plan.Description.Equal(state.Description) {
		value := plan.Description.ValueString()
		update.Description = &value
	}
	var key *zilliz.ServiceAccountKeyResponse
	var err error
	if update.Name != nil || update.Description != nil {
		key, err = r.client.UpdateServiceAccountKey(ctx, state.ServiceAccountId.ValueString(), state.Id.ValueString(), update)
	} else {
		key, err = r.client.GetServiceAccountKey(ctx, state.ServiceAccountId.ValueString(), state.Id.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Service Account API key", err.Error())
		return
	}
	plan.Secret = state.Secret
	plan.populate(key)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceAccountKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServiceAccountKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteServiceAccountKey(ctx, data.ServiceAccountId.ValueString(), data.Id.ValueString()); err != nil && !zilliz.IsServiceAccountKeyNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete Service Account API key", err.Error())
	}
}

func (r *ServiceAccountKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected service_account_id/api_key_id.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_account_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
