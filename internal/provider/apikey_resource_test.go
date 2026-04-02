package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

const testProjectId = "proj-dee71c5a02aee5d781b156"

func TestAccApiKeyResource_Member(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create Member API key with project access
			{
				Config: provider.ProviderConfig + fmt.Sprintf(`
resource "zillizcloud_api_key" "test" {
  name = "tf-acc-test-member"
  role = "Member"

  project_access {
    project_id  = %q
    role        = "Admin"
    all_cluster = true
  }
}
`, testProjectId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("zillizcloud_api_key.test", "id"),
					resource.TestCheckResourceAttr("zillizcloud_api_key.test", "name", "tf-acc-test-member"),
					resource.TestCheckResourceAttr("zillizcloud_api_key.test", "role", "Member"),
					resource.TestCheckResourceAttrSet("zillizcloud_api_key.test", "key_value"),
					resource.TestCheckResourceAttrSet("zillizcloud_api_key.test", "create_time"),
					resource.TestCheckResourceAttr("zillizcloud_api_key.test", "project_access.#", "1"),
					resource.TestCheckResourceAttr("zillizcloud_api_key.test", "project_access.0.project_id", testProjectId),
					resource.TestCheckResourceAttr("zillizcloud_api_key.test", "project_access.0.role", "Admin"),
					resource.TestCheckResourceAttr("zillizcloud_api_key.test", "project_access.0.all_cluster", "true"),
				),
			},
			// Step 2: Update name
			{
				Config: provider.ProviderConfig + fmt.Sprintf(`
resource "zillizcloud_api_key" "test" {
  name = "tf-acc-test-renamed"
  role = "Member"

  project_access {
    project_id  = %q
    role        = "Admin"
    all_cluster = true
  }
}
`, testProjectId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_api_key.test", "name", "tf-acc-test-renamed"),
					resource.TestCheckResourceAttr("zillizcloud_api_key.test", "role", "Member"),
					resource.TestCheckResourceAttrSet("zillizcloud_api_key.test", "key_value"),
				),
			},
			// Step 3: Update role to Owner (project_access removed)
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_api_key" "test" {
  name = "tf-acc-test-renamed"
  role = "Owner"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_api_key.test", "name", "tf-acc-test-renamed"),
					resource.TestCheckResourceAttr("zillizcloud_api_key.test", "role", "Owner"),
					resource.TestCheckResourceAttr("zillizcloud_api_key.test", "project_access.#", "0"),
				),
			},
			// Step 4: Import
			{
				ResourceName:      "zillizcloud_api_key.test",
				ImportState:       true,
				ImportStateVerify: true,
				// key_value is not retrievable after creation, so skip verification
				ImportStateVerifyIgnore: []string{"key_value"},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources["zillizcloud_api_key.test"]
					if !ok {
						return "", fmt.Errorf("zillizcloud_api_key.test not found")
					}
					return rs.Primary.Attributes["id"], nil
				},
			},
		},
	})
}

func TestAccApiKeyResource_Owner(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create Owner key without project access
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_api_key" "owner" {
  name = "tf-acc-test-owner"
  role = "Owner"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("zillizcloud_api_key.owner", "id"),
					resource.TestCheckResourceAttr("zillizcloud_api_key.owner", "name", "tf-acc-test-owner"),
					resource.TestCheckResourceAttr("zillizcloud_api_key.owner", "role", "Owner"),
					resource.TestCheckResourceAttrSet("zillizcloud_api_key.owner", "key_value"),
					resource.TestCheckResourceAttr("zillizcloud_api_key.owner", "project_access.#", "0"),
				),
			},
		},
	})
}

func TestAccApiKeyResource_ProjectRoles(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with Read-Only project role
			{
				Config: provider.ProviderConfig + fmt.Sprintf(`
resource "zillizcloud_api_key" "readonly" {
  name = "tf-acc-test-readonly"
  role = "Member"

  project_access {
    project_id  = %q
    role        = "Read-Only"
    all_cluster = true
  }
}
`, testProjectId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_api_key.readonly", "project_access.0.role", "Read-Only"),
				),
			},
			// Update to Read-Write
			{
				Config: provider.ProviderConfig + fmt.Sprintf(`
resource "zillizcloud_api_key" "readonly" {
  name = "tf-acc-test-readonly"
  role = "Member"

  project_access {
    project_id  = %q
    role        = "Read-Write"
    all_cluster = true
  }
}
`, testProjectId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_api_key.readonly", "project_access.0.role", "Read-Write"),
				),
			},
		},
	})
}
