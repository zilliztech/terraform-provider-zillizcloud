package cluster_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	provider "github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

// Tests for PR#205: implicit cu_size, replica on create, combined autoscaling API call.

// TestAccImplicitCuSizeFromAutoscaleMin verifies that when cu_settings.dynamic_scaling.min
// is set without an explicit cu_size, the cluster is created with cu_size = min.
func TestAccImplicitCuSizeFromAutoscaleMin(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "implicit-cusize-test"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  cu_settings  = {
    dynamic_scaling = {
      min = 4
      max = 8
    }
  }
  timeouts {
    create = "120m"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "status", "RUNNING"),
					// cu_size should equal cu_settings.dynamic_scaling.min
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_size", "4"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_settings.dynamic_scaling.min", "4"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_settings.dynamic_scaling.max", "8"),
				),
			},
		},
	})
}

// TestAccReplicaOnCreate verifies that replica > 1 can be set during creation.
// The provider creates the cluster, waits for RUNNING, then calls ModifyReplica.
// Requires cu_size >= 12 for multi-replica.
func TestAccReplicaOnCreate(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "replica-on-create-test"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  cu_settings  = {
    dynamic_scaling = {
      min = 12
      max = 16
    }
  }
  replica = 2
  timeouts {
    create = "120m"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "status", "RUNNING"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_size", "12"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "replica", "2"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_settings.dynamic_scaling.min", "12"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_settings.dynamic_scaling.max", "16"),
				),
			},
		},
	})
}

// TestAccCombinedCuAndReplicaSettings verifies that cu_settings and replica_settings
// are sent in a single API call so neither overwrites the other.
func TestAccCombinedCuAndReplicaSettings(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "test" {
  cluster_name     = "combined-settings-test"
  region_id        = "aws-us-west-2"
  plan             = "Enterprise"
  cu_type          = "Performance-optimized"
  project_id       = data.zillizcloud_project.default.id
  cu_settings = {
    dynamic_scaling = {
      min = 4
      max = 8
    }
  }
  replica_settings = {
    dynamic_scaling = {
      min = 1
      max = 3
    }
  }
  timeouts {
    create = "120m"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "status", "RUNNING"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_size", "4"),
					// Both cu and replica autoscaling must be present (not overwritten)
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_settings.dynamic_scaling.min", "4"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_settings.dynamic_scaling.max", "8"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "replica_settings.dynamic_scaling.min", "1"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "replica_settings.dynamic_scaling.max", "3"),
				),
			},
		},
	})
}

// TestAccUpdateAddReplicaSettingsToExistingCuSettings verifies that adding
// replica_settings to a cluster that already has cu_settings does not remove
// the existing cu_settings (combined API call on update).
func TestAccUpdateAddReplicaSettingsToExistingCuSettings(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create with cu_settings only
			{
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "update-add-replica-test"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  cu_settings = {
    dynamic_scaling = {
      min = 2
      max = 4
    }
  }
  timeouts {
    create = "120m"
    update = "120m"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "status", "RUNNING"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_settings.dynamic_scaling.min", "2"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_settings.dynamic_scaling.max", "4"),
				),
			},
			// Step 2: add replica_settings — cu_settings must be preserved
			{
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "update-add-replica-test"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  cu_settings = {
    dynamic_scaling = {
      min = 2
      max = 4
    }
  }
  replica_settings = {
    dynamic_scaling = {
      min = 1
      max = 2
    }
  }
  timeouts {
    create = "120m"
    update = "120m"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "status", "RUNNING"),
					// cu_settings must still be present after adding replica_settings
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_settings.dynamic_scaling.min", "2"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_settings.dynamic_scaling.max", "4"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "replica_settings.dynamic_scaling.min", "1"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "replica_settings.dynamic_scaling.max", "2"),
				),
			},
		},
	})
}
