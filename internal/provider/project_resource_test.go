package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_project" "test" {
  project_name = "test-project-terraform"
  plan         = "Standard"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_project.test", "project_name", "test-project-terraform"),
					resource.TestCheckResourceAttr("zillizcloud_project.test", "plan", "Standard"),
					resource.TestCheckResourceAttrSet("zillizcloud_project.test", "id"),
				),
			},
			// Update and Read testing - upgrade plan
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_project" "test" {
  project_name = "test-project-terraform"
  plan         = "Enterprise"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_project.test", "project_name", "test-project-terraform"),
					resource.TestCheckResourceAttr("zillizcloud_project.test", "plan", "Enterprise"),
					resource.TestCheckResourceAttrSet("zillizcloud_project.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "zillizcloud_project.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources["zillizcloud_project.test"]
					if !ok {
						return "", fmt.Errorf("zillizcloud_project.test not found")
					}
					return rs.Primary.Attributes["id"], nil
				},
			},
		},
	})
}

func TestAccProjectResource_DefaultPlan(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create project without specifying plan - should use default "Enterprise"
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_project" "test" {
  project_name = "test-project-default-plan"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_project.test", "project_name", "test-project-default-plan"),
					resource.TestCheckResourceAttr("zillizcloud_project.test", "plan", "Enterprise"),
					resource.TestCheckResourceAttrSet("zillizcloud_project.test", "id"),
				),
			},
		},
	})
}
