package cluster_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccClusterResource(t *testing.T) {

	t.Run("SaaSEnv", func(t *testing.T) {

		t.Run("FreePlan", testAccClusterResourceFreePlan)
		t.Run("ServerlessPlan", testAccClusterResourceServerlessPlan)
		t.Run("StandardPlan", testAccClusterResourceStandardPlan)
	})
	t.Run("BYOCEnv", func(t *testing.T) {
		t.Run("UpdateLabels", testAccClusterResourceUpdateLabels)
	})
}

func testAccClusterResourceFreePlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "TestCluster"
  plan         = "Free"
  region_id    = "gcp-us-west1"
  project_id   = data.zillizcloud_project.default.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cluster_name", "TestCluster"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "plan", "Free"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "status", "RUNNING"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "connect_address"),
				),
				PreventPostDestroyRefresh: true,
			},
			{
				ResourceName:            "zillizcloud_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "prompt", "username", "replica"},
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

func testAccClusterResourceServerlessPlan(t *testing.T) {
	t.Parallel()
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
  region_id    = "gcp-us-west1"
  project_id   = data.zillizcloud_project.default.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cluster_name", "TestCluster"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "plan", "Serverless"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "status", "RUNNING"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "connect_address"),
				),
				PreventPostDestroyRefresh: true,
			},
			{
				ResourceName:            "zillizcloud_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "prompt", "username", "replica"},
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
func testAccClusterResourceStandardPlan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "a-standard-cluster"
  region_id    = "aws-us-west-2"
  plan         = "Standard"
  cu_size      = "1"                                 
  cu_type      = "Performance-optimized"             
  project_id   = data.zillizcloud_project.default.id
  timeouts {
	create = "120m"
	update = "120m"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cluster_name", "a-standard-cluster"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "plan", "Standard"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "status", "RUNNING"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "replica", "1"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_size", "1"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_type", "Performance-optimized"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "project_id"),
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
			// suspend the cluster
			{
				Config: provider.ProviderConfig + `
		        data "zillizcloud_project" "default" {
				}
				resource "zillizcloud_cluster" "test" {
				    cluster_name = "a-standard-cluster"
					region_id    = "aws-us-west-2"
					plan         = "Standard"
					cu_size      = "1"                                 
					cu_type      = "Performance-optimized"             
					project_id   = data.zillizcloud_project.default.id
					timeouts {
						create = "120m"
						update = "120m"
					}

					desired_status = "SUSPENDED" # suspend the cluster
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "status", "SUSPENDED"),
				),
			},
			// resume the cluster
			{
				Config: provider.ProviderConfig + `
		        data "zillizcloud_project" "default" {
				}
				resource "zillizcloud_cluster" "test" {
				    cluster_name = "a-standard-cluster"
					region_id    = "aws-us-west-2"
					plan         = "Standard"
					cu_size      = "1"                                 
					cu_type      = "Performance-optimized"             
					project_id   = data.zillizcloud_project.default.id
					timeouts {
						create = "120m"
						update = "120m"
					}

					desired_status = "RUNNING" # resume the cluster
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "status", "RUNNING"),
				),
			},
			// Test update
			{
				Config: provider.ProviderConfig + `
				data "zillizcloud_project" "default" {
				}

				resource "zillizcloud_cluster" "test" {
					cluster_name = "a-standard-cluster"
					region_id    = "aws-us-west-2"
					plan         = "Standard"
					cu_size      = "2"                                 # change the cu_size
					replica      = "8"                                 # change the replica
					cu_type      = "Performance-optimized"             
					project_id   = data.zillizcloud_project.default.id
					timeouts {
						create = "120m"
						update = "120m"
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
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "connect_address"),
				),
			},
		},
	})
}

// test update labels
func testAccClusterResourceUpdateLabels(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provider.ProviderConfig + `
					data "zillizcloud_project" "default" {
					}

					resource "zillizcloud_cluster" "test" {
					cluster_name = "a-byoc-cluster"
					project_id   = data.zillizcloud_project.default.id
					labels = {
						"key1" = "value1"
						"key2" = "value2"
					}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "labels.key1", "value1"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "labels.key2", "value2"),
				),
			},
			// update labels
			{
				Config: provider.ProviderConfig + `
				data "zillizcloud_project" "default" {
				}

				resource "zillizcloud_cluster" "test" {
					cluster_name = "a-byoc-cluster"
					project_id   = data.zillizcloud_project.default.id
					labels = {
						"key2" = "val2"
						"key3" = "value3"
					}
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "labels.key2", "val2"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "labels.key3", "value3"),
				),
			},
			// delete labels
			{
				Config: provider.ProviderConfig + `
				data "zillizcloud_project" "default" {
				}

				resource "zillizcloud_cluster" "test" {
					cluster_name = "a-byoc-cluster"
					project_id   = data.zillizcloud_project.default.id
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// no attribute should be set
					resource.TestCheckNoResourceAttr("zillizcloud_cluster.test", "labels.%"),
				),
			},
		},
	})
}
