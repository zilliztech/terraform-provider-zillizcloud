package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccClusterResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "TestCluster"
  plan         = "Serverless"
  project_id   = data.zillizcloud_project.default.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cluster_name", "TestCluster"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "plan", "Serverless"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "status", "RUNNING"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "username"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "password"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "prompt"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "connect_address"),
				),
				PreventPostDestroyRefresh: true,
			},
			{
				ResourceName:            "zillizcloud_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "prompt", "username"},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources["zillizcloud_cluster.test"]
					if !ok {
						return "", fmt.Errorf("zillizcloud_cluster.test not found")
					}
					clusterId := rs.Primary.Attributes["id"]
					regionId := rs.Primary.Attributes["region_id"]
					//        Expected import identifier with format: clusterId,regionId
					return fmt.Sprintf("%s,%s", clusterId, regionId), nil
				},
			},
		},
	})
}

// Append a new test function to cover the Standard plan.
func TestAccClusterResourceStandardPlan(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "a-standard-cluster"
  plan         = "Standard"
  cu_size      = "1"                                 # The size of the compute unit
  cu_type      = "Performance-optimized"             # The type of compute unit, optimized for performance
  project_id   = data.zillizcloud_project.default.id
  timeouts {
    create = "10m"
	update = "10m"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cluster_name", "a-standard-cluster"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "plan", "Standard"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "status", "RUNNING"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_size", "1"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_type", "Performance-optimized"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "username"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "password"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "prompt"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "connect_address"),
				),
				PreventPostDestroyRefresh: true,
			},
			// Test import
			{
				ResourceName:            "zillizcloud_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "prompt", "username"},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources["zillizcloud_cluster.test"]
					if !ok {
						return "", fmt.Errorf("zillizcloud_cluster.test not found")
					}
					clusterId := rs.Primary.Attributes["id"]
					regionId := rs.Primary.Attributes["region_id"]
					return fmt.Sprintf("%s,%s", clusterId, regionId), nil
				},
			},
			// Test update
			{
				Config: provider.ProviderConfig + `
				resource "zillizcloud_cluster" "test" {
					cluster_name = "a-standard-cluster"
					plan         = "Standard"
					cu_size      = "2"                                 # The size of the compute unit
					cu_type      = "Performance-optimized"             # The type of compute unit, optimized for performance
					project_id   = data.zillizcloud_project.default.id
					timeouts {
						create = "10m"
						update = "10m"
					}
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cluster_name", "a-standard-cluster"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "plan", "Standard"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "status", "RUNNING"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_size", "2"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_type", "Performance-optimized"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "username"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "password"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "prompt"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "connect_address"),
				),
			},
		},
	})
}
