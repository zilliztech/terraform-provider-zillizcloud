package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fwdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	fwdschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func TestProjectDataSourceLookupByID(t *testing.T) {
	ctx := context.Background()
	d, schema := testProjectDataSource(t, func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", req.Method)
		}
		if req.URL.Path != "/projects/proj-1" {
			t.Fatalf("path = %s, want /projects/proj-1", req.URL.Path)
		}
		writeProviderProjectResponse(t, w, zilliz.Project{
			ProjectId:     "proj-1",
			ProjectName:   "Project One",
			InstanceCount: 2,
			CreateTime:    "2026-05-07T08:00:00Z",
			Plan:          "Enterprise",
			RegionIds:     []string{"aws-us-east-1", "gcp-us-west1"},
			OrgType:       "SAAS",
		})
	})

	resp := fwdatasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	d.Read(ctx, fwdatasource.ReadRequest{Config: testProjectDataSourceConfig(t, ctx, schema, "proj-1", "")}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %s", resp.Diagnostics)
	}

	var got ProjectsDataSourceModel
	diags := resp.State.Get(ctx, &got)
	if diags.HasError() {
		t.Fatalf("state get diagnostics: %s", diags)
	}
	if got.Id.ValueString() != "proj-1" || got.ProjectName.ValueString() != "Project One" || got.Plan.ValueString() != "Enterprise" || got.CreateTime.ValueString() != "2026-05-07T08:00:00Z" {
		t.Fatalf("state = %#v", got)
	}
	assertProjectRegionIDs(t, ctx, got.RegionIds, []string{"aws-us-east-1", "gcp-us-west1"})
}

func TestProjectDataSourceLookupByDeprecatedName(t *testing.T) {
	ctx := context.Background()
	d, schema := testProjectDataSource(t, func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", req.Method)
		}
		switch req.URL.Path {
		case "/projects":
			writeProviderProjectResponse(t, w, []zilliz.Project{
				{ProjectId: "proj-1", ProjectName: "Project One", Plan: "Standard"},
				{ProjectId: "proj-2", ProjectName: "Project Two", Plan: "Enterprise"},
			})
		case "/projects/proj-2":
			writeProviderProjectResponse(t, w, zilliz.Project{
				ProjectId:   "proj-2",
				ProjectName: "Project Two",
				Plan:        "Enterprise",
				RegionIds:   []string{"aws-us-east-1"},
			})
		default:
			t.Fatalf("path = %s, want /projects or /projects/proj-2", req.URL.Path)
		}
	})

	resp := fwdatasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	d.Read(ctx, fwdatasource.ReadRequest{Config: testProjectDataSourceConfig(t, ctx, schema, "", "Project Two")}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %s", resp.Diagnostics)
	}

	var got ProjectsDataSourceModel
	diags := resp.State.Get(ctx, &got)
	if diags.HasError() {
		t.Fatalf("state get diagnostics: %s", diags)
	}
	if got.Id.ValueString() != "proj-2" || got.Name.ValueString() != "Project Two" || got.Plan.ValueString() != "Enterprise" {
		t.Fatalf("state = %#v", got)
	}
	assertProjectRegionIDs(t, ctx, got.RegionIds, []string{"aws-us-east-1"})
}

func TestProjectDataSourceNotFound(t *testing.T) {
	ctx := context.Background()
	d, schema := testProjectDataSource(t, func(w http.ResponseWriter, req *http.Request) {
		writeProviderProjectResponse(t, w, []zilliz.Project{{ProjectId: "proj-1", ProjectName: "Project One"}})
	})

	resp := fwdatasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	d.Read(ctx, fwdatasource.ReadRequest{Config: testProjectDataSourceConfig(t, ctx, schema, "", "Missing")}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected not found diagnostic")
	}
	if !strings.Contains(resp.Diagnostics[0].Detail(), "Project not found with name: Missing") {
		t.Fatalf("detail = %q", resp.Diagnostics[0].Detail())
	}
}

func TestProjectDataSourceDuplicateName(t *testing.T) {
	ctx := context.Background()
	d, schema := testProjectDataSource(t, func(w http.ResponseWriter, req *http.Request) {
		writeProviderProjectResponse(t, w, []zilliz.Project{
			{ProjectId: "proj-1", ProjectName: "Duplicate"},
			{ProjectId: "proj-2", ProjectName: "Duplicate"},
		})
	})

	resp := fwdatasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	d.Read(ctx, fwdatasource.ReadRequest{Config: testProjectDataSourceConfig(t, ctx, schema, "", "Duplicate")}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected duplicate name diagnostic")
	}
	if !strings.Contains(resp.Diagnostics[0].Detail(), "Multiple projects found with name: Duplicate") {
		t.Fatalf("detail = %q", resp.Diagnostics[0].Detail())
	}
}

func testProjectDataSource(t *testing.T, handler http.HandlerFunc) (*ProjectDataSource, fwdschema.Schema) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	c, err := zilliz.NewClient(
		zilliz.WithApiKey("test-api-key"),
		zilliz.WithBaseUrl(server.URL),
	)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	d := &ProjectDataSource{client: c}
	var schemaResp fwdatasource.SchemaResponse
	d.Schema(context.Background(), fwdatasource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %s", schemaResp.Diagnostics)
	}

	return d, schemaResp.Schema
}

func testProjectDataSourceConfig(t *testing.T, ctx context.Context, schema fwdschema.Schema, id, name string) tfsdk.Config {
	t.Helper()

	model := ProjectsDataSourceModel{
		Id:            types.StringNull(),
		Name:          types.StringNull(),
		ProjectName:   types.StringNull(),
		InstanceCount: types.Int64Null(),
		CreatedAt:     types.Int64Null(),
		CreateTime:    types.StringNull(),
		Plan:          types.StringNull(),
		RegionIds:     types.SetNull(types.StringType),
		OrgType:       types.StringNull(),
	}
	if id != "" {
		model.Id = types.StringValue(id)
	}
	if name != "" {
		model.Name = types.StringValue(name)
	}

	state := tfsdk.State{Schema: schema}
	diags := state.Set(ctx, &model)
	if diags.HasError() {
		t.Fatalf("config state set diagnostics: %s", diags)
	}

	return tfsdk.Config{
		Raw:    state.Raw,
		Schema: schema,
	}
}
