package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithConfigure = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	client *zilliz.Client
}

type UserResourceModel struct {
	Id             types.String `tfsdk:"id"` // /connection/{connection_id}/user/{username}
	ConnectAddress types.String `tfsdk:"connect_address"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a user in a specific cluster using its connect_address",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Service generated identifier for the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connect_address": schema.StringAttribute{
				MarkdownDescription: "Service connect address.",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for the user.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for the user.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...) // Get planned data
	if resp.Diagnostics.HasError() {
		return
	}

	// Always set ConnectionID as connect_address without 'https://' prefix
	connectionID := NormalizeConnectionID(data.ConnectAddress.ValueString())

	client, err := r.client.User(data.ConnectAddress.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get user client",
			fmt.Sprintf("ConnectAddress: %s, error: %s", data.ConnectAddress.ValueString(), err.Error()),
		)
		return
	}

	err = client.CreateUser(&zilliz.CreateUserParams{
		Username: data.Username.ValueString(),
		Password: data.Password.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create user",
			fmt.Sprintf("ConnectAddress: %s, Username: %s, error: %s", data.ConnectAddress.ValueString(), data.Username.ValueString(), err.Error()),
		)
		return
	}

	data.Id = types.StringValue(BuildUserID(connectionID, data.Username.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // Get current state
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.User(state.ConnectAddress.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get user client",
			fmt.Sprintf("ConnectAddress: %s, error: %s", state.ConnectAddress.ValueString(), err.Error()),
		)
		return
	}

	_, err = client.DescribeUser(state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read user info",
			fmt.Sprintf("ConnectAddress: %s, Username: %s, error: %s", state.ConnectAddress.ValueString(), state.Username.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...) // Keep state
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // Get current state
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.User(state.ConnectAddress.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get user client",
			fmt.Sprintf("ConnectAddress: %s, error: %s", state.ConnectAddress.ValueString(), err.Error()),
		)
		return
	}

	err = client.DropUser(state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete user",
			fmt.Sprintf("ConnectAddress: %s, Username: %s, error: %s", state.ConnectAddress.ValueString(), state.Username.ValueString(), err.Error()),
		)
	}
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// 1. Get current state (old values)
	var state UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // Old state
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. Get plan (new values)
	var plan UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...) // New plan
	if resp.Diagnostics.HasError() {
		return
	}

	// Always set ConnectionID as connect_address without 'https://' prefix
	connectionID := NormalizeConnectionID(plan.ConnectAddress.ValueString())

	// 3. Get user client for old connect address
	client, err := r.client.User(state.ConnectAddress.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get user client",
			fmt.Sprintf("ConnectAddress: %s, error: %s", state.ConnectAddress.ValueString(), err.Error()),
		)
		return
	}

	// 4. Delete old user
	err = client.DropUser(state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete user (during update)",
			fmt.Sprintf("ConnectAddress: %s, Username: %s, error: %s", state.ConnectAddress.ValueString(), state.Username.ValueString(), err.Error()),
		)
		return
	}

	// 5. Get user client for new connect address
	client, err = r.client.User(plan.ConnectAddress.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get user client",
			fmt.Sprintf("ConnectAddress: %s, error: %s", plan.ConnectAddress.ValueString(), err.Error()),
		)
		return
	}
	err = client.CreateUser(&zilliz.CreateUserParams{
		Username: plan.Username.ValueString(),
		Password: plan.Password.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create user (during update)",
			fmt.Sprintf("ConnectAddress: %s, Username: %s, error: %s", plan.ConnectAddress.ValueString(), plan.Username.ValueString(), err.Error()),
		)
		return
	}

	// 6. Set new state
	plan.Id = types.StringValue(BuildUserID(connectionID, plan.Username.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...) // Update state
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse import ID, format: "/connection/{connection_id}/user/{username}"
	connectionID, username, ok := ParseUserID(req.ID)
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			"Import ID must be in the format '/connection/{connection_id}/user/{username}'",
		)
		return
	}

	// Set connect_address as 'https://' + connectionID
	connectAddress := "https://" + connectionID

	client, err := r.client.User(connectAddress)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get user client (import)",
			fmt.Sprintf("ConnectAddress: %s, error: %s", connectAddress, err),
		)
		return
	}

	// Check if user exists
	_, err = client.DescribeUser(username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to import user: user does not exist or cannot be retrieved",
			fmt.Sprintf("ConnectAddress: %s, Username: %s, error: %s", connectAddress, username, err.Error()),
		)
		return
	}

	// Set import state
	// Note: Password is set to empty during import for security reasons
	state := UserResourceModel{
		Id:             types.StringValue(req.ID),
		ConnectAddress: types.StringValue(connectAddress),
		Username:       types.StringValue(username),
		Password:       types.StringValue(""), // Password is empty on import
	}

	// Save state to response
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func BuildUserID(connectionID, username string) string {
	return fmt.Sprintf("/connection/%s/user/%s", connectionID, username)
}

func ParseUserID(id string) (string, string, bool) {
	parts := strings.Split(id, "/")
	if len(parts) != 4 || parts[0] != "connection" || parts[2] != "user" {
		return "", "", false
	}
	return parts[1], parts[3], true
}
