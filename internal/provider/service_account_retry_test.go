package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestServiceAccountRetryAfterCommittedSyncFailure(t *testing.T) {
	for _, roles := range []bool{
		false,
		true,
	} {
		t.Run(fmt.Sprintf("roles=%t", roles), func(t *testing.T) {
			api, _ := newServiceAccountTestAPI(t)
			mutations := 0
			pendingSync := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch && r.Method != http.MethodPut {
					api.serveHTTP(w, r)
					return
				}
				// The API commits before publishing key availability/permissions.
				// A GET can see the new values even when publication failed.
				recorder := httptest.NewRecorder()
				api.serveHTTP(recorder, r)
				api.mu.Lock()
				defer api.mu.Unlock()
				mutations++
				if mutations == 1 {
					pendingSync = true
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprint(w, `{"code":500,"message":"key sync failed after database commit"}`)
					return
				}
				if recorder.Code == http.StatusOK {
					pendingSync = false
				}
				w.WriteHeader(recorder.Code)
				_, _ = w.Write(recorder.Body.Bytes())
			}))
			defer server.Close()
			config := func(updated bool) string {
				status, grants := "active", "project_assignments = {}"
				if roles {
					if !updated {
						grants = `org_role_id = "r-billing-admin"`
					}
				} else if updated {
					status = "disable"
				}
				return fmt.Sprintf(`
provider "zillizcloud" {
  api_key = "test"
  host_address = %q
}
resource "zillizcloud_service_account" "test" {
  name = "test"
  %s
  status = %q
}
`, server.URL+"/v2", grants, status)
			}
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{Config: config(false)},
					{
						Config: config(true),
						Check: func(_ *terraform.State) error {
							api.mu.Lock()
							defer api.mu.Unlock()
							if pendingSync || mutations != 2 {
								return fmt.Errorf("sync was not retried before apply completed: pending=%t, mutation calls=%d", pendingSync, mutations)
							}
							return nil
						},
					},
				},
			})
		})
	}
}
