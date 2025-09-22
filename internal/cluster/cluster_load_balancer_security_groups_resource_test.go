package cluster_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccClusterLoadBalancerSecurityGroupsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
  id = "proj-test123456789"
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "test-cluster-sg"
  project_id   = data.zillizcloud_project.default.id
  plan         = "Standard"
  cu_size      = 1
  cu_type      = "Performance-optimized"
}

resource "zillizcloud_cluster_load_balancer_security_groups" "test" {
  cluster_id = zillizcloud_cluster.test.id

  security_group_ids = [
    "sg-test123456789abcdef0",
    "sg-test987654321fedcba",
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("zillizcloud_cluster_load_balancer_security_groups.test", "cluster_id", "zillizcloud_cluster.test", "id"),
					resource.TestCheckResourceAttr("zillizcloud_cluster_load_balancer_security_groups.test", "security_group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr("zillizcloud_cluster_load_balancer_security_groups.test", "security_group_ids.*", "sg-test123456789abcdef0"),
					resource.TestCheckTypeSetElemAttr("zillizcloud_cluster_load_balancer_security_groups.test", "security_group_ids.*", "sg-test987654321fedcba"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "zillizcloud_cluster_load_balancer_security_groups.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
  id = "proj-test123456789"
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "test-cluster-sg"
  project_id   = data.zillizcloud_project.default.id
  plan         = "Standard"
  cu_size      = 1
  cu_type      = "Performance-optimized"
}

resource "zillizcloud_cluster_load_balancer_security_groups" "test" {
  cluster_id = zillizcloud_cluster.test.id

  security_group_ids = [
    "sg-test123456789abcdef0",
    "sg-testnew111222333444",
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("zillizcloud_cluster_load_balancer_security_groups.test", "cluster_id", "zillizcloud_cluster.test", "id"),
					resource.TestCheckResourceAttr("zillizcloud_cluster_load_balancer_security_groups.test", "security_group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr("zillizcloud_cluster_load_balancer_security_groups.test", "security_group_ids.*", "sg-test123456789abcdef0"),
					resource.TestCheckTypeSetElemAttr("zillizcloud_cluster_load_balancer_security_groups.test", "security_group_ids.*", "sg-testnew111222333444"),
				),
			},
			// Test with single security group
			{
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
  id = "proj-test123456789"
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "test-cluster-sg"
  project_id   = data.zillizcloud_project.default.id
  plan         = "Standard"
  cu_size      = 1
  cu_type      = "Performance-optimized"
}

resource "zillizcloud_cluster_load_balancer_security_groups" "test" {
  cluster_id = zillizcloud_cluster.test.id

  security_group_ids = [
    "sg-testsingle123456789",
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("zillizcloud_cluster_load_balancer_security_groups.test", "cluster_id", "zillizcloud_cluster.test", "id"),
					resource.TestCheckResourceAttr("zillizcloud_cluster_load_balancer_security_groups.test", "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("zillizcloud_cluster_load_balancer_security_groups.test", "security_group_ids.*", "sg-testsingle123456789"),
				),
			},
		},
	})
}
