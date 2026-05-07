package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func TestProjectResourceSchemaIncludesRegionIDsAndPlanDefault(t *testing.T) {
	ctx := context.Background()
	r := NewProjectResource()

	var resp fwresource.SchemaResponse
	r.Schema(ctx, fwresource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %s", resp.Diagnostics)
	}

	regionAttr, ok := resp.Schema.Attributes["region_ids"].(fwschema.SetAttribute)
	if !ok {
		t.Fatalf("region_ids attribute = %T, want schema.SetAttribute", resp.Schema.Attributes["region_ids"])
	}
	if !regionAttr.Optional {
		t.Fatal("region_ids should be optional")
	}
	if regionAttr.ElementType != types.StringType {
		t.Fatalf("region_ids element type = %s, want string", regionAttr.ElementType)
	}

	planAttr, ok := resp.Schema.Attributes["plan"].(fwschema.StringAttribute)
	if !ok {
		t.Fatalf("plan attribute = %T, want schema.StringAttribute", resp.Schema.Attributes["plan"])
	}
	if !planAttr.Optional || !planAttr.Computed {
		t.Fatalf("plan Optional=%t Computed=%t, want both true", planAttr.Optional, planAttr.Computed)
	}
	if len(planAttr.Validators) == 0 {
		t.Fatal("plan should validate allowed API plan values")
	}
}

func TestProjectResourceDiffStringSets(t *testing.T) {
	ctx := context.Background()
	previous, diags := types.SetValueFrom(ctx, types.StringType, []string{"aws-us-east-1", "gcp-us-west1"})
	if diags.HasError() {
		t.Fatalf("previous set diagnostics: %s", diags)
	}
	next, diags := types.SetValueFrom(ctx, types.StringType, []string{"aws-us-east-1", "azure-westus"})
	if diags.HasError() {
		t.Fatalf("next set diagnostics: %s", diags)
	}

	added, removed, diags := diffStringSets(ctx, previous, next)
	if diags.HasError() {
		t.Fatalf("diff diagnostics: %s", diags)
	}
	if len(added) != 1 || added[0] != "azure-westus" {
		t.Fatalf("added = %#v, want [azure-westus]", added)
	}
	if len(removed) != 1 || removed[0] != "gcp-us-west1" {
		t.Fatalf("removed = %#v, want [gcp-us-west1]", removed)
	}
}

func TestProjectResourceCreateSendsRegionsAndDefaultPlan(t *testing.T) {
	ctx := context.Background()
	r, schema := testProjectResource(t, func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", req.Method)
		}
		if req.URL.Path != "/projects" {
			t.Fatalf("path = %s, want /projects", req.URL.Path)
		}
		var got zilliz.CreateProjectRequest
		decodeProviderProjectRequest(t, req, &got)
		want := zilliz.CreateProjectRequest{
			ProjectName: "test-project",
			Plan:        "Enterprise",
			Regions:     []string{"aws-us-east-1"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("body = %#v, want %#v", got, want)
		}
		writeProviderProjectResponse(t, w, "proj-created")
	})

	regionIds, diags := types.SetValueFrom(ctx, types.StringType, []string{"aws-us-east-1"})
	if diags.HasError() {
		t.Fatalf("region_ids diagnostics: %s", diags)
	}
	model := ProjectResourceModel{
		ProjectName: types.StringValue("test-project"),
		Plan:        types.StringValue("Enterprise"),
		RegionIds:   regionIds,
	}
	plan := tfsdk.Plan{Schema: schema}
	diags = plan.Set(ctx, &model)
	if diags.HasError() {
		t.Fatalf("plan set diagnostics: %s", diags)
	}

	resp := fwresource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(ctx, fwresource.CreateRequest{Plan: plan}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("create diagnostics: %s", resp.Diagnostics)
	}

	var got ProjectResourceModel
	diags = resp.State.Get(ctx, &got)
	if diags.HasError() {
		t.Fatalf("state get diagnostics: %s", diags)
	}
	if got.Id.ValueString() != "proj-created" {
		t.Fatalf("id = %q, want proj-created", got.Id.ValueString())
	}
}

func TestProjectResourceUpdatePlanAndAppendRegions(t *testing.T) {
	ctx := context.Background()
	var paths []string
	r, schema := testProjectResource(t, func(w http.ResponseWriter, req *http.Request) {
		paths = append(paths, req.Method+" "+req.URL.Path)
		switch req.URL.Path {
		case "/projects/proj-1/plan":
			if req.Method != http.MethodPatch {
				t.Fatalf("method = %s, want PATCH", req.Method)
			}
			var got zilliz.UpgradeProjectPlanRequest
			decodeProviderProjectRequest(t, req, &got)
			if got.Plan != "Enterprise" {
				t.Fatalf("plan = %s, want Enterprise", got.Plan)
			}
			writeProviderProjectResponse(t, w, "proj-1")
		case "/projects/proj-1/regions":
			if req.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", req.Method)
			}
			var got zilliz.AddProjectRegionsRequest
			decodeProviderProjectRequest(t, req, &got)
			want := zilliz.AddProjectRegionsRequest{Regions: []string{"gcp-us-west1"}}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("body = %#v, want %#v", got, want)
			}
			writeProviderProjectResponse(t, w, []string{"aws-us-east-1", "gcp-us-west1"})
		default:
			t.Fatalf("unexpected path: %s", req.URL.Path)
		}
	})

	state := testProjectState(t, ctx, schema, ProjectResourceModel{
		Id:          types.StringValue("proj-1"),
		ProjectName: types.StringValue("test-project"),
		Plan:        types.StringValue("Standard"),
		RegionIds:   testStringSet(t, ctx, "aws-us-east-1"),
	})
	plan := testProjectPlan(t, ctx, schema, ProjectResourceModel{
		Id:          types.StringValue("proj-1"),
		ProjectName: types.StringValue("test-project"),
		Plan:        types.StringValue("Enterprise"),
		RegionIds:   testStringSet(t, ctx, "aws-us-east-1", "gcp-us-west1"),
	})

	resp := fwresource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(ctx, fwresource.UpdateRequest{State: state, Plan: plan}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("update diagnostics: %s", resp.Diagnostics)
	}
	wantPaths := []string{"PATCH /projects/proj-1/plan", "POST /projects/proj-1/regions"}
	if !reflect.DeepEqual(paths, wantPaths) {
		t.Fatalf("paths = %#v, want %#v", paths, wantPaths)
	}
}

func TestProjectResourceUpdateDoesNotAppendUnchangedRegions(t *testing.T) {
	ctx := context.Background()
	var calls int
	r, schema := testProjectResource(t, func(w http.ResponseWriter, req *http.Request) {
		calls++
		t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
	})

	model := ProjectResourceModel{
		Id:          types.StringValue("proj-1"),
		ProjectName: types.StringValue("test-project"),
		Plan:        types.StringValue("Enterprise"),
		RegionIds:   testStringSet(t, ctx, "aws-us-east-1"),
	}
	state := testProjectState(t, ctx, schema, model)
	plan := testProjectPlan(t, ctx, schema, model)

	resp := fwresource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(ctx, fwresource.UpdateRequest{State: state, Plan: plan}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("update diagnostics: %s", resp.Diagnostics)
	}
	if calls != 0 {
		t.Fatalf("calls = %d, want 0", calls)
	}
}

func TestProjectResourceUpdateRejectsRegionRemoval(t *testing.T) {
	ctx := context.Background()
	r, schema := testProjectResource(t, func(w http.ResponseWriter, req *http.Request) {
		t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
	})

	state := testProjectState(t, ctx, schema, ProjectResourceModel{
		Id:          types.StringValue("proj-1"),
		ProjectName: types.StringValue("test-project"),
		Plan:        types.StringValue("Enterprise"),
		RegionIds:   testStringSet(t, ctx, "aws-us-east-1", "gcp-us-west1"),
	})
	plan := testProjectPlan(t, ctx, schema, ProjectResourceModel{
		Id:          types.StringValue("proj-1"),
		ProjectName: types.StringValue("test-project"),
		Plan:        types.StringValue("Enterprise"),
		RegionIds:   testStringSet(t, ctx, "aws-us-east-1"),
	})

	resp := fwresource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(ctx, fwresource.UpdateRequest{State: state, Plan: plan}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected region removal diagnostic")
	}
	if !strings.Contains(resp.Diagnostics[0].Summary(), "Project region removal is not supported") {
		t.Fatalf("summary = %q", resp.Diagnostics[0].Summary())
	}
}

func TestProjectResourceDeleteReturnsErrorAndPreservesState(t *testing.T) {
	ctx := context.Background()
	r := NewProjectResource()

	var schemaResp fwresource.SchemaResponse
	r.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %s", schemaResp.Diagnostics)
	}

	state := tfsdk.State{Schema: schemaResp.Schema}
	regionIds, diags := types.SetValueFrom(ctx, types.StringType, []string{"aws-us-east-1"})
	if diags.HasError() {
		t.Fatalf("region_ids diagnostics: %s", diags)
	}
	model := ProjectResourceModel{
		Id:          types.StringValue("proj-1"),
		ProjectName: types.StringValue("test-project"),
		Plan:        types.StringValue("Enterprise"),
		RegionIds:   regionIds,
	}
	diags = state.Set(ctx, &model)
	if diags.HasError() {
		t.Fatalf("state set diagnostics: %s", diags)
	}

	req := fwresource.DeleteRequest{State: state}
	resp := fwresource.DeleteResponse{State: state}
	r.Delete(ctx, req, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("Delete should return an error diagnostic")
	}
	if !strings.Contains(resp.Diagnostics[0].Summary(), "Project delete is not implemented") {
		t.Fatalf("delete summary = %q", resp.Diagnostics[0].Summary())
	}

	var got ProjectResourceModel
	diags = resp.State.Get(ctx, &got)
	if diags.HasError() {
		t.Fatalf("state get diagnostics: %s", diags)
	}
	if got.Id.ValueString() != "proj-1" {
		t.Fatalf("state id = %q, want proj-1", got.Id.ValueString())
	}
}

func testProjectResource(t *testing.T, handler http.HandlerFunc) (*ProjectResource, fwschema.Schema) {
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

	r := &ProjectResource{client: c}
	var schemaResp fwresource.SchemaResponse
	r.Schema(context.Background(), fwresource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %s", schemaResp.Diagnostics)
	}

	return r, schemaResp.Schema
}

func testStringSet(t *testing.T, ctx context.Context, values ...string) types.Set {
	t.Helper()
	set, diags := types.SetValueFrom(ctx, types.StringType, values)
	if diags.HasError() {
		t.Fatalf("set diagnostics: %s", diags)
	}
	return set
}

func testProjectState(t *testing.T, ctx context.Context, schema fwschema.Schema, model ProjectResourceModel) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: schema}
	diags := state.Set(ctx, &model)
	if diags.HasError() {
		t.Fatalf("state set diagnostics: %s", diags)
	}
	return state
}

func testProjectPlan(t *testing.T, ctx context.Context, schema fwschema.Schema, model ProjectResourceModel) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: schema}
	diags := plan.Set(ctx, &model)
	if diags.HasError() {
		t.Fatalf("plan set diagnostics: %s", diags)
	}
	return plan
}

func decodeProviderProjectRequest(t *testing.T, req *http.Request, target any) {
	t.Helper()
	if err := json.NewDecoder(req.Body).Decode(target); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
}

func writeProviderProjectResponse(t *testing.T, w http.ResponseWriter, data any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"code": 0,
		"data": data,
	}); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}
