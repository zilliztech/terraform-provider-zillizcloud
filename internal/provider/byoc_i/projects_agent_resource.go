package byoc_op

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
	util "github.com/zilliztech/terraform-provider-zillizcloud/client/retry"
)

const (
	defaultBYOCOpProjectAgentCreateTimeout time.Duration = 120 * time.Minute
)

var _ resource.Resource = &BYOCOpProjectAgentResource{}
var _ resource.ResourceWithConfigure = &BYOCOpProjectAgentResource{}

func NewBYOCOpProjectAgentResource() resource.Resource {
	return &BYOCOpProjectAgentResource{}
}

type BYOCOpProjectAgentResource struct {
	client *zilliz.Client
}

type BYOCOpProjectAgentResourceModel struct {
	ID             types.String   `tfsdk:"id"`
	ProjectID      types.String   `tfsdk:"project_id"`
	DataPlaneID    types.String   `tfsdk:"data_plane_id"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
	Status         types.String   `tfsdk:"status"`
	WaitUntilReady types.Bool     `tfsdk:"wait_until_ready"`
}

func (r *BYOCOpProjectAgentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_byoc_i_project_agent"
}

func (r *BYOCOpProjectAgentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "BYOC Op Project Agent resource for managing project agents.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Agent identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data_plane_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the data plane",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the BYOC agent",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"wait_until_ready": schema.BoolAttribute{
				MarkdownDescription: "Wait until the BYOC agent is ready",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *BYOCOpProjectAgentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *zilliz.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *BYOCOpProjectAgentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BYOCOpProjectAgentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating BYOC Op Project Agent...")

	// Set initial state
	data.ID = types.StringValue(data.ProjectID.ValueString())
	data.Status = types.StringValue(BYOCProjectStatusInit.String())

	// Save initial state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Wait for agent to be ready
	timeout, diags := data.Timeouts.Create(ctx, defaultBYOCOpProjectAgentCreateTimeout)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// query the agent status
	query := func() (*zilliz.DescribeByocAgentResponse, error) {
		request := &zilliz.DescribeByocAgentRequest{
			ProjectId:   data.ProjectID.ValueString(),
			DataPlaneID: data.DataPlaneID.ValueString(),
		}
		response, err := r.client.DescribeByocAgent(request)
		if err != nil {
			return nil, fmt.Errorf("failed to check BYOC Op project agent status: %w", err)
		}
		return response, nil
	}

	willingToWait := data.WaitUntilReady.ValueBool()

	// if not willing to wait, return the current status
	if !willingToWait {
		response, err := query()
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read BYOC Op project agent, got error: %s", err))
			return
		}
		data.Status = types.StringValue(BYOCProjectStatus(response.Status).String())
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// when willing to wait
	_, err := util.Poll[any](ctx, timeout, func() (*any, *util.Err) {
		response, err := query()
		if err != nil {
			return nil, &util.Err{Halt: true, Err: fmt.Errorf("failed to check BYOC Op project agent status: %w", err)}
		}

		tflog.Info(ctx, fmt.Sprintf("Describe BYOC Op project agent response: %+v", response))

		// wait until the agent is connected
		if response.Status != int(BYOCProjectStatusConnected) {
			return nil, &util.Err{Halt: false, Err: fmt.Errorf("agent is in status: %s", BYOCProjectStatus(response.Status))}
		}

		data.Status = types.StringValue(BYOCProjectStatus(response.Status).String())
		// update status
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return nil, nil
	})

	if err != nil {
		resp.Diagnostics.AddError("Creation Error", fmt.Sprintf("Failed to create BYOC Op project agent: %s", err))
		return
	}

	tflog.Info(ctx, "Created BYOC Op Project Agent")
}

func (r *BYOCOpProjectAgentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BYOCOpProjectAgentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading BYOC Op Project Agent...")

	response, err := r.client.DescribeByocAgent(&zilliz.DescribeByocAgentRequest{
		ProjectId:   state.ProjectID.ValueString(),
		DataPlaneID: state.DataPlaneID.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read BYOC Op project agent, got error: %s", err))
		return
	}

	state.Status = types.StringValue(BYOCProjectStatus(response.Status).String())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BYOCOpProjectAgentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BYOCOpProjectAgentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *BYOCOpProjectAgentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BYOCOpProjectAgentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting BYOC Op Project Agent...")

}
