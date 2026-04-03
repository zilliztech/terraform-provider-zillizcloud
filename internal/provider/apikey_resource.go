package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ resource.Resource = &ApiKeyResource{}
var _ resource.ResourceWithConfigure = &ApiKeyResource{}
var _ resource.ResourceWithImportState = &ApiKeyResource{}
var _ resource.ResourceWithValidateConfig = &ApiKeyResource{}

func NewApiKeyResource() resource.Resource {
	return &ApiKeyResource{}
}

type ApiKeyResource struct {
	client *zilliz.Client
}

type ApiKeyProjectAccessModel struct {
	ProjectId  types.String `tfsdk:"project_id"`
	Role       types.String `tfsdk:"role"`
	AllCluster types.Bool   `tfsdk:"all_cluster"`
	ClusterIds types.List   `tfsdk:"cluster_ids"`
}

type ApiKeyResourceModel struct {
	Id           types.String               `tfsdk:"id"`
	Name         types.String               `tfsdk:"name"`
	Role         types.String               `tfsdk:"role"`
	ProjectAccess []ApiKeyProjectAccessModel `tfsdk:"project_access"`
	KeyValue     types.String               `tfsdk:"key_value"`
	CreatorName  types.String               `tfsdk:"creator_name"`
	CreatorEmail types.String               `tfsdk:"creator_email"`
	CreateTime   types.String               `tfsdk:"create_time"`
}

func (r *ApiKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *ApiKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages an API key in Zilliz Cloud.

This resource creates and manages shared API keys with scoped permissions.
The API key value is only available at creation time and stored in Terraform state.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the API key.",
			},
			"role": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: `The organization role for this API key. Valid values: "Owner", "Member", "Billing-Admin".`,
				Validators: []validator.String{
					stringvalidator.OneOf("Owner", "Member", "Billing-Admin"),
				},
			},
			"key_value": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The API key value. Only available at creation time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creator_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the API key creator.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creator_email": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The email of the API key creator.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"create_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation time in ISO 8601 format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"project_access": schema.ListNestedBlock{
				MarkdownDescription: "Project access configuration. Required when role is Member.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"project_id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The project ID to grant access to.",
						},
						"role": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: `The project role. Valid values: "Admin", "Read-Write", "Read-Only".`,
							Validators: []validator.String{
								stringvalidator.OneOf("Admin", "Read-Write", "Read-Only"),
							},
						},
						"all_cluster": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(true),
							MarkdownDescription: "Whether to include all clusters in this project.",
						},
						"cluster_ids": schema.ListAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Specific cluster IDs when all_cluster is false.",
						},
					},
				},
			},
		},
	}
}

func (r *ApiKeyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ApiKeyResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Role.ValueString() == "Member" && len(data.ProjectAccess) == 0 {
		resp.Diagnostics.AddError(
			"Missing project_access",
			`At least one "project_access" block is required when role is "Member".`,
		)
	}
}

func (r *ApiKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider type",
			fmt.Sprintf("Expected *zilliz.Client, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *ApiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ApiKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &zilliz.CreateApiKeyRequest{
		Name:    data.Name.ValueString(),
		OrgRole: data.Role.ValueString(),
	}

	for _, pa := range data.ProjectAccess {
		p := zilliz.ApiKeyProjectAccess{
			ProjectId: pa.ProjectId.ValueString(),
			Role:      pa.Role.ValueString(),
		}
		allCluster := pa.AllCluster.ValueBool()
		p.AllCluster = &allCluster
		// Default allVolume to true since the API requires it but we don't expose it in schema
		allVolume := true
		p.AllVolume = &allVolume

		if !pa.ClusterIds.IsNull() {
			var ids []string
			resp.Diagnostics.Append(pa.ClusterIds.ElementsAs(ctx, &ids, false)...)
			p.ClusterIds = ids
		}
		createReq.Projects = append(createReq.Projects, p)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	createResp, err := r.client.CreateApiKey(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API key", err.Error())
		return
	}

	data.Id = types.StringValue(createResp.ApiKeyId)
	data.KeyValue = types.StringValue(createResp.ApiKey)

	// Fetch the full API key details to populate computed fields
	apiKey, err := r.client.GetApiKey(createResp.ApiKeyId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read API key after creation", err.Error())
		return
	}

	data.CreatorName = types.StringValue(apiKey.CreatorName)
	data.CreatorEmail = types.StringValue(apiKey.CreatorEmail)
	data.CreateTime = types.StringValue(apiKey.CreateTime)
	data.Role = types.StringValue(apiKey.OrgRole)
	populateProjectAccess(&data, apiKey.Projects)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApiKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey, err := r.client.GetApiKey(state.Id.ValueString())
	if err != nil {
		var apiErr *zilliz.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read API key", err.Error())
		return
	}

	state.Name = types.StringValue(apiKey.Name)
	state.Role = types.StringValue(apiKey.OrgRole)
	state.CreatorName = types.StringValue(apiKey.CreatorName)
	state.CreatorEmail = types.StringValue(apiKey.CreatorEmail)
	state.CreateTime = types.StringValue(apiKey.CreateTime)
	populateProjectAccess(&state, apiKey.Projects)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApiKeyResourceModel
	var state ApiKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &zilliz.UpdateApiKeyRequest{
		Name:    plan.Name.ValueString(),
		OrgRole: plan.Role.ValueString(),
	}

	for _, pa := range plan.ProjectAccess {
		p := zilliz.ApiKeyProjectAccess{
			ProjectId: pa.ProjectId.ValueString(),
			Role:      pa.Role.ValueString(),
		}
		allCluster := pa.AllCluster.ValueBool()
		p.AllCluster = &allCluster
		// Default allVolume to true since the API requires it but we don't expose it in schema
		allVolume := true
		p.AllVolume = &allVolume

		if !pa.ClusterIds.IsNull() {
			var ids []string
			resp.Diagnostics.Append(pa.ClusterIds.ElementsAs(ctx, &ids, false)...)
			p.ClusterIds = ids
		}
		updateReq.Projects = append(updateReq.Projects, p)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateApiKey(state.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update API key", err.Error())
		return
	}

	state.Name = types.StringValue(updated.Name)
	state.Role = types.StringValue(updated.OrgRole)
	state.CreatorName = types.StringValue(updated.CreatorName)
	state.CreatorEmail = types.StringValue(updated.CreatorEmail)
	state.CreateTime = types.StringValue(updated.CreateTime)
	populateProjectAccess(&state, updated.Projects)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApiKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteApiKey(state.Id.ValueString())
	if err != nil {
		var apiErr *zilliz.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return
		}
		resp.Diagnostics.AddError("Failed to delete API key", err.Error())
	}
}

func (r *ApiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	apiKeyId := req.ID

	apiKey, err := r.client.GetApiKey(apiKeyId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import API key", fmt.Sprintf("API key ID: %s, error: %s", apiKeyId, err.Error()))
		return
	}

	var state ApiKeyResourceModel
	state.Id = types.StringValue(apiKeyId)
	state.Name = types.StringValue(apiKey.Name)
	state.Role = types.StringValue(apiKey.OrgRole)
	state.KeyValue = types.StringValue("")
	state.CreatorName = types.StringValue(apiKey.CreatorName)
	state.CreatorEmail = types.StringValue(apiKey.CreatorEmail)
	state.CreateTime = types.StringValue(apiKey.CreateTime)
	populateProjectAccess(&state, apiKey.Projects)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func populateProjectAccess(data *ApiKeyResourceModel, projects []zilliz.ApiKeyProjectResponse) {
	if len(projects) == 0 {
		data.ProjectAccess = nil
		return
	}
	access := make([]ApiKeyProjectAccessModel, 0, len(projects))
	for _, p := range projects {
		pa := ApiKeyProjectAccessModel{
			ProjectId:  types.StringValue(p.ProjectId),
			Role:       types.StringValue(p.Role),
			AllCluster: types.BoolValue(p.AllCluster),
		}

		if len(p.Clusters) > 0 {
			clusterIds := make([]attr.Value, 0, len(p.Clusters))
			for _, c := range p.Clusters {
				clusterIds = append(clusterIds, types.StringValue(c.ClusterId))
			}
			pa.ClusterIds, _ = types.ListValue(types.StringType, clusterIds)
		} else {
			pa.ClusterIds = types.ListNull(types.StringType)
		}

		access = append(access, pa)
	}
	data.ProjectAccess = access
}
