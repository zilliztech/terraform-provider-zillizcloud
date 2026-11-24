package provider

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ resource.ResourceWithConfigure = &ServiceAccountResource{}
var _ resource.ResourceWithImportState = &ServiceAccountResource{}
var _ resource.ResourceWithValidateConfig = &ServiceAccountResource{}

type ServiceAccountResource struct {
	client *zilliz.Client
}

func NewServiceAccountResource() resource.Resource {
	return &ServiceAccountResource{}
}

type ServiceAccountResourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Status             types.String `tfsdk:"status"`
	ProjectScope       types.String `tfsdk:"project_scope"`
	OrgRoleId          types.String `tfsdk:"org_role_id"`
	ProjectRoleIds     types.Set    `tfsdk:"project_role_ids"`
	ProjectAssignments types.Map    `tfsdk:"project_assignments"`
	CreatedBy          types.String `tfsdk:"created_by"`
	CreatorName        types.String `tfsdk:"creator_name"`
	CreateTime         types.String `tfsdk:"create_time"`
}

var serviceAccountRoleSetType = types.SetType{ElemType: types.StringType}

func (r *ServiceAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *ServiceAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 1,
		MarkdownDescription: "Manages an organization or project-local Service Account, including all its roles and project assignments. " +
			"Organization accounts use exactly one of org_role_id or project_assignments. " +
			"Project-local accounts use project_scope and optional project_role_ids. " +
			"Delete all associated API keys before deleting the account. " +
			"A disabled account must be enabled before editing metadata or permissions; use a separate apply to enable it before creating or updating keys.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Service Account ID.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Validators:          serviceAccountNameValidators(),
				MarkdownDescription: "Unique account name, 1–128 characters, without leading or trailing whitespace.",
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
				MarkdownDescription: "Description. An empty string clears it.",
			},
			"status": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("active"),
				Validators: []validator.String{
					stringvalidator.OneOf("active", "disable"),
				},
				MarkdownDescription: "Account status: active or disable. Disabling makes all associated keys unusable.",
			},
			"project_scope": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
				MarkdownDescription: "Owning project ID for a project-local account. Omit for an organization account. Changes replace the account.",
			},
			"org_role_id": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Predefined Billing Admin or Org Owner role ID. " +
					"Mutually exclusive with project_assignments and project_scope. " +
					"To clear all access, omit this field and set project_assignments = {}.",
			},
			"project_assignments": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				Default:     mapdefault.StaticValue(types.MapValueMust(serviceAccountRoleSetType, map[string]attr.Value{})),
				ElementType: serviceAccountRoleSetType,
				MarkdownDescription: "Complete map of project IDs to sets of project role IDs for an organization account. " +
					"Use an empty set to assign a project without roles, or {} for no assignments. " +
					"Removing a map entry unassigns that project. " +
					"Mutually exclusive with org_role_id and project_scope.",
			},
			"project_role_ids": schema.SetAttribute{
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				ElementType:         types.StringType,
				MarkdownDescription: "Complete role ID set in project_scope for a project-local account. Omitted or [] grants no roles.",
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

func serviceAccountNameValidators() []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(1, 128),
		stringvalidator.RegexMatches(regexp.MustCompile(`^[^\s\p{Z}\x{0085}](?:[\s\S]*[^\s\p{Z}\x{0085}])?$`), "must not contain leading or trailing whitespace"),
	}
}

func (r *ServiceAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	var ok bool
	r.client, ok = req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider type", fmt.Sprintf("Expected *zilliz.Client, got %T", req.ProviderData))
	}
}

func (r *ServiceAccountResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ServiceAccountResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !data.ProjectScope.IsNull() {
		if !data.OrgRoleId.IsNull() || !data.ProjectAssignments.IsNull() {
			resp.Diagnostics.AddError("Invalid project-local permissions", "Use only project_role_ids with project_scope.")
		}
	} else {
		if data.OrgRoleId.IsNull() == data.ProjectAssignments.IsNull() {
			resp.Diagnostics.AddError(
				"Missing or conflicting organization permissions",
				"Specify exactly one of org_role_id or project_assignments. Use project_assignments = {} for no access.",
			)
		}
		if !data.ProjectRoleIds.IsNull() {
			resp.Diagnostics.AddError("Invalid organization permissions", "project_role_ids requires project_scope.")
		}
	}
	if !data.OrgRoleId.IsNull() {
		validateServiceAccountRole(data.OrgRoleId, &resp.Diagnostics)
	}
	validateServiceAccountRoles(data.ProjectRoleIds, &resp.Diagnostics)
	projects := map[string]bool{}
	for id, value := range data.ProjectAssignments.Elements() {
		normalized := strings.ToLower(id)
		if id == "" || strings.TrimSpace(id) != id || len(id) > 32 || strings.ContainsAny(id, ",\x00") || projects[normalized] {
			resp.Diagnostics.AddError(
				"Invalid project assignment",
				"Project IDs must be nonblank, at most 32 bytes, and unique ignoring case, without surrounding whitespace, commas or NUL.",
			)
		}
		projects[normalized] = true
		roles, ok := value.(types.Set)
		if !ok {
			resp.Diagnostics.AddError("Invalid project roles", "Expected a set of role IDs.")
			continue
		}
		if roles.IsNull() {
			resp.Diagnostics.AddError("Invalid project roles", "Project role sets cannot be null. Use [] for no roles.")
		}
		validateServiceAccountRoles(roles, &resp.Diagnostics)
	}
}

func validateServiceAccountRoles(roles types.Set, diagnostics *diag.Diagnostics) {
	for _, value := range roles.Elements() {
		role, ok := value.(types.String)
		if !ok {
			diagnostics.AddError("Invalid role ID", "Expected a string role ID.")
			continue
		}
		validateServiceAccountRole(role, diagnostics)
	}
}

func validateServiceAccountRole(role types.String, diagnostics *diag.Diagnostics) {
	if role.IsUnknown() {
		return
	}
	id := role.ValueString()
	if role.IsNull() || id == "" || len(id) > 64 || strings.TrimSpace(id) != id || strings.ContainsRune(id, '\x00') {
		diagnostics.AddError("Invalid role ID", "Role IDs must be nonblank strings of at most 64 bytes without surrounding whitespace or NUL.")
	}
}

func roleIDs(roles []zilliz.ServiceAccountRole) []string {
	ids := make([]string, 0, len(roles))
	for _, role := range roles {
		ids = append(ids, role.RoleId)
	}
	sort.Strings(ids)
	return ids
}

func serviceAccountRoles(ctx context.Context, value attr.Value, diagnostics *diag.Diagnostics) []string {
	ids := []string{}
	roles, ok := value.(types.Set)
	if !ok {
		diagnostics.AddError("Invalid roles", "Expected a set of role IDs.")
		return ids
	}
	if !roles.IsNull() {
		diagnostics.Append(roles.ElementsAs(ctx, &ids, false)...)
	}
	sort.Strings(ids)
	return ids
}

func serviceAccountAssignments(ctx context.Context, value types.Map, diagnostics *diag.Diagnostics) []zilliz.ServiceAccountProjectRoles {
	assignments := make([]zilliz.ServiceAccountProjectRoles, 0, len(value.Elements()))
	for id, roles := range value.Elements() {
		assignments = append(assignments, zilliz.ServiceAccountProjectRoles{
			ProjectId:      id,
			ProjectRoleIds: serviceAccountRoles(ctx, roles, diagnostics),
		})
	}
	sort.Slice(assignments, func(i, j int) bool { return assignments[i].ProjectId < assignments[j].ProjectId })
	return assignments
}

func (data *ServiceAccountResourceModel) populate(ctx context.Context, account *zilliz.ServiceAccountResponse) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	// Canonical empty collections keep imported and configured permissions identical.
	data.ProjectRoleIds = types.SetValueMust(types.StringType, []attr.Value{})
	data.Id = types.StringValue(account.ServiceAccountId)
	data.Name = types.StringValue(account.Name)
	data.Description = types.StringValue(account.Description)
	data.Status = types.StringValue(account.Status)
	data.ProjectScope = types.StringPointerValue(account.ProjectScope)
	data.CreatedBy = types.StringValue(account.CreatedBy)
	data.CreatorName = types.StringValue(account.CreatorName)
	data.CreateTime = types.StringValue(account.CreateTime)
	data.OrgRoleId = types.StringNull()
	if len(account.OrgRoles) > 1 {
		diagnostics.AddError("Unexpected organization roles", "The Service Account API returned multiple organization roles; only one organization role is supported.")
		return diagnostics
	}
	if len(account.OrgRoles) == 1 {
		data.OrgRoleId = types.StringValue(account.OrgRoles[0].RoleId)
	}
	assignments := map[string]attr.Value{}
	for _, assignment := range account.ProjectAssignments {
		roles, d := types.SetValueFrom(ctx, types.StringType, roleIDs(assignment.ProjectRoles))
		diagnostics.Append(d...)
		if account.ProjectScope != nil && assignment.ProjectId == *account.ProjectScope {
			data.ProjectRoleIds = roles
		} else {
			assignments[assignment.ProjectId] = roles
		}
	}
	value, d := types.MapValue(serviceAccountRoleSetType, assignments)
	diagnostics.Append(d...)
	data.ProjectAssignments = value
	return diagnostics
}

func (r *ServiceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServiceAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var account *zilliz.ServiceAccountResponse
	var err error
	if !data.ProjectScope.IsNull() {
		roles := serviceAccountRoles(ctx, data.ProjectRoleIds, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		account, err = r.client.CreateProjectServiceAccount(ctx, data.ProjectScope.ValueString(), &zilliz.CreateProjectServiceAccountRequest{
			Name:           data.Name.ValueString(),
			Description:    data.Description.ValueString(),
			ProjectRoleIds: roles,
		})
	} else {
		create := &zilliz.CreateServiceAccountRequest{
			Name:        data.Name.ValueString(),
			Description: data.Description.ValueString(),
		}
		if !data.OrgRoleId.IsNull() {
			create.OrgRoleId = data.OrgRoleId.ValueStringPointer()
		} else {
			assignments := serviceAccountAssignments(ctx, data.ProjectAssignments, &resp.Diagnostics)
			create.ProjectAssignments = &assignments
		}
		if resp.Diagnostics.HasError() {
			return
		}
		account, err = r.client.CreateServiceAccount(ctx, create)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Service Account", err.Error())
		return
	}
	desiredStatus := data.Status.ValueString()
	resp.Diagnostics.Append(data.populate(ctx, account)...)
	// Save the identity before the second mutation so a failed disable is recoverable.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if desiredStatus != account.Status {
		account, err = r.client.UpdateServiceAccount(ctx, account.ServiceAccountId, &zilliz.UpdateServiceAccountRequest{Status: &desiredStatus})
		if err != nil {
			resp.Diagnostics.AddError("Failed to set Service Account status", err.Error())
			return
		}
		resp.Diagnostics.Append(data.populate(ctx, account)...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func (r *ServiceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, err := r.client.GetServiceAccount(ctx, data.Id.ValueString())
	if zilliz.IsServiceAccountNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Service Account", err.Error())
		return
	}
	resp.Diagnostics.Append(data.populate(ctx, account)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ServiceAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.updateAccount(ctx, &plan, &state, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Service Account", err.Error())
	}
	// Refresh actual permissions after partial failure; never record an unapplied plan.
	account, readErr := r.client.GetServiceAccount(ctx, state.Id.ValueString())
	if readErr != nil {
		resp.Diagnostics.AddError("Failed to refresh Service Account", readErr.Error())
		return
	}
	resp.Diagnostics.Append(plan.populate(ctx, account)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceAccountResource) updateAccount(ctx context.Context, plan, state *ServiceAccountResourceModel, diagnostics *diag.Diagnostics) error {
	id := state.Id.ValueString()
	metadataChanged := !plan.Name.Equal(state.Name) || !plan.Description.Equal(state.Description)
	rolesChanged := !plan.OrgRoleId.Equal(state.OrgRoleId) || !plan.ProjectRoleIds.Equal(state.ProjectRoleIds) || !plan.ProjectAssignments.Equal(state.ProjectAssignments)
	if state.Status.ValueString() == "disable" {
		if plan.Status.ValueString() != "active" && (metadataChanged || rolesChanged) {
			return fmt.Errorf("enable this account before changing its metadata or permissions")
		}
		if plan.Status.ValueString() == "active" {
			status := "active"
			if _, err := r.client.UpdateServiceAccount(ctx, id, &zilliz.UpdateServiceAccountRequest{Status: &status}); err != nil {
				return err
			}
		}
	}
	if metadataChanged {
		update := &zilliz.UpdateServiceAccountRequest{}
		if !plan.Name.Equal(state.Name) {
			value := plan.Name.ValueString()
			update.Name = &value
		}
		if !plan.Description.Equal(state.Description) {
			value := plan.Description.ValueString()
			update.Description = &value
		}
		if _, err := r.client.UpdateServiceAccount(ctx, id, update); err != nil {
			return err
		}
	}
	if rolesChanged {
		if err := r.updateAccountAccess(ctx, plan, state, diagnostics); err != nil {
			return err
		}
	}
	if plan.Status.ValueString() == "disable" && state.Status.ValueString() != "disable" {
		status := "disable"
		_, err := r.client.UpdateServiceAccount(ctx, id, &zilliz.UpdateServiceAccountRequest{Status: &status})
		return err
	}
	return nil
}

func (r *ServiceAccountResource) updateAccountAccess(ctx context.Context, plan, state *ServiceAccountResourceModel, diagnostics *diag.Diagnostics) error {
	id := state.Id.ValueString()
	var projectUpdates []zilliz.ServiceAccountProjectRoles
	if !plan.ProjectScope.IsNull() {
		projectUpdates = []zilliz.ServiceAccountProjectRoles{{
			ProjectId:      plan.ProjectScope.ValueString(),
			ProjectRoleIds: serviceAccountRoles(ctx, plan.ProjectRoleIds, diagnostics),
		}}
	} else {
		desired := plan.ProjectAssignments.Elements()
		previous := state.ProjectAssignments.Elements()
		projectRolesOnly := plan.OrgRoleId.IsNull() && state.OrgRoleId.IsNull() && len(desired) == len(previous)
		for project := range desired {
			if _, exists := previous[project]; !exists {
				projectRolesOnly = false
				break
			}
		}
		if !projectRolesOnly {
			// Scope switches and assignment changes replace the complete access configuration.
			access := &zilliz.ReplaceServiceAccountAccessRequest{}
			if !plan.OrgRoleId.IsNull() {
				access.OrgRoleId = plan.OrgRoleId.ValueStringPointer()
			} else {
				assignments := serviceAccountAssignments(ctx, plan.ProjectAssignments, diagnostics)
				access.ProjectAssignments = &assignments
			}
			if diagnostics.HasError() {
				return fmt.Errorf("invalid project assignments")
			}
			_, err := r.client.ReplaceServiceAccountAccess(ctx, id, access)
			return err
		}
		// Existing project roles can be edited with project-scoped permissions.
		// Sorted assignments keep multiple PATCH requests in a deterministic order.
		for _, assignment := range serviceAccountAssignments(ctx, plan.ProjectAssignments, diagnostics) {
			if !desired[assignment.ProjectId].Equal(previous[assignment.ProjectId]) {
				projectUpdates = append(projectUpdates, assignment)
			}
		}
	}
	if diagnostics.HasError() {
		return fmt.Errorf("invalid project roles")
	}
	for _, assignment := range projectUpdates {
		if _, err := r.client.UpdateServiceAccountProjectRoles(ctx, id, assignment.ProjectId, &zilliz.UpdateServiceAccountProjectRolesRequest{
			ProjectRoleIds: assignment.ProjectRoleIds,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *ServiceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServiceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteServiceAccount(ctx, data.Id.ValueString()); err != nil && !zilliz.IsServiceAccountNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete Service Account", err.Error())
	}
}

func (r *ServiceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
