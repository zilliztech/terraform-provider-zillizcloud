package cluster_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
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
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "connect_address"),
					waitForClusterRunning("zillizcloud_cluster.test"),
				),
				PreventPostDestroyRefresh: true,
			},
			{
				RefreshState: true,
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
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "connect_address"),
					waitForClusterRunning("zillizcloud_cluster.test"),
				),
				PreventPostDestroyRefresh: true,
			},
			{
				RefreshState: true,
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
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "status", "CREATING"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "replica", "1"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_size", "1"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_type", "Performance-optimized"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_cluster.test", "connect_address"),
					// wait for cluster to be RUNNING
					waitForClusterRunning("zillizcloud_cluster.test"),
				),
				PreventPostDestroyRefresh: true,
			},
			{
				RefreshState: true,
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
					cluster_name = "a-standard-cluster-renamed"        # change the cluster name
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
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cluster_name", "a-standard-cluster-renamed"),
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
					waitForClusterRunning("zillizcloud_cluster.test"),
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

// waitForClusterRunning waits for the cluster to reach RUNNING status
//
//nolint:unparam
func waitForClusterRunning(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		clusterId := rs.Primary.ID
		if clusterId == "" {
			return fmt.Errorf("cluster ID is empty")
		}

		// Create a new client for testing
		apiKey := os.Getenv("ZILLIZCLOUD_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("ZILLIZCLOUD_API_KEY environment variable is required for testing")
		}

		client, err := zilliz.NewClient(
			zilliz.WithApiKey(apiKey),
			zilliz.WithHostAddress("http://127.0.0.1:8080/v2"))
		if err != nil {
			return fmt.Errorf("failed to create client: %v", err)
		}

		// Poll for up to 10 minutes
		timeout := 10 * time.Minute
		pollInterval := 3 * time.Second

		start := time.Now()
		for time.Since(start) < timeout {
			cluster, err := client.DescribeCluster(clusterId)
			if err != nil {
				return fmt.Errorf("failed to describe cluster: %v", err)
			}

			if cluster.Status == "RUNNING" {
				return nil
			}

			time.Sleep(pollInterval)
		}

		return fmt.Errorf("cluster did not reach RUNNING status within %v", timeout)
	}
}
