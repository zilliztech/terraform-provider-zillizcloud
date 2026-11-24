package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func TestServiceAccountAccessUpdateRoutes(t *testing.T) {
	type request struct {
		method string
		path   string
		body   string
	}
	for _, tc := range []struct {
		name                   string
		previousRole, nextRole string
		previous, next         map[string][]string
		projectScope           string
		want                   []request
	}{
		{
			name:     "unchanged",
			previous: map[string][]string{"proj-a": {"r-old"}},
			next:     map[string][]string{"proj-a": {"r-old"}},
		},
		{
			name:     "project administrator changes existing roles",
			previous: map[string][]string{"proj-a": {"r-old"}, "proj-b": {"r-keep"}},
			next:     map[string][]string{"proj-a": {"r-new"}, "proj-b": {"r-keep"}},
			want: []request{
				{"PATCH", "/v2/serviceAccounts/sa-test/projectAssignments/proj-a", `{"projectRoleIds":["r-new"]}`},
			},
		},
		{
			name:     "clear roles while retaining assignment",
			previous: map[string][]string{"proj-a": {"r-old"}},
			next:     map[string][]string{"proj-a": {}},
			want: []request{
				{"PATCH", "/v2/serviceAccounts/sa-test/projectAssignments/proj-a", `{"projectRoleIds":[]}`},
			},
		},
		{
			name:     "multiple changed projects in stable order",
			previous: map[string][]string{"proj-b": {"r-old"}, "proj-a": {"r-old"}, "proj-c": {}},
			next:     map[string][]string{"proj-b": {}, "proj-a": {"r-new"}, "proj-c": {}},
			want: []request{
				{"PATCH", "/v2/serviceAccounts/sa-test/projectAssignments/proj-a", `{"projectRoleIds":["r-new"]}`},
				{"PATCH", "/v2/serviceAccounts/sa-test/projectAssignments/proj-b", `{"projectRoleIds":[]}`},
			},
		},
		{
			name:     "add project and change existing roles",
			previous: map[string][]string{"proj-a": {"r-old"}},
			next:     map[string][]string{"proj-a": {"r-new"}, "proj-b": {}},
			want: []request{
				{"PUT", "/v2/serviceAccounts/sa-test/access", `{"projectAssignments":[{"projectId":"proj-a","projectRoleIds":["r-new"]},{"projectId":"proj-b","projectRoleIds":[]}]}`},
			},
		},
		{
			name:     "remove project",
			previous: map[string][]string{"proj-a": {}, "proj-b": {}},
			next:     map[string][]string{"proj-b": {}},
			want: []request{
				{"PUT", "/v2/serviceAccounts/sa-test/access", `{"projectAssignments":[{"projectId":"proj-b","projectRoleIds":[]}]}`},
			},
		},
		{
			name:     "replace project with same assignment count",
			previous: map[string][]string{"proj-a": {}},
			next:     map[string][]string{"proj-b": {}},
			want: []request{
				{"PUT", "/v2/serviceAccounts/sa-test/access", `{"projectAssignments":[{"projectId":"proj-b","projectRoleIds":[]}]}`},
			},
		},
		{
			name:     "clear all project access",
			previous: map[string][]string{"proj-a": {"r-old"}},
			want: []request{
				{"PUT", "/v2/serviceAccounts/sa-test/access", `{"projectAssignments":[]}`},
			},
		},
		{
			name:     "projects to organization role",
			previous: map[string][]string{"proj-a": {"r-old"}},
			nextRole: "r-billing-admin",
			want: []request{
				{"PUT", "/v2/serviceAccounts/sa-test/access", `{"orgRoleId":"r-billing-admin"}`},
			},
		},
		{
			name:         "organization role to projects",
			previousRole: "r-billing-admin",
			next:         map[string][]string{"proj-a": {}},
			want: []request{
				{"PUT", "/v2/serviceAccounts/sa-test/access", `{"projectAssignments":[{"projectId":"proj-a","projectRoleIds":[]}]}`},
			},
		},
		{
			name:         "change organization role",
			previousRole: "r-billing-admin",
			nextRole:     "r-org-owner",
			want: []request{
				{"PUT", "/v2/serviceAccounts/sa-test/access", `{"orgRoleId":"r-org-owner"}`},
			},
		},
		{
			name:         "clear organization role",
			previousRole: "r-billing-admin",
			want: []request{
				{"PUT", "/v2/serviceAccounts/sa-test/access", `{"projectAssignments":[]}`},
			},
		},
		{
			name:         "project-local roles",
			projectScope: "proj-a",
			previous:     map[string][]string{"proj-a": {"r-old"}},
			next:         map[string][]string{"proj-a": {}},
			want: []request{
				{"PATCH", "/v2/serviceAccounts/sa-test/projectAssignments/proj-a", `{"projectRoleIds":[]}`},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			account := &zilliz.ServiceAccountResponse{ServiceAccountId: "sa-test", Name: "test", Status: "active"}
			if tc.previousRole != "" {
				account.OrgRoles = []zilliz.ServiceAccountRole{{RoleId: tc.previousRole}}
			}
			if tc.projectScope != "" {
				account.ProjectScope = &tc.projectScope
			}
			for project, roles := range tc.previous {
				assignment := zilliz.ServiceAccountProjectAssignment{ProjectId: project}
				for _, role := range roles {
					assignment.ProjectRoles = append(assignment.ProjectRoles, zilliz.ServiceAccountRole{RoleId: role})
				}
				account.ProjectAssignments = append(account.ProjectAssignments, assignment)
			}
			api := &serviceAccountTestAPI{accounts: map[string]*zilliz.ServiceAccountResponse{"sa-test": account}}
			var mu sync.Mutex
			var requests []request
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				mu.Lock()
				requests = append(requests, request{r.Method, r.URL.Path, string(body)})
				mu.Unlock()
				// A project administrator can PATCH roles but cannot replace all access.
				if len(tc.want) > 0 && tc.want[0].method == "PATCH" && r.Method == "PUT" {
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"code":403,"message":"Org Owner required"}`))
					return
				}
				r.Body = io.NopCloser(bytes.NewReader(body))
				api.serveHTTP(w, r)
			}))
			defer server.Close()
			client, err := zilliz.NewClient(zilliz.WithApiKey("test"), zilliz.WithBaseUrl(server.URL+"/v2"))
			require.NoError(t, err)
			var state ServiceAccountResourceModel
			require.False(t, state.populate(ctx, account).HasError())
			plan := state
			plan.OrgRoleId = types.StringNull()
			if tc.nextRole != "" {
				plan.OrgRoleId = types.StringValue(tc.nextRole)
			}
			assignments := map[string]attr.Value{}
			for project, roles := range tc.next {
				value, diagnostics := types.SetValueFrom(ctx, types.StringType, roles)
				require.False(t, diagnostics.HasError())
				if tc.projectScope != "" {
					plan.ProjectRoleIds = value
				} else {
					assignments[project] = value
				}
			}
			plan.ProjectAssignments = types.MapValueMust(serviceAccountRoleSetType, assignments)
			var diagnostics diag.Diagnostics
			r := &ServiceAccountResource{client: client}
			require.NoError(t, r.updateAccount(ctx, &plan, &state, &diagnostics))
			require.False(t, diagnostics.HasError())
			mu.Lock()
			defer mu.Unlock()
			require.Len(t, requests, len(tc.want))
			for i, want := range tc.want {
				require.Equal(t, want.method, requests[i].method)
				require.Equal(t, want.path, requests[i].path)
				require.JSONEq(t, want.body, requests[i].body)
			}
			api.mu.Lock()
			defer api.mu.Unlock()
			var actual ServiceAccountResourceModel
			require.False(t, actual.populate(ctx, account).HasError())
			require.True(t, plan.OrgRoleId.Equal(actual.OrgRoleId))
			require.True(t, plan.ProjectAssignments.Equal(actual.ProjectAssignments))
			require.True(t, plan.ProjectRoleIds.Equal(actual.ProjectRoleIds))
		})
	}
}

func TestServiceAccountProjectRoleUpdateRecoversPartialFailure(t *testing.T) {
	ctx := context.Background()
	account := &zilliz.ServiceAccountResponse{
		ServiceAccountId: "sa-test",
		Name:             "test",
		Status:           "active",
		ProjectAssignments: []zilliz.ServiceAccountProjectAssignment{
			{ProjectId: "proj-a", ProjectRoles: []zilliz.ServiceAccountRole{{RoleId: "r-old"}}},
			{ProjectId: "proj-b", ProjectRoles: []zilliz.ServiceAccountRole{{RoleId: "r-old"}}},
		},
	}
	api := &serviceAccountTestAPI{accounts: map[string]*zilliz.ServiceAccountResponse{"sa-test": account}}
	var mu sync.Mutex
	var patches []string
	denySecond := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			mu.Lock()
			patches = append(patches, r.URL.Path)
			deny := denySecond && r.URL.Path == "/v2/serviceAccounts/sa-test/projectAssignments/proj-b"
			mu.Unlock()
			if deny {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]any{"code": 403, "message": "project grant denied"})
				return
			}
		}
		api.serveHTTP(w, r)
	}))
	defer server.Close()
	client, err := zilliz.NewClient(zilliz.WithApiKey("project-admin"), zilliz.WithBaseUrl(server.URL+"/v2"))
	require.NoError(t, err)
	r := &ServiceAccountResource{client: client}
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	var state ServiceAccountResourceModel
	require.False(t, state.populate(ctx, account).HasError())
	plan := state
	plan.ProjectAssignments = types.MapValueMust(serviceAccountRoleSetType, map[string]attr.Value{
		"proj-a": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("r-new")}),
		"proj-b": types.SetValueMust(types.StringType, []attr.Value{}),
	})
	prior := tfsdk.State{Schema: schema.Schema}
	planned := tfsdk.Plan{Schema: schema.Schema}
	require.False(t, prior.Set(ctx, &state).HasError())
	require.False(t, planned.Set(ctx, &plan).HasError())
	first := resource.UpdateResponse{State: prior}
	r.Update(ctx, resource.UpdateRequest{State: prior, Plan: planned}, &first)
	require.True(t, first.Diagnostics.HasError())
	var partial ServiceAccountResourceModel
	require.False(t, first.State.Get(ctx, &partial).HasError())
	require.True(t, partial.ProjectAssignments.Elements()["proj-a"].Equal(plan.ProjectAssignments.Elements()["proj-a"]))
	require.True(t, partial.ProjectAssignments.Elements()["proj-b"].Equal(state.ProjectAssignments.Elements()["proj-b"]))
	mu.Lock()
	denySecond = false
	mu.Unlock()
	second := resource.UpdateResponse{State: first.State}
	r.Update(ctx, resource.UpdateRequest{State: first.State, Plan: planned}, &second)
	require.False(t, second.Diagnostics.HasError(), second.Diagnostics)
	var actual ServiceAccountResourceModel
	require.False(t, second.State.Get(ctx, &actual).HasError())
	require.True(t, actual.ProjectAssignments.Equal(plan.ProjectAssignments))
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []string{
		"/v2/serviceAccounts/sa-test/projectAssignments/proj-a",
		"/v2/serviceAccounts/sa-test/projectAssignments/proj-b",
		"/v2/serviceAccounts/sa-test/projectAssignments/proj-b",
	}, patches)
}
