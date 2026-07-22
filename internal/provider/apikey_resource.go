package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ resource.Resource = &ApiKeyResource{}
var _ resource.ResourceWithConfigure = &ApiKeyResource{}
var _ resource.ResourceWithImportState = &ApiKeyResource{}
var _ resource.ResourceWithValidateConfig = &ApiKeyResource{}
var _ resource.ResourceWithModifyPlan = &ApiKeyResource{}
var _ resource.ResourceWithUpgradeState = &ApiKeyResource{}

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
	AllVolume  types.Bool   `tfsdk:"all_volume"`
	VolumeIds  types.List   `tfsdk:"volume_ids"`
}

type ApiKeyResourceModel struct {
	Id            types.String               `tfsdk:"id"`
	Name          types.String               `tfsdk:"name"`
	Description   types.String               `tfsdk:"description"`
	Role          types.String               `tfsdk:"role"`
	ProjectAccess []ApiKeyProjectAccessModel `tfsdk:"project_access"`
	KeyValue      types.String               `tfsdk:"key_value"`
	CreatorName   types.String               `tfsdk:"creator_name"`
	CreatedBy     types.String               `tfsdk:"created_by"`
	CreateTime    types.String               `tfsdk:"create_time"`
}

func (r *ApiKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

// apiKeyRoleValidator validates org/project role strings.
// FC1: passes through unknown roles with a warning; only hard-blocks "Owner" on create/update.
type apiKeyRoleValidator struct {
	known     []string
	blockRole string // if non-empty, hard-block this value
}

func (v apiKeyRoleValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("role string; known values: %v", v.known)
}

func (v apiKeyRoleValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v apiKeyRoleValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()

	if v.blockRole != "" && val == v.blockRole {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid role value",
			fmt.Sprintf("%q keys cannot be created or updated via API. Use the Zilliz Cloud Console instead.", v.blockRole),
		)
		return
	}

	for _, k := range v.known {
		if val == k {
			return // known value, no warning
		}
	}

	// FC1: unknown role — pass through with warning
	resp.Diagnostics.AddAttributeWarning(
		req.Path,
		"Unrecognized role value",
		fmt.Sprintf("Role %q is not in the known set %v. It will be passed through to the API as-is. "+
			"If the API rejects it, the apply will fail.", val, v.known),
	)
}

func (r *ApiKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 1, // bumped from 0 for all_volume/volume_ids restore (G2)
		MarkdownDescription: `Manages a Customized API key in Zilliz Cloud.

This resource creates and manages organization-scoped API keys with project-level permissions.
The API key value is only available at creation time and stored in Terraform state.

**Important:** API key management requires an Org Owner API key. A project-scoped key cannot
manage API keys. See the bootstrap pattern in the provider documentation.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the API key (apiKeyId).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the API key.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Description of the API key.",
			},
			"role": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `The organization role for this API key. Known values: "Member", "Billing-Admin".

**Note:** "Owner" keys cannot be created or updated via API — use the Console instead.
"Owner" is accepted for import of Console-created keys. Unknown role values are passed
through to the API (forward compatibility).`,
				Validators: []validator.String{
					apiKeyRoleValidator{
						known:     []string{"Member", "Billing-Admin"},
						blockRole: "Owner",
					},
				},
			},
			"key_value": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The API key secret. Only available at creation time; not retrievable after.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creator_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The display name of the API key creator.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The creator identifier (email address or API key ID).",
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
			"project_access": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Project access configuration. Required when role is `Member`.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"project_id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The project ID to grant access to.",
						},
						"role": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: `The project role. Known values: "Admin", "Read-Write", "Read-Only".`,
							Validators: []validator.String{
								apiKeyRoleValidator{
									known: []string{"Admin", "Read-Write", "Read-Only"},
								},
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
							MarkdownDescription: "Specific cluster IDs when `all_cluster` is false.",
						},
						"all_volume": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(true),
							MarkdownDescription: "Whether to include all volumes in this project.",
						},
						"volume_ids": schema.ListAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Specific volume IDs when `all_volume` is false.",
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
			`At least one "project_access" entry is required when role is "Member".`,
		)
	}

	for _, pa := range data.ProjectAccess {
		// In config phase, all_cluster is null when user omitted it (Default
		// hasn't been applied yet). A non-null true means user explicitly
		// wrote all_cluster = true, which conflicts with cluster_ids.
		if !pa.AllCluster.IsNull() && pa.AllCluster.ValueBool() && !pa.ClusterIds.IsNull() {
			resp.Diagnostics.AddError(
				"Conflicting configuration",
				`Cannot set all_cluster = true and cluster_ids at the same time. `+
					`Either remove all_cluster (or set it to false) when using cluster_ids, `+
					`or remove cluster_ids when using all_cluster = true.`,
			)
		}
		// Same validation for volumes
		if !pa.AllVolume.IsNull() && pa.AllVolume.ValueBool() && !pa.VolumeIds.IsNull() {
			resp.Diagnostics.AddError(
				"Conflicting configuration",
				`Cannot set all_volume = true and volume_ids at the same time. `+
					`Either remove all_volume (or set it to false) when using volume_ids, `+
					`or remove volume_ids when using all_volume = true.`,
			)
		}
	}
}

func (r *ApiKeyResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy — nothing to do.
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan ApiKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	modified := false
	for i := range plan.ProjectAccess {
		// Auto-correct all_cluster when cluster_ids is provided
		if !plan.ProjectAccess[i].ClusterIds.IsNull() && !plan.ProjectAccess[i].ClusterIds.IsUnknown() {
			var ids []string
			resp.Diagnostics.Append(plan.ProjectAccess[i].ClusterIds.ElementsAs(ctx, &ids, false)...)
			if len(ids) > 0 && plan.ProjectAccess[i].AllCluster.ValueBool() {
				plan.ProjectAccess[i].AllCluster = types.BoolValue(false)
				modified = true
			}
		}
		// Auto-correct all_volume when volume_ids is provided
		if !plan.ProjectAccess[i].VolumeIds.IsNull() && !plan.ProjectAccess[i].VolumeIds.IsUnknown() {
			var ids []string
			resp.Diagnostics.Append(plan.ProjectAccess[i].VolumeIds.ElementsAs(ctx, &ids, false)...)
			if len(ids) > 0 && plan.ProjectAccess[i].AllVolume.ValueBool() {
				plan.ProjectAccess[i].AllVolume = types.BoolValue(false)
				modified = true
			}
		}
	}

	if modified {
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
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
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		OrgRole:     data.Role.ValueString(),
	}

	for _, pa := range data.ProjectAccess {
		p := zilliz.ApiKeyProjectAccess{
			ProjectId: pa.ProjectId.ValueString(),
			Role:      pa.Role.ValueString(),
		}
		allCluster := pa.AllCluster.ValueBool()
		p.AllCluster = &allCluster
		allVolume := pa.AllVolume.ValueBool()
		p.AllVolume = &allVolume

		if !pa.ClusterIds.IsNull() {
			var ids []string
			resp.Diagnostics.Append(pa.ClusterIds.ElementsAs(ctx, &ids, false)...)
			p.ClusterIds = ids
		}
		if !pa.VolumeIds.IsNull() {
			var ids []string
			resp.Diagnostics.Append(pa.VolumeIds.ElementsAs(ctx, &ids, false)...)
			p.VolumeIds = ids
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
	data.CreatedBy = types.StringValue(apiKey.CreatedBy)
	data.CreateTime = types.StringValue(apiKey.CreateTime)
	data.Role = types.StringValue(apiKey.OrgRole)
	data.Description = types.StringValue(apiKey.Description)
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
		// FC4: accept both 404 and 80001 as not-found
		if zilliz.IsApiKeyNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read API key", err.Error())
		return
	}

	state.Name = types.StringValue(apiKey.Name)
	state.Description = types.StringValue(apiKey.Description)
	state.Role = types.StringValue(apiKey.OrgRole)
	state.CreatorName = types.StringValue(apiKey.CreatorName)
	state.CreatedBy = types.StringValue(apiKey.CreatedBy)
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

	// FC5: always send the full desired state
	updateReq := &zilliz.UpdateApiKeyRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		OrgRole:     plan.Role.ValueString(),
	}

	for _, pa := range plan.ProjectAccess {
		p := zilliz.ApiKeyProjectAccess{
			ProjectId: pa.ProjectId.ValueString(),
			Role:      pa.Role.ValueString(),
		}
		allCluster := pa.AllCluster.ValueBool()
		p.AllCluster = &allCluster
		allVolume := pa.AllVolume.ValueBool()
		p.AllVolume = &allVolume

		if !pa.ClusterIds.IsNull() {
			var ids []string
			resp.Diagnostics.Append(pa.ClusterIds.ElementsAs(ctx, &ids, false)...)
			p.ClusterIds = ids
		}
		if !pa.VolumeIds.IsNull() {
			var ids []string
			resp.Diagnostics.Append(pa.VolumeIds.ElementsAs(ctx, &ids, false)...)
			p.VolumeIds = ids
		}
		updateReq.Projects = append(updateReq.Projects, p)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// FC5: nil slice serializes as "projects": null which may differ from
	// "projects": [] on the API side. Always send an explicit empty array.
	if updateReq.Projects == nil {
		updateReq.Projects = []zilliz.ApiKeyProjectAccess{}
	}

	updated, err := r.client.UpdateApiKey(state.Id.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update API key", err.Error())
		return
	}

	state.Name = types.StringValue(updated.Name)
	state.Description = types.StringValue(updated.Description)
	state.Role = types.StringValue(updated.OrgRole)
	state.CreatorName = types.StringValue(updated.CreatorName)
	state.CreatedBy = types.StringValue(updated.CreatedBy)
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
		// FC4: not-found on delete is a success (externally deleted)
		if zilliz.IsApiKeyNotFound(err) {
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
	state.Description = types.StringValue(apiKey.Description)
	state.Role = types.StringValue(apiKey.OrgRole)
	state.KeyValue = types.StringValue("")
	state.CreatorName = types.StringValue(apiKey.CreatorName)
	state.CreatedBy = types.StringValue(apiKey.CreatedBy)
	state.CreateTime = types.StringValue(apiKey.CreateTime)
	populateProjectAccess(&state, apiKey.Projects)

	resp.Diagnostics.AddWarning(
		"API key value unavailable",
		"The API key secret is only returned at creation time. The key_value attribute will be empty for imported keys.",
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// UpgradeState implements schema version 0 → 1.
// Adds all_volume (default true) and volume_ids (null) to each project_access entry,
// matching the hardcoded allVolume=true that v0 sent. Result: upgrade is a silent no-op plan.
func (r *ApiKeyResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			// v0 schema: project_access without all_volume/volume_ids
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id":           schema.StringAttribute{Computed: true},
					"name":         schema.StringAttribute{Required: true},
					"role":         schema.StringAttribute{Required: true},
					"key_value":    schema.StringAttribute{Computed: true, Sensitive: true},
					"creator_name": schema.StringAttribute{Computed: true},
					"created_by":   schema.StringAttribute{Computed: true},
					"create_time":  schema.StringAttribute{Computed: true},
					"project_access": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"project_id":  schema.StringAttribute{Required: true},
								"role":        schema.StringAttribute{Optional: true, Computed: true},
								"all_cluster": schema.BoolAttribute{Optional: true, Computed: true},
								"cluster_ids": schema.ListAttribute{Optional: true, ElementType: types.StringType},
							},
						},
					},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				// Read old state into a generic structure
				type oldProjectAccess struct {
					ProjectId  types.String `tfsdk:"project_id"`
					Role       types.String `tfsdk:"role"`
					AllCluster types.Bool   `tfsdk:"all_cluster"`
					ClusterIds types.List   `tfsdk:"cluster_ids"`
				}
				type oldModel struct {
					Id            types.String      `tfsdk:"id"`
					Name          types.String      `tfsdk:"name"`
					Role          types.String      `tfsdk:"role"`
					KeyValue      types.String      `tfsdk:"key_value"`
					CreatorName   types.String      `tfsdk:"creator_name"`
					CreatedBy     types.String      `tfsdk:"created_by"`
					CreateTime    types.String      `tfsdk:"create_time"`
					ProjectAccess []oldProjectAccess `tfsdk:"project_access"`
				}

				var old oldModel
				resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
				if resp.Diagnostics.HasError() {
					return
				}

				// Build new state with all_volume=true, volume_ids=null, description=""
				var newAccess []ApiKeyProjectAccessModel
				for _, pa := range old.ProjectAccess {
					newAccess = append(newAccess, ApiKeyProjectAccessModel{
						ProjectId:  pa.ProjectId,
						Role:       pa.Role,
						AllCluster: pa.AllCluster,
						ClusterIds: pa.ClusterIds,
						AllVolume:  types.BoolValue(true),
						VolumeIds:  types.ListNull(types.StringType),
					})
				}

				newState := ApiKeyResourceModel{
					Id:            old.Id,
					Name:          old.Name,
					Description:   types.StringValue(""), // v0 didn't have description
					Role:          old.Role,
					ProjectAccess: newAccess,
					KeyValue:      old.KeyValue,
					CreatorName:   old.CreatorName,
					CreatedBy:     old.CreatedBy,
					CreateTime:    old.CreateTime,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
			},
		},
	}
}

// populateProjectAccess fills state from API response in an order-stable way (G6).
// It preserves the order from existing state/plan, appending any new projects sorted by ID.
func populateProjectAccess(data *ApiKeyResourceModel, projects []zilliz.ApiKeyProjectResponse) {
	if len(projects) == 0 {
		data.ProjectAccess = nil
		return
	}

	// Build lookup map from API response
	respMap := make(map[string]zilliz.ApiKeyProjectResponse, len(projects))
	for _, p := range projects {
		respMap[p.ProjectId] = p
	}

	// Preserve existing state order
	var result []ApiKeyProjectAccessModel
	seen := make(map[string]bool)

	for _, existing := range data.ProjectAccess {
		pid := existing.ProjectId.ValueString()
		if p, ok := respMap[pid]; ok {
			result = append(result, projectResponseToModel(p, &existing))
			seen[pid] = true
		}
		// If the project was removed from API response, drop it from state
	}

	// Append any new projects not in prior state, sorted by ID for determinism
	var newPids []string
	for pid := range respMap {
		if !seen[pid] {
			newPids = append(newPids, pid)
		}
	}
	sort.Strings(newPids)
	for _, pid := range newPids {
		p := respMap[pid]
		result = append(result, projectResponseToModel(p, nil))
	}

	data.ProjectAccess = result
}

// projectResponseToModel converts an API project response to a state model.
// If prior is non-nil, inner list ordering (cluster_ids, volume_ids) is preserved
// from prior state for stability.
func projectResponseToModel(p zilliz.ApiKeyProjectResponse, prior *ApiKeyProjectAccessModel) ApiKeyProjectAccessModel {
	pa := ApiKeyProjectAccessModel{
		ProjectId:  types.StringValue(p.ProjectId),
		Role:       types.StringValue(p.Role),
		AllCluster: types.BoolValue(p.AllCluster),
		AllVolume:  types.BoolValue(p.AllVolume),
	}

	pa.ClusterIds = idsFromResponse(p.Clusters, prior, func(c zilliz.ApiKeyClusterResponse) string { return c.ClusterId }, extractClusterIds)
	pa.VolumeIds = idsFromResponse(p.Volumes, prior, func(v zilliz.ApiKeyVolumeResponse) string { return v.VolumeId }, extractVolumeIds)

	return pa
}

// idsFromResponse builds an order-stable list of IDs from the API response.
func idsFromResponse[T any](
	items []T,
	prior *ApiKeyProjectAccessModel,
	getID func(T) string,
	extractPrior func(*ApiKeyProjectAccessModel) []string,
) types.List {
	if len(items) == 0 {
		return types.ListNull(types.StringType)
	}

	// Get IDs from response
	respIDs := make(map[string]bool, len(items))
	for _, item := range items {
		respIDs[getID(item)] = true
	}

	var ordered []string
	seen := make(map[string]bool)

	// Preserve prior order for known IDs
	if prior != nil {
		for _, pid := range extractPrior(prior) {
			if respIDs[pid] {
				ordered = append(ordered, pid)
				seen[pid] = true
			}
		}
	}

	// Append new IDs sorted
	var newIDs []string
	for _, item := range items {
		id := getID(item)
		if !seen[id] {
			newIDs = append(newIDs, id)
		}
	}
	sort.Strings(newIDs)
	ordered = append(ordered, newIDs...)

	vals := make([]attr.Value, 0, len(ordered))
	for _, id := range ordered {
		vals = append(vals, types.StringValue(id))
	}
	list, _ := types.ListValue(types.StringType, vals)
	return list
}

func extractClusterIds(pa *ApiKeyProjectAccessModel) []string {
	if pa == nil || pa.ClusterIds.IsNull() || pa.ClusterIds.IsUnknown() {
		return nil
	}
	var ids []string
	pa.ClusterIds.ElementsAs(context.Background(), &ids, false)
	return ids
}

func extractVolumeIds(pa *ApiKeyProjectAccessModel) []string {
	if pa == nil || pa.VolumeIds.IsNull() || pa.VolumeIds.IsUnknown() {
		return nil
	}
	var ids []string
	pa.VolumeIds.ElementsAs(context.Background(), &ids, false)
	return ids
}

// Verify interfaces are satisfied at compile time.
var _ validator.String = apiKeyRoleValidator{}
