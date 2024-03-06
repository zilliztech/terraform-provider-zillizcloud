// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ProjectsDataSource{}

func NewProjectsDataSource() datasource.DataSource {
	return &ProjectsDataSource{}
}

// ProjectsDataSource defines the data source implementation.
type ProjectsDataSource struct {
	client *zilliz.Client
}

// ProjectDataSourceModel describes the data source data model.
type ProjectItem struct {
	ProjectId     types.String `tfsdk:"project_id"`
	ProjectName   types.String `tfsdk:"project_name"`
	InstanceCount types.Int64  `tfsdk:"instance_count"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
}


// ProjectsDataSourceModel describes the data source data model.
type ProjectsDataSourceModel struct {
	Projects []ProjectItem `tfsdk:"projects"`
	Id       types.String  `tfsdk:"id"`
}

func (d *ProjectsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

func (d *ProjectsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cloud Providers data source",

		Attributes: map[string]schema.Attribute{
			"projects": schema.ListNestedAttribute{
				MarkdownDescription: "List of Projects",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"project_id": schema.StringAttribute{
							MarkdownDescription: "Project Identifier",
							Computed:            true,
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
					},
				},
			},
		},
	}
}

func (d *ProjectsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ProjectsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "sending list projects request...")
	projects, err := d.client.ListProjects()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to ListProjects, got error: %s", err))
		return
	}

	state.Id = types.StringValue(strconv.FormatInt(rand.Int63(), 10))

	for _, p := range projects {

		item := ProjectItem{
			ProjectId:     types.StringValue(p.ProjectId),
			ProjectName:   types.StringValue(p.ProjectName),
			InstanceCount: types.Int64Value(p.InstanceCount),
			CreatedAt:     types.Int64Value(p.CreateTimeMilli),
		}

		state.Projects = append(state.Projects, item)

	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
