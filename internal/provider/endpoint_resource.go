package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ resource.Resource = &EndpointResource{}
var _ resource.ResourceWithConfigure = &EndpointResource{}

func NewEndpointResource() resource.Resource {
	return &EndpointResource{}
}

type EndpointResource struct {
	client *zilliz.Client
}

type EndpointResourceModel struct {
	Id                    types.String `tfsdk:"id"`
	ProjectId             types.String `tfsdk:"project_id"`
	RegionId              types.String `tfsdk:"region_id"`
	EndpointId            types.String `tfsdk:"endpoint_id"`
	GcpProjectId          types.String `tfsdk:"gcp_project_id"`
	CloudId               types.String `tfsdk:"cloud_id"`
	EndpointService       types.String `tfsdk:"endpoint_service"`
	EndpointServiceStatus types.String `tfsdk:"endpoint_service_status"`
	EndpointStatus        types.String `tfsdk:"endpoint_status"`
}

func (r *EndpointResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoint"
}

func (r *EndpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a private link endpoint for a Zilliz Cloud project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "Cloud region ID.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"endpoint_id": schema.StringAttribute{
				MarkdownDescription: "VPC endpoint ID.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"gcp_project_id": schema.StringAttribute{
				MarkdownDescription: "GCP project ID (required for GCP regions).",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"cloud_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoint_service": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoint_service_status": schema.StringAttribute{
				Computed: true,
			},
			"endpoint_status": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *EndpointResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider type",
			fmt.Sprintf("Expected *zilliz.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

// findEndpoint scans all pages of ListEndpoints looking for endpointId. Returns nil if not found.
func (r *EndpointResource) findEndpoint(projectId, endpointId string) (*zilliz.Endpoint, error) {
	const pageSize = 100
	page := 1
	for {
		eps, pg, err := r.client.ListEndpoints(projectId, page, pageSize)
		if err != nil {
			return nil, err
		}
		for i := range eps {
			if eps[i].EndpointId == endpointId {
				return &eps[i], nil
			}
		}
		if page*pageSize >= pg.Count || len(eps) == 0 {
			return nil, nil
		}
		page++
	}
}

func (r *EndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EndpointResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &zilliz.CreateEndpointRequest{
		RegionId:   data.RegionId.ValueString(),
		EndpointId: data.EndpointId.ValueString(),
	}
	if !data.GcpProjectId.IsNull() && !data.GcpProjectId.IsUnknown() {
		body.GcpProjectId = data.GcpProjectId.ValueString()
	}

	created, err := r.client.CreateEndpoint(data.ProjectId.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create endpoint",
			fmt.Sprintf("projectId=%s endpointId=%s error=%s",
				data.ProjectId.ValueString(), data.EndpointId.ValueString(), err))
		return
	}

	data.Id = types.StringValue(created.EndpointId)

	// Refresh computed status by finding the endpoint.
	ep, err := r.findEndpoint(data.ProjectId.ValueString(), created.EndpointId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read endpoint after create", err.Error())
		return
	}
	if ep != nil {
		data.CloudId = types.StringValue(ep.CloudId)
		data.EndpointService = types.StringValue(ep.EndpointService)
		data.EndpointServiceStatus = types.StringValue(ep.EndpointServiceStatus)
		data.EndpointStatus = types.StringValue(ep.EndpointStatus)
	} else {
		data.CloudId = types.StringNull()
		data.EndpointService = types.StringNull()
		data.EndpointServiceStatus = types.StringNull()
		data.EndpointStatus = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ep, err := r.findEndpoint(state.ProjectId.ValueString(), state.EndpointId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to list endpoints", err.Error())
		return
	}
	if ep == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.CloudId = types.StringValue(ep.CloudId)
	state.EndpointService = types.StringValue(ep.EndpointService)
	state.EndpointServiceStatus = types.StringValue(ep.EndpointServiceStatus)
	state.EndpointStatus = types.StringValue(ep.EndpointStatus)
	state.RegionId = types.StringValue(ep.RegionId)
	if ep.GcpProjectId != nil {
		state.GcpProjectId = types.StringValue(*ep.GcpProjectId)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *EndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All user-supplied attributes are RequiresReplace; Update is unreachable in practice,
	// but the framework requires the method. Pass plan to state unchanged.
	var plan EndpointResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var gcp *string
	if !state.GcpProjectId.IsNull() && state.GcpProjectId.ValueString() != "" {
		v := state.GcpProjectId.ValueString()
		gcp = &v
	}

	err := r.client.DeleteEndpoint(
		state.ProjectId.ValueString(),
		state.EndpointId.ValueString(),
		state.RegionId.ValueString(),
		gcp,
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete endpoint",
			fmt.Sprintf("projectId=%s endpointId=%s error=%s",
				state.ProjectId.ValueString(), state.EndpointId.ValueString(), err))
		return
	}
}
