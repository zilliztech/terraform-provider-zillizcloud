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
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider/utils"
)

var _ resource.Resource = &DatabaseResource{}
var _ resource.ResourceWithConfigure = &DatabaseResource{}
var _ resource.ResourceWithImportState = &DatabaseResource{}

func NewDatabaseResource() resource.Resource {
	return &DatabaseResource{}
}

type DatabaseResource struct {
	client *zilliz.Client
}

type DatabaseResourceModel struct {
	Id             types.String `tfsdk:"id"` // /connections/{connect_address}/databases/{db_name}
	ConnectAddress types.String `tfsdk:"connect_address"`
	DbName         types.String `tfsdk:"db_name"`
	Properties     types.Map    `tfsdk:"properties"`
}

func (r *DatabaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *DatabaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a database in a Zilliz Cloud cluster.
This resource allows you to create, read, and delete databases for a specific cluster using its connect address.
Typical use case: managing logical databases for multi-tenancy, isolation, and organization of data.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: `The unique identifier for the database resource, generated by the service.

**Format:**
` + "`" + `/connections/{connect_address}/databases/{db_name}` + "`" + `
**Fields:**
- ` + "`connect_address`" + ` — The address used to connect to the cluster.
- ` + "`db_name`" + ` — The name of the database.

**Example:**
` + "`" + `/connections/in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534/databases/mydb` + "`" + `

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
			},
			"db_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: `The name of the database to be managed. Must be unique within the cluster.`,
			},
			"properties": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				MarkdownDescription: `A map of database properties.

**Example:**

` + "`" + `{
  "database.replica.number": "3",
  "database.max.collections": "1000",
  "database.force.deny.writing": "true"
}` + "`" + `

> All values should be provided as strings and will be converted to appropriate types internally. Support properties can be found [here](https://docs.zilliz.com/reference/restful/create-database-v2)

**Reference:** https://github.com/milvus-io/milvus/blob/v2.5.12/pkg/common/common.go#L186
`,
			},
		},
	}
}

func (r *DatabaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...) // Get planned data
	if resp.Diagnostics.HasError() {
		return
	}

	connectAddress := data.ConnectAddress.ValueString()
	client, err := r.client.Cluster(connectAddress)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get cluster client",
			fmt.Sprintf("ConnectAddress: %s, error: %s", connectAddress, err.Error()),
		)
		return
	}

	// Parse properties only if they are provided
	var props map[string]any
	if !data.Properties.IsNull() {
		var err error
		props, err = utils.ConvertPropertiesToMap(data.Properties)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to convert properties",
				err.Error(),
			)
			return
		}
	}

	_, err = client.CreateDatabase(zilliz.CreateDatabaseParams{
		DbName:     data.DbName.ValueString(),
		Properties: props,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create database",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, error: %s", connectAddress, data.DbName.ValueString(), err.Error()),
		)
		return
	}

	// Always normalize connect_address for ID, but keep full address in state
	normalizedAddress := NormalizeConnectionID(connectAddress)
	data.Id = types.StringValue(BuildDatabaseID(normalizedAddress, data.DbName.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state
}

func (r *DatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // Get current state
	if resp.Diagnostics.HasError() {
		return
	}

	connectAddress := state.ConnectAddress.ValueString()
	client, err := r.client.Cluster(connectAddress)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get cluster client",
			fmt.Sprintf("ConnectAddress: %s, error: %s", connectAddress, err.Error()),
		)
		return
	}

	db, err := client.DescribeDatabase(zilliz.DescribeDatabaseParams{
		DbName: state.DbName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read database info",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, error: %s", connectAddress, state.DbName.ValueString(), err.Error()),
		)
		return
	}

	// Convert API properties to Map format
	apiProps := make(map[string]any)
	for _, kv := range db.Properties {
		if k, ok := kv["key"].(string); ok {
			if v, ok := kv["value"].(string); ok {
				apiProps[k] = v
			}
		}
	}

	propsMap, err := utils.ConvertMapToProperties(apiProps)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to convert properties",
			err.Error(),
		)
		return
	}
	state.Properties = propsMap

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...) // Keep state
}

func (r *DatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // Get current state
	if resp.Diagnostics.HasError() {
		return
	}

	connectAddress := state.ConnectAddress.ValueString()
	client, err := r.client.Cluster(connectAddress)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get cluster client",
			fmt.Sprintf("Could not get cluster client for connect_address '%s': %s", connectAddress, err.Error()),
		)
		return
	}

	_, err = client.DropDatabase(zilliz.DropDatabaseParams{
		DbName: state.DbName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete database",
			fmt.Sprintf("Could not delete database '%s' on connect_address '%s': %s", state.DbName.ValueString(), connectAddress, err.Error()),
		)
		return
	}
}

func (r *DatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse import ID, expected format: "/connections/{connect_address}/databases/{db_name}"
	connectAddress, dbName, ok := ParseDatabaseID(req.ID)
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			"Import ID must be in the format '/connections/{connect_address}/databases/{db_name}'. Please check your import command.",
		)
		return
	}

	connectAddress = "https://" + connectAddress

	// Set connect_address as is
	state := DatabaseResourceModel{
		Id:             types.StringValue(req.ID),
		ConnectAddress: types.StringValue(connectAddress),
		DbName:         types.StringValue(dbName),
		Properties:     types.MapNull(types.StringType),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...) // Save state
}

// BuildDatabaseID constructs the resource ID in the format: /connections/{connect_address}/databases/{db_name}.
func BuildDatabaseID(connectAddress, dbName string) string {
	if strings.HasPrefix(connectAddress, "https://") {
		connectAddress = NormalizeConnectionID(connectAddress)
	}
	return fmt.Sprintf("/connections/%s/databases/%s", connectAddress, dbName)
}

// ParseDatabaseID parses the resource ID and extracts connect_address and db_name.
// Expected format: /connections/{connect_address}/databases/{db_name}.
func ParseDatabaseID(id string) (string, string, bool) {
	parts := strings.Split(id, "/")
	if len(parts) != 5 || parts[1] != "connections" || parts[3] != "databases" {
		return "", "", false
	}
	return parts[2], parts[4], true
}

// Update updates the database properties if there are any changes.
func (r *DatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: update current has permission issue, directly return for now
	var plan, state DatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)   // New plan
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // Old state
	if resp.Diagnostics.HasError() {
		return
	}

	connectAddress := plan.ConnectAddress.ValueString()
	client, err := r.client.Cluster(connectAddress)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get cluster client",
			fmt.Sprintf("Could not get cluster client for connect_address '%s': %s", connectAddress, err.Error()),
		)
		return
	}

	dbName := plan.DbName.ValueString()
	db, err := client.DescribeDatabase(zilliz.DescribeDatabaseParams{DbName: dbName})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to describe database",
			fmt.Sprintf("Could not describe database '%s' on connect_address '%s': %s", dbName, connectAddress, err.Error()),
		)
		return
	}

	// Parse the desired properties from the plan
	planProps, err := utils.ConvertPropertiesToMap(plan.Properties)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to convert properties",
			err.Error(),
		)
		return
	}

	// Extract the current properties from the API response (array of key-value maps)
	dbProps := map[string]string{}
	for _, kv := range db.Properties {
		k, vok := kv["key"].(string)
		v, vok2 := kv["value"].(string)
		if vok && vok2 {
			dbProps[k] = v
		}
	}

	// Compare plan and current properties. If any difference, update is needed.
	needUpdate := false
	if len(planProps) != len(dbProps) {
		needUpdate = true
	} else {
		for k, v := range planProps {
			if dbProps[k] != fmt.Sprintf("%v", v) {
				needUpdate = true
				break
			}
		}
		for k, v := range dbProps {
			if fmt.Sprintf("%v", planProps[k]) != v {
				needUpdate = true
				break
			}
		}
	}

	if needUpdate {
		propsAny := make(map[string]any, len(planProps))
		for k, v := range planProps {
			propsAny[k] = v
		}
		_, err := client.UpdateDatabase(zilliz.UpdateDatabaseParams{
			DbName:     dbName,
			Properties: propsAny,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update database properties",
				fmt.Sprintf("Could not update properties for database '%s' on connect_address '%s': %s", dbName, connectAddress, err.Error()),
			)
			return
		}
	}

	// Refresh state after update
	plan.Id = types.StringValue(BuildDatabaseID(NormalizeConnectionID(connectAddress), dbName))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...) // Update state
}
