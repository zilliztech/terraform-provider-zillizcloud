package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

// These tests run Terraform against an in-memory HTTP API, never a cloud account.
type serviceAccountTestAPI struct {
	mu       sync.Mutex
	accounts map[string]*zilliz.ServiceAccountResponse
	keys     map[string]*zilliz.ServiceAccountKeyResponse
	next     int
}

func newServiceAccountTestAPI(t *testing.T) (*serviceAccountTestAPI, *httptest.Server) {
	t.Helper()
	api := &serviceAccountTestAPI{
		accounts: map[string]*zilliz.ServiceAccountResponse{},
		keys:     map[string]*zilliz.ServiceAccountKeyResponse{},
	}
	server := httptest.NewServer(http.HandlerFunc(api.serveHTTP))
	t.Cleanup(server.Close)
	return api, server
}

func (api *serviceAccountTestAPI) serveHTTP(w http.ResponseWriter, r *http.Request) {
	api.mu.Lock()
	defer api.mu.Unlock()
	reply := func(v any) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": v,
		})
	}
	fail := func(code int) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":    code,
			"message": "mock rejected request",
		})
	}
	var body map[string]json.RawMessage
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			fail(400)
			return
		}
	}
	text := func(field string) string {
		var value string
		_ = json.Unmarshal(body[field], &value)
		return value
	}
	roles := func(raw json.RawMessage) []zilliz.ServiceAccountRole {
		var ids []string
		_ = json.Unmarshal(raw, &ids)
		result := []zilliz.ServiceAccountRole{}
		for _, id := range ids {
			result = append(result, zilliz.ServiceAccountRole{
				RoleId:   id,
				RoleName: id,
			})
		}
		return result
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v2/"), "/")
	if r.Method == "POST" && ((len(parts) == 1 && parts[0] == "serviceAccounts") || (len(parts) == 3 && parts[0] == "project")) {
		api.next++
		id := fmt.Sprintf("sa-%d", api.next)
		account := &zilliz.ServiceAccountResponse{
			ServiceAccountId:   id,
			Name:               text("name"),
			Description:        text("description"),
			Status:             "active",
			CreatedBy:          "key-actor",
			CreatorName:        "test",
			CreateTime:         "2026-11-23T00:00:00Z",
			OrgRoles:           []zilliz.ServiceAccountRole{},
			ProjectAssignments: []zilliz.ServiceAccountProjectAssignment{},
		}
		if parts[0] == "project" {
			account.ProjectScope = &parts[1]
			account.ProjectAssignments = append(account.ProjectAssignments, zilliz.ServiceAccountProjectAssignment{
				ProjectId:    parts[1],
				ProjectRoles: roles(body["projectRoleIds"]),
			})
		} else {
			_, org := body["orgRoleId"]
			_, projects := body["projectAssignments"]
			if org == projects {
				fail(400)
				return
			}
			if org {
				if text("orgRoleId") != "r-billing-admin" && text("orgRoleId") != "r-org-owner" {
					fail(400)
					return
				}
				account.OrgRoles = []zilliz.ServiceAccountRole{{RoleId: text("orgRoleId")}}
			}
			var assignments []struct {
				ProjectId      string          `json:"projectId"`
				ProjectRoleIds json.RawMessage `json:"projectRoleIds"`
			}
			_ = json.Unmarshal(body["projectAssignments"], &assignments)
			for _, a := range assignments {
				account.ProjectAssignments = append(account.ProjectAssignments, zilliz.ServiceAccountProjectAssignment{
					ProjectId:    a.ProjectId,
					ProjectRoles: roles(a.ProjectRoleIds),
				})
			}
		}
		api.accounts[id] = account
		reply(account)
		return
	}
	if len(parts) == 1 && r.Method == "GET" {
		result := []zilliz.ServiceAccountResponse{}
		for _, a := range api.accounts {
			result = append(result, *a)
		}
		reply(zilliz.ServiceAccountListResponse{
			Count:           len(result),
			CurrentPage:     1,
			PageSize:        100,
			ServiceAccounts: result,
		})
		return
	}
	if len(parts) < 2 || parts[0] != "serviceAccounts" {
		fail(400)
		return
	}
	account, exists := api.accounts[parts[1]]
	if !exists {
		fail(21155)
		return
	}
	if len(parts) == 2 {
		switch r.Method {
		case "GET":
			reply(account)
		case "PATCH":
			if account.Status == "disable" && (body["name"] != nil || body["description"] != nil) {
				fail(400)
				return
			}
			if body["name"] != nil {
				account.Name = text("name")
			}
			if body["description"] != nil {
				account.Description = text("description")
			}
			if body["status"] != nil {
				account.Status = text("status")
			}
			reply(account)
		case "DELETE":
			for _, key := range api.keys {
				if key.ServiceAccountId == account.ServiceAccountId {
					fail(400)
					return
				}
			}
			delete(api.accounts, account.ServiceAccountId)
			reply(true)
		default:
			fail(400)
		}
		return
	}
	if parts[2] == "access" && r.Method == "PUT" {
		if account.Status != "active" || account.ProjectScope != nil {
			fail(400)
			return
		}
		_, org := body["orgRoleId"]
		_, projects := body["projectAssignments"]
		if org == projects || len(body) != 1 {
			fail(400)
			return
		}
		if org && text("orgRoleId") != "r-billing-admin" && text("orgRoleId") != "r-org-owner" {
			fail(400)
			return
		}
		account.OrgRoles = []zilliz.ServiceAccountRole{}
		account.ProjectAssignments = []zilliz.ServiceAccountProjectAssignment{}
		if org {
			account.OrgRoles = []zilliz.ServiceAccountRole{{RoleId: text("orgRoleId")}}
		}
		var assignments []struct {
			ProjectId      string          `json:"projectId"`
			ProjectRoleIds json.RawMessage `json:"projectRoleIds"`
		}
		_ = json.Unmarshal(body["projectAssignments"], &assignments)
		for _, a := range assignments {
			account.ProjectAssignments = append(account.ProjectAssignments, zilliz.ServiceAccountProjectAssignment{
				ProjectId: a.ProjectId, ProjectRoles: roles(a.ProjectRoleIds),
			})
		}
		reply(zilliz.ServiceAccountRolesResponse{
			ServiceAccountId:   account.ServiceAccountId,
			OrgRoles:           account.OrgRoles,
			ProjectAssignments: account.ProjectAssignments,
		})
		return
	}
	if parts[2] == "projectAssignments" && len(parts) == 4 {
		if account.Status != "active" || len(account.OrgRoles) > 0 || (account.ProjectScope != nil && (r.Method != "PATCH" || *account.ProjectScope != parts[3])) {
			fail(400)
			return
		}
		index := -1
		for i, a := range account.ProjectAssignments {
			if a.ProjectId == parts[3] {
				index = i
			}
		}
		if r.Method == "PATCH" {
			if index < 0 || len(body) != 1 || body["projectRoleIds"] == nil {
				fail(400)
				return
			}
			account.ProjectAssignments[index].ProjectRoles = roles(body["projectRoleIds"])
			reply(zilliz.ServiceAccountAssignmentResponse{
				ServiceAccountId: account.ServiceAccountId,
				ProjectId:        parts[3],
				ProjectRoles:     account.ProjectAssignments[index].ProjectRoles,
			})
			return
		}
		if r.Method == "DELETE" {
			if index >= 0 {
				account.ProjectAssignments = append(account.ProjectAssignments[:index], account.ProjectAssignments[index+1:]...)
			}
			reply(true)
			return
		}
		if r.Method == "POST" {
			if index < 0 {
				account.ProjectAssignments = append(account.ProjectAssignments, zilliz.ServiceAccountProjectAssignment{
					ProjectId:    parts[3],
					ProjectRoles: roles(body["projectRoleIds"]),
				})
				index = len(account.ProjectAssignments) - 1
			}
			reply(zilliz.ServiceAccountAssignmentResponse{
				ServiceAccountId: account.ServiceAccountId,
				ProjectId:        parts[3],
				ProjectRoles:     account.ProjectAssignments[index].ProjectRoles,
			})
			return
		}
	}
	if parts[2] == "apiKeys" {
		if len(parts) == 3 {
			if r.Method == "GET" {
				keys := []zilliz.ServiceAccountKeyResponse{}
				for _, key := range api.keys {
					if key.ServiceAccountId == account.ServiceAccountId {
						keys = append(keys, *key)
					}
				}
				reply(zilliz.ServiceAccountKeyListResponse{
					Count:       len(keys),
					CurrentPage: 1,
					PageSize:    100,
					ApiKeys:     keys,
				})
				return
			}
			if r.Method == "POST" {
				if account.Status != "active" {
					fail(400)
					return
				}
				var days int
				_ = json.Unmarshal(body["expireDays"], &days)
				if days < 1 || days > 730 {
					fail(400)
					return
				}
				api.next++
				id := fmt.Sprintf("key-%d", api.next)
				created, _ := time.Parse(time.RFC3339, account.CreateTime)
				expires := created.Add(time.Duration(days) * 24 * time.Hour).Format(time.RFC3339)
				key := &zilliz.ServiceAccountKeyResponse{
					ApiKeyId:         id,
					ServiceAccountId: account.ServiceAccountId,
					Name:             text("name"),
					Description:      text("description"),
					ExpireTime:       &expires,
					CreateTime:       account.CreateTime,
					CreatedBy:        "key-actor",
					CreatorName:      "test",
				}
				api.keys[id] = key
				reply(zilliz.CreatedServiceAccountKeyResponse{
					ApiKey: *key,
					Secret: "test-secret-" + id,
				})
				return
			}
		}
		if len(parts) == 4 {
			key, ok := api.keys[parts[3]]
			if !ok || key.ServiceAccountId != account.ServiceAccountId {
				fail(21159)
				return
			}
			switch r.Method {
			case "GET":
				reply(key)
			case "PATCH":
				if account.Status != "active" {
					fail(400)
					return
				}
				if body["name"] != nil {
					key.Name = text("name")
				}
				if body["description"] != nil {
					key.Description = text("description")
				}
				reply(key)
			case "DELETE":
				delete(api.keys, key.ApiKeyId)
				reply(true)
			default:
				fail(400)
			}
			return
		}
	}
	fail(400)
}

func serviceAccountTestConfig(server *httptest.Server, permissions, name, description, status string, days int) string {
	return fmt.Sprintf(`
provider "zillizcloud" {
 api_key = "test-only"
 host_address = %q
}
resource "zillizcloud_service_account" "test" {
 name = %q
 description = %q
 status = %q
 %s
}
resource "zillizcloud_service_account_api_key" "test" {
 service_account_id = zillizcloud_service_account.test.id
 name = "test-key"
 description = %q
 expire_days = %d
}
data "zillizcloud_service_account" "test" { id = zillizcloud_service_account.test.id }
data "zillizcloud_service_accounts" "test" { name = zillizcloud_service_account.test.name }
data "zillizcloud_service_account_api_key" "test" {
 service_account_id = zillizcloud_service_account.test.id
 id = zillizcloud_service_account_api_key.test.id
}
data "zillizcloud_service_account_api_keys" "test" {
 service_account_id = zillizcloud_service_account.test.id
 depends_on = [zillizcloud_service_account_api_key.test]
}
`, server.URL+"/v2", name, description, status, permissions, description, days)
}

func TestServiceAccountTerraformLifecycle(t *testing.T) {
	api, server := newServiceAccountTestAPI(t)
	config := func(permissions, name, description string, days int) string {
		return serviceAccountTestConfig(server, permissions, name, description, "active", days)
	}
	initial := config(`project_assignments = { "proj-a" = ["r-writer"], "proj-b" = [] }`, "initial", "description", 90)
	updated := config(`project_assignments = { "proj-a" = [], "proj-c" = ["r-reader"] }`, "updated", "", 90)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: initial,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_service_account.test", "project_assignments.%", "2"),
					resource.TestCheckResourceAttr("zillizcloud_service_account_api_key.test", "secret", "test-secret-key-2"),
					resource.TestCheckResourceAttr("data.zillizcloud_service_accounts.test", "total_count", "1"),
					resource.TestCheckResourceAttr("data.zillizcloud_service_account_api_keys.test", "total_count", "1"),
				),
			},
			{
				Config: updated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_service_account.test", "project_assignments.proj-a.#", "0"),
					resource.TestCheckResourceAttr("zillizcloud_service_account.test", "description", ""),
					resource.TestCheckResourceAttr("zillizcloud_service_account_api_key.test", "secret", "test-secret-key-2"),
				),
			},
			{
				ResourceName:      "zillizcloud_service_account.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:            "zillizcloud_service_account_api_key.test",
				ImportState:             true,
				ImportStateId:           "sa-1/key-2",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
			{
				Config: config(`org_role_id = "r-billing-admin"`, "updated", "", 90),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_service_account.test", "org_role_id", "r-billing-admin"),
					resource.TestCheckResourceAttr("zillizcloud_service_account.test", "project_assignments.%", "0"),
				),
			},
			{
				Config: config(`org_role_id = "r-org-owner"`, "updated", "", 90),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_service_account.test", "org_role_id", "r-org-owner"),
					resource.TestCheckResourceAttr("zillizcloud_service_account.test", "project_assignments.%", "0"),
					resource.TestCheckResourceAttr("data.zillizcloud_service_account.test", "org_role_id", "r-org-owner"),
				),
			},
			{
				Config: config(`project_assignments = { "proj-d" = ["r-reader"], "proj-e" = [] }`, "updated", "", 90),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("zillizcloud_service_account.test", "org_role_id"),
					resource.TestCheckResourceAttr("zillizcloud_service_account.test", "project_assignments.%", "2"),
					resource.TestCheckResourceAttr("zillizcloud_service_account.test", "project_assignments.proj-e.#", "0"),
				),
			},
			{
				Config: config(`project_assignments = {}`, "updated", "", 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_service_account.test", "project_assignments.%", "0"),
					resource.TestCheckResourceAttr("zillizcloud_service_account_api_key.test", "expire_days", "30"),
					resource.TestCheckResourceAttr("zillizcloud_service_account_api_key.test", "secret", "test-secret-key-3"),
				),
			},
		},
	})
	api.mu.Lock()
	defer api.mu.Unlock()
	require.Empty(t, api.accounts)
	require.Empty(t, api.keys)
}

func TestServiceAccountTerraformProjectLocal(t *testing.T) {
	_, server := newServiceAccountTestAPI(t)
	config := func(roles string) string {
		return serviceAccountTestConfig(server, "project_scope = \"proj-local\"\nproject_role_ids = "+roles, "local", "", "active", 730)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config(`["r-writer"]`),
				Check:  resource.TestCheckResourceAttr("zillizcloud_service_account.test", "project_scope", "proj-local"),
			},
			{
				Config: config(`[]`),
				Check:  resource.TestCheckResourceAttr("zillizcloud_service_account.test", "project_role_ids.#", "0"),
			},
			{
				ResourceName:      "zillizcloud_service_account.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config(`["r-writer"]`),
			},
			{
				Config: serviceAccountTestConfig(server, `project_scope = "proj-local"`, "local", "", "active", 730),
				Check:  resource.TestCheckResourceAttr("zillizcloud_service_account.test", "project_role_ids.#", "0"),
			},
		},
	})
}

func TestServiceAccountTerraformStatus(t *testing.T) {
	_, server := newServiceAccountTestAPI(t)
	config := func(name, status string) string {
		return fmt.Sprintf(`provider "zillizcloud" {
 api_key = "test"
 host_address = %q
}
resource "zillizcloud_service_account" "test" {
 name = %q
 status = %q
 project_assignments = {}
}
`, server.URL+"/v2", name, status)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config("test", "disable"),
				Check:  resource.TestCheckResourceAttr("zillizcloud_service_account.test", "status", "disable"),
			},
			{
				Config: config("renamed", "active"),
				Check:  resource.TestCheckResourceAttr("zillizcloud_service_account.test", "name", "renamed"),
			},
			{
				Config: config("renamed", "disable"),
				Check:  resource.TestCheckResourceAttr("zillizcloud_service_account.test", "status", "disable"),
			},
		},
	})
}

func TestServiceAccountTerraformInvalidConfig(t *testing.T) {
	_, server := newServiceAccountTestAPI(t)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      serviceAccountTestConfig(server, `org_role_id = ""`, "test", "", "active", 90),
				ExpectError: regexp.MustCompile("Invalid role ID"),
			},
			{
				Config:      serviceAccountTestConfig(server, `org_role_id = ["r-billing-admin"]`, "test", "", "active", 90),
				ExpectError: regexp.MustCompile("string required"),
			},
			{
				Config: serviceAccountTestConfig(server, `project_scope = "proj-local"
org_role_id = "r-billing-admin"`, "test", "", "active", 90),
				ExpectError: regexp.MustCompile("Invalid project-local permissions"),
			},
			{
				Config:      serviceAccountTestConfig(server, "", "test", "", "active", 90),
				ExpectError: regexp.MustCompile("Missing or conflicting organization permissions"),
			},
			{
				Config: serviceAccountTestConfig(server, `org_role_id = "r-billing-admin"
project_assignments = {}`, "test", "", "active", 90),
				ExpectError: regexp.MustCompile("conflicting organization permissions"),
			},
			{
				Config: serviceAccountTestConfig(server, `project_scope = "proj-a"
project_assignments = {}`, "test", "", "active", 90),
				ExpectError: regexp.MustCompile("Invalid project-local permissions"),
			},
			{
				Config:      serviceAccountTestConfig(server, `project_assignments = {}`, "test", "", "active", 731),
				ExpectError: regexp.MustCompile("between 1 and 730"),
			},
		},
	})
}

func TestServiceAccountReadDoesNotForgetOnPermissionError(t *testing.T) {
	ctx := context.Background()
	for _, code := range []int{
		21155,
		403,
		404,
	} {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(403)
				fmt.Fprintf(w, `{"code":%d,"message":"test"}`, code)
			}))
			defer server.Close()
			c, err := zilliz.NewClient(zilliz.WithApiKey("test"), zilliz.WithBaseUrl(server.URL))
			require.NoError(t, err)
			r := &ServiceAccountResource{client: c}
			var schema frameworkresource.SchemaResponse
			r.Schema(ctx, frameworkresource.SchemaRequest{}, &schema)
			state := tfsdk.State{Schema: schema.Schema}
			data := ServiceAccountResourceModel{
				Id:                 types.StringValue("sa-test"),
				OrgRoleId:          types.StringNull(),
				ProjectRoleIds:     types.SetNull(types.StringType),
				ProjectAssignments: types.MapNull(serviceAccountRoleSetType),
			}
			require.False(t, state.Set(ctx, &data).HasError())
			resp := frameworkresource.ReadResponse{State: state}
			r.Read(ctx, frameworkresource.ReadRequest{State: state}, &resp)
			require.Equal(t, code == 21155, resp.State.Raw.IsNull())
			require.Equal(t, code != 21155, resp.Diagnostics.HasError())
		})
	}
}

func TestServiceAccountDisabledUpdateDoesNotEnableImplicitly(t *testing.T) {
	r := &ServiceAccountResource{}
	state := ServiceAccountResourceModel{
		Name:   types.StringValue("old"),
		Status: types.StringValue("disable"),
	}
	plan := state
	plan.Name = types.StringValue("new")
	var diagnostics diag.Diagnostics
	require.ErrorContains(t, r.updateAccount(context.Background(), &plan, &state, &diagnostics), "enable this account")
}

func TestServiceAccountKeyMetadataPreservesSecret(t *testing.T) {
	expires := "2027-02-21T00:00:00Z"
	data := ServiceAccountKeyResourceModel{
		Secret:     types.StringValue("saved-secret"),
		ExpireDays: types.Int64Null(),
	}
	data.populate(&zilliz.ServiceAccountKeyResponse{
		CreateTime: "2026-11-23T00:00:00Z",
		ExpireTime: &expires,
	})
	require.Equal(t, "saved-secret", data.Secret.ValueString())
	require.EqualValues(t, 90, data.ExpireDays.ValueInt64())
}

func TestServiceAccountNullRoleValidation(t *testing.T) {
	var diagnostics diag.Diagnostics
	validateServiceAccountRoles(types.SetValueMust(types.StringType, []attr.Value{types.StringNull()}), &diagnostics)
	require.True(t, diagnostics.HasError())
}

func TestServiceAccountAccessReplacementReconcilesRoles(t *testing.T) {
	ctx := context.Background()
	api, server := newServiceAccountTestAPI(t)
	api.accounts["sa-test"] = &zilliz.ServiceAccountResponse{
		ServiceAccountId: "sa-test",
		Status:           "active",
		ProjectAssignments: []zilliz.ServiceAccountProjectAssignment{{
			ProjectId:    "proj-a",
			ProjectRoles: []zilliz.ServiceAccountRole{{RoleId: "r-old"}},
		}},
	}
	c, err := zilliz.NewClient(zilliz.WithApiKey("test"), zilliz.WithBaseUrl(server.URL+"/v2"))
	require.NoError(t, err)
	state := ServiceAccountResourceModel{
		Id:                 types.StringValue("sa-test"),
		Status:             types.StringValue("active"),
		OrgRoleId:          types.StringNull(),
		ProjectRoleIds:     types.SetNull(types.StringType),
		ProjectAssignments: types.MapValueMust(serviceAccountRoleSetType, map[string]attr.Value{}),
	}
	plan := state
	plan.ProjectAssignments = types.MapValueMust(serviceAccountRoleSetType, map[string]attr.Value{
		"proj-a": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("r-new")}),
	})
	var diagnostics diag.Diagnostics
	r := &ServiceAccountResource{client: c}
	require.NoError(t, r.updateAccount(ctx, &plan, &state, &diagnostics))
	require.False(t, diagnostics.HasError())
	api.mu.Lock()
	defer api.mu.Unlock()
	require.Equal(t, "r-new", api.accounts["sa-test"].ProjectAssignments[0].ProjectRoles[0].RoleId)
}

func TestServiceAccountCreateFailureKeepsIdentity(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			fmt.Fprint(w, `{"code":0,"data":{"serviceAccountId":"sa-created","name":"test","status":"active","orgRoles":[],"projectAssignments":[]}}`)
			return
		}
		w.WriteHeader(403)
		fmt.Fprint(w, `{"code":403,"message":"cannot disable"}`)
	}))
	defer server.Close()
	c, err := zilliz.NewClient(zilliz.WithApiKey("test"), zilliz.WithBaseUrl(server.URL))
	require.NoError(t, err)
	r := &ServiceAccountResource{client: c}
	var schema frameworkresource.SchemaResponse
	r.Schema(ctx, frameworkresource.SchemaRequest{}, &schema)
	plan := tfsdk.State{Schema: schema.Schema}
	data := ServiceAccountResourceModel{
		Name:               types.StringValue("test"),
		Description:        types.StringValue(""),
		Status:             types.StringValue("disable"),
		OrgRoleId:          types.StringNull(),
		ProjectRoleIds:     types.SetNull(types.StringType),
		ProjectAssignments: types.MapNull(serviceAccountRoleSetType),
	}
	require.False(t, plan.Set(ctx, &data).HasError())
	resp := frameworkresource.CreateResponse{State: tfsdk.State{Schema: schema.Schema}}
	r.Create(ctx, frameworkresource.CreateRequest{
		Config: tfsdk.Config{
			Schema: schema.Schema,
			Raw:    plan.Raw,
		},
		Plan: tfsdk.Plan{
			Schema: schema.Schema,
			Raw:    plan.Raw,
		},
	}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	var result ServiceAccountResourceModel
	require.False(t, resp.State.Get(ctx, &result).HasError())
	require.Equal(t, "sa-created", result.Id.ValueString())
	require.Equal(t, "active", result.Status.ValueString())
}
