// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
type ProjectModel struct {
	ProjectId   types.String `tfsdk:"project_id"`
	ProjectName types.String `tfsdk:"project_name"`
}

func (p ProjectModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"project_id":   types.StringType,
		"project_name": types.StringType,
	}
}

// ProjectsDataSourceModel describes the data source data model.
type ProjectsDataSourceModel struct {
	Projects types.List   `tfsdk:"projects"`
	Id       types.String `tfsdk:"id"`
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
					},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Projects identifier",
				Computed:            true,
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
	var data ProjectsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projects, err := d.client.ListProjects()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to ListProjects, got error: %s", err))
		return
	}

	// Save data into Terraform state
	data.Id = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))

	var ps []ProjectModel
	for _, p := range projects {
		ps = append(ps, ProjectModel{ProjectId: types.StringValue(p.ProjectId), ProjectName: types.StringValue(p.ProjectName)})
	}
	var diag diag.Diagnostics
	data.Projects, diag = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: ProjectModel{}.AttrTypes()}, ps)
	resp.Diagnostics.Append(diag...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
