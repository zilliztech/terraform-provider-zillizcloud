// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ProjectDataSource{}

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

// ProjectDataSource defines the data source implementation.
type ProjectDataSource struct {
	client *zilliz.Client
}

type ProjectsDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ProjectName   types.String `tfsdk:"project_name"`
	InstanceCount types.Int64  `tfsdk:"instance_count"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
	CreateTime    types.String `tfsdk:"create_time"`
	Plan          types.String `tfsdk:"plan"`
	OrgType       types.String `tfsdk:"org_type"`
}

func (d *ProjectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Project Identifier",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Project Name",
				DeprecationMessage:  "This attribute is deprecated and will be removed in a future version. Please use 'id' instead.",
				Optional:            true,
			},
			"project_name": schema.StringAttribute{
				MarkdownDescription: "Project Name",
				Computed:            true,
			},
			"instance_count": schema.Int64Attribute{
				MarkdownDescription: "Instance Count",
				Computed:            true,
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "Created At",
				Computed:            true,
			},
			"create_time": schema.StringAttribute{
				MarkdownDescription: "Project creation time returned by the project API.",
				Computed:            true,
			},
			"plan": schema.StringAttribute{
				MarkdownDescription: "Project plan.",
				Computed:            true,
			},
			"org_type": schema.StringAttribute{
				MarkdownDescription: "Organization type returned by the project API.",
				Computed:            true,
			},
		},
	}
}

func (d *ProjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*zilliz.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ProjectsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var p zilliz.Project
	if !state.Id.IsNull() && state.Id.ValueString() != "" {
		project, err := d.client.GetProjectById(state.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Project not found for id %q: %s", state.Id.ValueString(), err))
			return
		}
		p = *project
	} else {
		tflog.Trace(ctx, "sending list projects request...")
		projects, err := d.client.ListProjects()
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to ListProjects, got error: %s", err))
			return
		}

		name := state.Name.ValueString()
		if state.Name.IsNull() || name == "" {
			if projectID := os.Getenv("ZILLIZCLOUD_PROJECT_ID"); projectID != "" {
				for _, project := range projects {
					if project.ProjectId == projectID {
						p = project
						break
					}
				}
			}
		} else {
			for _, project := range projects {
				if project.ProjectName == name {
					if p.ProjectId != "" {
						resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Multiple projects found with name: %s", name))
						return
					}
					p = project
				}
			}
		}

		if p.ProjectId == "" {
			if name != "" {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Project not found with name: %s", name))
			} else {
				resp.Diagnostics.AddError("Client Error", "Project not found")
			}
			return
		}
	}

	state.Id = types.StringValue(p.ProjectId)
	state.Name = types.StringValue(p.ProjectName)
	state.ProjectName = types.StringValue(p.ProjectName)
	state.InstanceCount = types.Int64Value(p.InstanceCount)
	state.CreatedAt = types.Int64Value(p.CreateTimeMilli)
	state.CreateTime = types.StringValue(p.CreateTime)
	state.Plan = types.StringValue(p.Plan)
	state.OrgType = types.StringValue(p.OrgType)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
