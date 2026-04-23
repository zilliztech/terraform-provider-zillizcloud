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

var _ resource.Resource = &EndpointWhitelistResource{}
var _ resource.ResourceWithConfigure = &EndpointWhitelistResource{}

func NewEndpointWhitelistResource() resource.Resource {
	return &EndpointWhitelistResource{}
}

type EndpointWhitelistResource struct {
	client *zilliz.Client
}

type EndpointWhitelistResourceModel struct {
	Id          types.String `tfsdk:"id"`
	ProjectId   types.String `tfsdk:"project_id"`
	RegionId    types.String `tfsdk:"region_id"`
	OuterUserId types.String `tfsdk:"outer_user_id"`
}

func (r *EndpointWhitelistResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoint_whitelist"
}

func (r *EndpointWhitelistResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Adds an external cloud account to the private link endpoint whitelist.

**Limitations:**
- The upstream API only supports adding whitelist entries; there is no list or delete API.
- **Read is a no-op** — drift cannot be detected.
- **Destroy is a no-op** — the whitelist entry is *not* removed from Zilliz Cloud; only from Terraform state.
- Because the resource ID is set to ` + "`project_id`" + `, at most one whitelist resource may be declared per project.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID. Also serves as the resource ID.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "Cloud region ID.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"outer_user_id": schema.StringAttribute{
				MarkdownDescription: "External cloud account identifier (Azure Subscription ID, Alibaba/Tencent/Huawei Account ID).",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *EndpointWhitelistResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EndpointWhitelistResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EndpointWhitelistResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.AddEndpointWhitelist(data.ProjectId.ValueString(), &zilliz.AddEndpointWhitelistRequest{
		RegionId:    data.RegionId.ValueString(),
		OuterUserId: data.OuterUserId.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to add endpoint whitelist",
			fmt.Sprintf("projectId=%s error=%s", data.ProjectId.ValueString(), err))
		return
	}

	data.Id = data.ProjectId
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read is a no-op — the upstream API does not expose a GET for whitelist entries.
func (r *EndpointWhitelistResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EndpointWhitelistResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is unreachable — all attributes are RequiresReplace.
func (r *EndpointWhitelistResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EndpointWhitelistResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is a no-op — the upstream API does not expose a DELETE for whitelist entries.
func (r *EndpointWhitelistResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
