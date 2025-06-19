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

// NewAliasResource returns a new alias resource.
func NewAliasResource() resource.Resource {
	return &AliasResource{}
}

type AliasResource struct {
	client *zilliz.Client
}

type AliasResourceModel struct {
	Id             types.String `tfsdk:"id"` // /connections/{connect_address}/databases/{db_name}/aliases/{alias_name}
	ConnectAddress types.String `tfsdk:"connect_address"`
	DbName         types.String `tfsdk:"db_name"`
	AliasName      types.String `tfsdk:"alias_name"`
	CollectionName types.String `tfsdk:"collection_name"`
}

var _ resource.ResourceWithImportState = &AliasResource{}

func (r *AliasResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alias"
}

func (r *AliasResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a collection alias in a Zilliz Cloud database.
Changing any field will force resource replacement as alias updates require drop and recreate operations.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: `The unique identifier for the alias resource.

**Format:**

` + "`/connections/{connect_address}/databases/{db_name}/aliases/{alias_name}`" + `

**Example:**

` + "`/connections/in01-xxx/databases/testdb/aliases/myalias`" + `

> **Note:** This value is automatically set and should not be manually specified.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connect_address": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `The connection address of the target Zilliz Cloud cluster.
You can obtain this value from the output of the ` + "`zillizcloud_cluster`" + ` resource, for example:
` + "`zillizcloud_cluster.example.connect_address`" + `

**Example:**
` + "`https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534`" + `

> **Note:** The address must include the protocol (e.g., ` + "`https://`" + `).`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"db_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: `The name of the database containing the alias.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"alias_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: `The name of the alias.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"collection_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: `The name of the collection to which the alias points.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *AliasResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func BuildAliasID(connectAddress, dbName, aliasName string) string {
	return fmt.Sprintf("/connections/%s/databases/%s/aliases/%s", connectAddress, dbName, aliasName)
}

func ParseAliasID(id string) (connectAddress, dbName, aliasName string, ok bool) {
	parts := strings.Split(id, "/")
	if len(parts) != 7 || parts[1] != "connections" || parts[3] != "databases" || parts[5] != "aliases" {
		return "", "", "", false
	}
	return parts[2], parts[4], parts[6], true
}

func (r *AliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AliasResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...) // Get planned data
	if resp.Diagnostics.HasError() {
		return
	}

	connectAddress := data.ConnectAddress.ValueString()
	client, err := r.client.Collection(connectAddress, data.DbName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get collection client",
			fmt.Sprintf("ConnectAddress: %s, error: %s", connectAddress, err.Error()),
		)
		return
	}

	err = client.CreateAlias(&zilliz.CreateAliasParams{
		DbName:         data.DbName.ValueString(),
		AliasName:      data.AliasName.ValueString(),
		CollectionName: data.CollectionName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create alias",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, AliasName: %s, error: %s", connectAddress, data.DbName.ValueString(), data.AliasName.ValueString(), err.Error()),
		)
		return
	}

	connectAddress = NormalizeConnectionID(connectAddress)
	data.Id = types.StringValue(BuildAliasID(connectAddress, data.DbName.ValueString(), data.AliasName.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state
}

func (r *AliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...) // Get current state
	if resp.Diagnostics.HasError() {
		return
	}

	connectAddress := data.ConnectAddress.ValueString()
	client, err := r.client.Collection(connectAddress, data.DbName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get collection client",
			fmt.Sprintf("ConnectAddress: %s, error: %s", connectAddress, err.Error()),
		)
		return
	}

	// Check if alias exists
	_, err = client.DescribeAlias(&zilliz.DescribeAliasParams{
		DbName:    data.DbName.ValueString(),
		AliasName: data.AliasName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to describe alias",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, AliasName: %s, error: %s", connectAddress, data.DbName.ValueString(), data.AliasName.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state
}

func (r *AliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...) // Get current state
	if resp.Diagnostics.HasError() {
		return
	}

	connectAddress := data.ConnectAddress.ValueString()
	client, err := r.client.Collection(connectAddress, data.DbName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get collection client",
			fmt.Sprintf("ConnectAddress: %s, error: %s", connectAddress, err.Error()),
		)
		return
	}

	err = client.DropAlias(&zilliz.DropAliasParams{
		DbName:    data.DbName.ValueString(),
		AliasName: data.AliasName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to drop alias",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, AliasName: %s, error: %s", connectAddress, data.DbName.ValueString(), data.AliasName.ValueString(), err.Error()),
		)
		return
	}
}

func (r *AliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse import ID, format: "/connections/{connect_address}/databases/{db_name}/aliases/{alias_name}"
	connectAddress, dbName, aliasName, ok := ParseAliasID(req.ID)
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			"Import ID must be in the format '/connections/{connect_address}/databases/{db_name}/aliases/{alias_name}'",
		)
		return
	}

	connectAddressFull := "https://" + connectAddress
	client, err := r.client.Collection(connectAddressFull, dbName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get collection client (import)",
			fmt.Sprintf("ConnectAddress: %s, error: %s", connectAddressFull, err),
		)
		return
	}

	_, err = client.DescribeAlias(&zilliz.DescribeAliasParams{
		DbName:    dbName,
		AliasName: aliasName,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to import alias: alias does not exist or cannot be retrieved",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, AliasName: %s, error: %s", connectAddressFull, dbName, aliasName, err.Error()),
		)
		return
	}

	state := AliasResourceModel{
		Id:             types.StringValue(req.ID),
		ConnectAddress: types.StringValue(connectAddressFull),
		DbName:         types.StringValue(dbName),
		AliasName:      types.StringValue(aliasName),
		CollectionName: types.StringValue(""), // Not available from import
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...) // Save state
}

// Update method: implements drop + create pattern for alias updates.
func (r *AliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state AliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // Old state
	if resp.Diagnostics.HasError() {
		return
	}
	var plan AliasResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...) // New plan
	if resp.Diagnostics.HasError() {
		return
	}

	connectAddress := state.ConnectAddress.ValueString()
	client, err := r.client.Collection(connectAddress, state.DbName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get collection client",
			fmt.Sprintf("ConnectAddress: %s, error: %s", connectAddress, err.Error()),
		)
		return
	}

	// Step 1: Drop the existing alias
	err = client.DropAlias(&zilliz.DropAliasParams{
		DbName:    state.DbName.ValueString(),
		AliasName: state.AliasName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to drop existing alias during update",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, AliasName: %s, error: %s", connectAddress, state.DbName.ValueString(), state.AliasName.ValueString(), err.Error()),
		)
		return
	}

	// Step 2: Create the new alias with updated values
	newConnectAddress := plan.ConnectAddress.ValueString()
	newClient, err := r.client.Collection(newConnectAddress, plan.DbName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get collection client for new alias",
			fmt.Sprintf("ConnectAddress: %s, error: %s", newConnectAddress, err.Error()),
		)
		return
	}

	err = newClient.CreateAlias(&zilliz.CreateAliasParams{
		DbName:         plan.DbName.ValueString(),
		AliasName:      plan.AliasName.ValueString(),
		CollectionName: plan.CollectionName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create new alias during update",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, AliasName: %s, error: %s", newConnectAddress, plan.DbName.ValueString(), plan.AliasName.ValueString(), err.Error()),
		)
		return
	}

	// Update the ID with normalized connection address
	normalizedConnectAddress := NormalizeConnectionID(newConnectAddress)
	plan.Id = types.StringValue(BuildAliasID(normalizedConnectAddress, plan.DbName.ValueString(), plan.AliasName.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...) // Save state
}
