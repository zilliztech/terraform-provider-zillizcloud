package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ resource.Resource = &UserRoleResource{}
var _ resource.ResourceWithConfigure = &UserRoleResource{}
var _ resource.ResourceWithImportState = &UserRoleResource{}

func NewUserRoleResource() resource.Resource {
	return &UserRoleResource{}
}

type UserRoleResource struct {
	client *zilliz.Client
}

type UserRoleResourceModel struct {
	Id             types.String   `tfsdk:"id"` // /connection/{connection_id}/user/{username}/roles
	ConnectAddress types.String   `tfsdk:"connect_address"`
	Username       types.String   `tfsdk:"username"`
	Roles          []types.String `tfsdk:"roles"`
}

func (r *UserRoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_role"
}

func (r *UserRoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages roles assigned to a user in a specific cluster",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connect_address": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Required: true,
			},
			"roles": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *UserRoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// BuildUserRoleID returns the RESTful ID for a user role resource.
func BuildUserRoleID(connectionID, username string) string {
	return fmt.Sprintf("/connection/%s/user/%s/roles", connectionID, username)
}

// ParseUserRoleID parses the RESTful ID and returns connectionID, username, and role.
func ParseUserRoleID(id string) (connectionID, username string, ok bool) {
	parts := strings.Split(id, "/")
	if len(parts) != 6 || parts[1] != "connection" || parts[3] != "user" || parts[5] != "roles" {
		return "", "", false
	}
	return parts[2], parts[4], true
}

func (r *UserRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionID := NormalizeConnectionID(data.ConnectAddress.ValueString())

	client, err := r.client.User(data.ConnectAddress.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get user client",
			fmt.Sprintf("ConnectAddress: %s, error: %s", data.ConnectAddress.ValueString(), err.Error()),
		)
		return
	}

	for _, role := range data.Roles {
		err := client.GrantRoleToUser(&zilliz.UserGrantRoleToUserParams{
			UserName: data.Username.ValueString(),
			RoleName: role.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to grant role to user",
				fmt.Sprintf("ConnectAddress: %s, Username: %s, Role: %s, error: %s", data.ConnectAddress.ValueString(), data.Username.ValueString(), role.ValueString(), err.Error()),
			)
			return
		}
	}

	// Set ID to /connection/{connection_id}/user/{username}/roles
	if len(data.Roles) > 0 {
		data.Id = types.StringValue(BuildUserRoleID(connectionID, data.Username.ValueString()))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.User(state.ConnectAddress.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("User Client Error", err.Error())
		return
	}

	roles, err := client.DescribeUser(state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("List Roles Error", err.Error())
		return
	}

	state.Roles = make([]types.String, 0, len(roles))
	for _, role := range roles {
		state.Roles = append(state.Roles, types.StringValue(role))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *UserRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.User(state.ConnectAddress.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("User Client Error", err.Error())
		return
	}

	for _, role := range state.Roles {
		err := client.RevokeRoleFromUser(&zilliz.UserRevokeRoleFromParams{
			UserName: state.Username.ValueString(),
			RoleName: role.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError("Revoke Role Error", err.Error())
			return
		}
	}
}

func (r *UserRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state UserRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionID := NormalizeConnectionID(plan.ConnectAddress.ValueString())

	existingRoles := make(map[string]bool)
	for _, r := range state.Roles {
		existingRoles[r.ValueString()] = true
	}

	plannedRoles := make(map[string]bool)
	for _, r := range plan.Roles {
		plannedRoles[r.ValueString()] = true
	}

	client, err := r.client.User(plan.ConnectAddress.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get user client", err.Error())
		return
	}

	// Revoke roles not in plan
	for role := range existingRoles {
		if !plannedRoles[role] {
			client.RevokeRoleFromUser(&zilliz.UserRevokeRoleFromParams{
				UserName: plan.Username.ValueString(),
				RoleName: role,
			})
		}
	}

	// Grant new roles
	for role := range plannedRoles {
		if !existingRoles[role] {
			client.GrantRoleToUser(&zilliz.UserGrantRoleToUserParams{
				UserName: plan.Username.ValueString(),
				RoleName: role,
			})
		}
	}

	// Set ID to /connection/{connection_id}/user/{username}/roles
	if len(plan.Roles) > 0 {
		plan.Id = types.StringValue(BuildUserRoleID(connectionID, plan.Username.ValueString()))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse import ID, format: "/connection/{connection_id}/user/{username}/roles"
	connectionID, username, ok := ParseUserRoleID(req.ID)
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			"Import ID must be in the format '/connection/{connection_id}/user/{username}/roles'",
		)
		return
	}

	connectAddress := "https://" + connectionID

	client, err := r.client.User(connectAddress)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get user client (import)", fmt.Sprintf("ConnectAddress: %s, error: %s", connectAddress, err))
		return
	}

	roles, err := client.DescribeUser(username)
	if err != nil {
		resp.Diagnostics.AddError("Import UserRole Error", fmt.Sprintf("Failed to list roles: %s", err.Error()))
		return
	}

	tfsRoles := make([]types.String, 0, len(roles))
	for _, r := range roles {
		tfsRoles = append(tfsRoles, types.StringValue(r))
	}

	state := UserRoleResourceModel{
		Id:             types.StringValue(req.ID),
		ConnectAddress: types.StringValue(connectAddress),
		Username:       types.StringValue(username),
		Roles:          tfsRoles,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
