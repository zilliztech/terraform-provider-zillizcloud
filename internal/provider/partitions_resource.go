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

// NewPartitionsResource returns a new partitions resource.
func NewPartitionsResource() resource.Resource {
	return &PartitionsResource{}
}

type PartitionsResource struct {
	client *zilliz.Client
}

type PartitionsResourceModel struct {
	Id             types.String `tfsdk:"id"` // /connections/{connect_address}/databases/{db_name}/collections/{collection_name}/partitions/{partition_name}
	ConnectAddress types.String `tfsdk:"connect_address"`
	DbName         types.String `tfsdk:"db_name"`
	CollectionName types.String `tfsdk:"collection_name"`
	PartitionName  types.String `tfsdk:"partition_name"`
}

var _ resource.ResourceWithImportState = &PartitionsResource{}

func (r *PartitionsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partitions"
}

func (r *PartitionsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a collection partition in a Zilliz Cloud database.
Changing db_name, connect_address, or collection_name will force resource replacement. Only partition_name can be updated in-place.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: `The unique identifier for the partition resource.

**Format:**
` + "`" + `/connections/{connect_address}/databases/{db_name}/collections/{collection_name}/partitions/{partition_name}` + "`" + `

**Example:**
` + "`" + `/connections/in01-xxx/databases/testdb/collections/testcollection/partitions/mypartition` + "`" + `

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
				MarkdownDescription: `The name of the database containing the partition.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"collection_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: `The name of the collection containing the partition.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"partition_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: `The name of the partition. (Can be updated in-place)`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *PartitionsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func BuildPartitionsID(connectAddress, dbName, collectionName, partitionName string) string {
	return fmt.Sprintf("/connections/%s/databases/%s/collections/%s/partitions/%s", connectAddress, dbName, collectionName, partitionName)
}

func ParsePartitionsID(id string) (connectAddress, dbName, collectionName, partitionName string, ok bool) {
	parts := strings.Split(id, "/")
	if len(parts) != 9 || parts[1] != "connections" || parts[3] != "databases" || parts[5] != "collections" || parts[7] != "partitions" {
		return "", "", "", "", false
	}
	return parts[2], parts[4], parts[6], parts[8], true
}

func (r *PartitionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PartitionsResourceModel
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

	err = client.CreatePartitions(&zilliz.CreatePartitionsParams{
		DbName:         data.DbName.ValueString(),
		PartitionsName: data.PartitionName.ValueString(),
		CollectionName: data.CollectionName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create partition",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, CollectionName: %s, PartitionName: %s, error: %s", connectAddress, data.DbName.ValueString(), data.CollectionName.ValueString(), data.PartitionName.ValueString(), err.Error()),
		)
		return
	}

	connectAddress = NormalizeConnectionID(connectAddress)
	data.Id = types.StringValue(BuildPartitionsID(connectAddress, data.DbName.ValueString(), data.CollectionName.ValueString(), data.PartitionName.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state
}

func (r *PartitionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PartitionsResourceModel
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

	// Check if partition exists by listing partitions
	partitions, err := client.ListPartitionses(&zilliz.ListPartitionsesParams{
		DbName:         data.DbName.ValueString(),
		CollectionName: data.CollectionName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list partitions",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, CollectionName: %s, error: %s", connectAddress, data.DbName.ValueString(), data.CollectionName.ValueString(), err.Error()),
		)
		return
	}

	// Check if our partition exists in the list
	partitionExists := false
	for _, partition := range partitions {
		if partition == data.PartitionName.ValueString() {
			partitionExists = true
			break
		}
	}

	if !partitionExists {
		resp.Diagnostics.AddError(
			"Partition not found",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, CollectionName: %s, PartitionName: %s not found in partition list", connectAddress, data.DbName.ValueString(), data.CollectionName.ValueString(), data.PartitionName.ValueString()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state
}

func (r *PartitionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PartitionsResourceModel
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

	err = client.DropPartitions(&zilliz.DropPartitionsParams{
		DbName:         data.DbName.ValueString(),
		PartitionsName: data.PartitionName.ValueString(),
		CollectionName: data.CollectionName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to drop partition",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, CollectionName: %s, PartitionName: %s, error: %s", connectAddress, data.DbName.ValueString(), data.CollectionName.ValueString(), data.PartitionName.ValueString(), err.Error()),
		)
		return
	}
}

func (r *PartitionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse import ID, format: "/connections/{connect_address}/databases/{db_name}/collections/{collection_name}/partitions/{partition_name}"
	connectAddress, dbName, collectionName, partitionName, ok := ParsePartitionsID(req.ID)
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			"Import ID must be in the format '/connections/{connect_address}/databases/{db_name}/collections/{collection_name}/partitions/{partition_name}'",
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

	// Check if partition exists by listing partitions
	partitions, err := client.ListPartitionses(&zilliz.ListPartitionsesParams{
		DbName:         dbName,
		CollectionName: collectionName,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list partitions during import",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, CollectionName: %s, error: %s", connectAddressFull, dbName, collectionName, err.Error()),
		)
		return
	}

	// Check if our partition exists in the list
	partitionExists := false
	for _, partition := range partitions {
		if partition == partitionName {
			partitionExists = true
			break
		}
	}

	if !partitionExists {
		resp.Diagnostics.AddError(
			"Failed to import partition: partition does not exist or cannot be retrieved",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, CollectionName: %s, PartitionName: %s not found in partition list", connectAddressFull, dbName, collectionName, partitionName),
		)
		return
	}

	state := PartitionsResourceModel{
		Id:             types.StringValue(req.ID),
		ConnectAddress: types.StringValue(connectAddressFull),
		DbName:         types.StringValue(dbName),
		CollectionName: types.StringValue(collectionName),
		PartitionName:  types.StringValue(partitionName),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...) // Save state
}

// Update method: will drop the partition and create a new one.
func (r *PartitionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PartitionsResourceModel
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

	err = client.DropPartitions(&zilliz.DropPartitionsParams{
		DbName:         data.DbName.ValueString(),
		PartitionsName: data.PartitionName.ValueString(),
		CollectionName: data.CollectionName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to drop partition",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, CollectionName: %s, PartitionName: %s, error: %s", connectAddress, data.DbName.ValueString(), data.CollectionName.ValueString(), data.PartitionName.ValueString(), err.Error()),
		)
		return
	}

	err = client.CreatePartitions(&zilliz.CreatePartitionsParams{
		DbName:         data.DbName.ValueString(),
		PartitionsName: data.PartitionName.ValueString(),
		CollectionName: data.CollectionName.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create partition",
			fmt.Sprintf("ConnectAddress: %s, DbName: %s, CollectionName: %s, PartitionName: %s, error: %s", connectAddress, data.DbName.ValueString(), data.CollectionName.ValueString(), data.PartitionName.ValueString(), err.Error()),
		)
		return
	}

	data.Id = types.StringValue(BuildPartitionsID(NormalizeConnectionID(connectAddress), data.DbName.ValueString(), data.CollectionName.ValueString(), data.PartitionName.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state
}
