package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func TestServiceAccountTerraformImportPermissions(t *testing.T) {
	for _, status := range []string{
		"active",
		"disable",
	} {
		for _, tc := range []struct {
			name        string
			permissions string
			project     string
			orgRole     string
		}{
			{
				name:        "organization_role",
				permissions: `org_role_id = "r-billing-admin"`,
				orgRole:     "r-billing-admin",
			},
			{
				name:        "project_assignments",
				permissions: "project_assignments = {}",
			},
			{
				name:        "local_empty_roles",
				permissions: "project_scope = \"proj-local\"\nproject_role_ids = []",
				project:     "proj-local",
			},
			{
				name:        "local_omitted_roles",
				permissions: "project_scope = \"proj-local\"",
				project:     "proj-local",
			},
		} {
			t.Run(status+"/"+tc.name, func(t *testing.T) {
				account := &zilliz.ServiceAccountResponse{
					ServiceAccountId: "sa-imported",
					Name:             "imported",
					Status:           status,
				}
				orgRoleCheck := resource.TestCheckNoResourceAttr("zillizcloud_service_account.test", "org_role_id")
				if tc.orgRole != "" {
					account.OrgRoles = []zilliz.ServiceAccountRole{{RoleId: tc.orgRole}}
					orgRoleCheck = resource.TestCheckResourceAttr("zillizcloud_service_account.test", "org_role_id", tc.orgRole)
				}
				if tc.project != "" {
					account.ProjectScope = &tc.project
					account.ProjectAssignments = []zilliz.ServiceAccountProjectAssignment{
						{ProjectId: tc.project},
					}
				}
				api := &serviceAccountTestAPI{
					accounts: map[string]*zilliz.ServiceAccountResponse{"sa-imported": account},
					keys:     map[string]*zilliz.ServiceAccountKeyResponse{},
				}
				var mutations atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// The test harness deletes the imported resource during cleanup.
					if r.Method != http.MethodGet && r.Method != http.MethodDelete {
						mutations.Add(1)
					}
					api.serveHTTP(w, r)
				}))
				defer server.Close()
				config := fmt.Sprintf(`
provider "zillizcloud" {
  api_key      = "test"
  host_address = %q
}
resource "zillizcloud_service_account" "test" {
  name   = "imported"
  status = %q
  %s
}
`, server.URL+"/v2", status, tc.permissions)
				resource.UnitTest(t, resource.TestCase{
					ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						{
							Config:             config,
							ResourceName:       "zillizcloud_service_account.test",
							ImportState:        true,
							ImportStateId:      "sa-imported",
							ImportStatePersist: true,
						},
						{
							Config:   config,
							PlanOnly: true,
						},
						{
							Config: config,
							Check: resource.ComposeAggregateTestCheckFunc(
								resource.TestCheckResourceAttr("zillizcloud_service_account.test", "status", status),
								orgRoleCheck,
								resource.TestCheckResourceAttr("zillizcloud_service_account.test", "project_role_ids.#", "0"),
								resource.TestCheckResourceAttr("zillizcloud_service_account.test", "project_assignments.%", "0"),
							),
						},
					},
				})
				require.Zero(t, mutations.Load(), "import and unchanged apply must not modify permissions or status")
			})
		}
	}
}
