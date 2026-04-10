package cluster_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	provider "github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

// simulateConsoleAutoscaling calls the Zilliz Cloud API directly to simulate
// a user configuring cu autoscaling via the console (outside of Terraform).
func simulateConsoleAutoscaling(clusterID string, cuMin, cuMax int) error {
	host := os.Getenv("ZILLIZCLOUD_HOST_ADDRESS")
	if host == "" {
		host = "https://api.cloud.zilliz.com/v2"
	}
	apiKey := os.Getenv("ZILLIZCLOUD_API_KEY")

	body := fmt.Sprintf(`{"autoscaling":{"cu":{"min":%d,"max":%d}}}`, cuMin, cuMax)
	req, err := http.NewRequest("POST", host+"/clusters/"+clusterID+"/modify", bytes.NewBufferString(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}
	if result.Code != 0 {
		return fmt.Errorf("API error code %d: %s", result.Code, result.Message)
	}
	return nil
}

// extractClusterID is a Check helper that captures the cluster ID for use in PreConfig.
func extractClusterID(resourceName string, dst *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		*dst = rs.Primary.ID
		return nil
	}
}

// TestAccReadSyncsCuSettingsFromConsole verifies that when autoscaling is
// configured via the console (API) on a cluster that was created without
// cu_settings, Read() populates cu_settings from the API so that adding
// a matching cu_settings block in HCL results in no diff.
func TestAccReadSyncsCuSettingsFromConsole(t *testing.T) {
	t.Parallel()
	var clusterID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create cluster without cu_settings
			{
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "drift-console-add"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  cu_size      = 2
  project_id   = data.zillizcloud_project.default.id
  timeouts {
    create = "120m"
    update = "120m"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "status", "RUNNING"),
					extractClusterID("zillizcloud_cluster.test", &clusterID),
				),
			},
			// Step 2: Console configures autoscaling, HCL adds matching cu_settings.
			// After fix, Read syncs from API → state matches HCL → apply is no-op.
			{
				PreConfig: func() {
					if err := simulateConsoleAutoscaling(clusterID, 2, 4); err != nil {
						t.Fatalf("failed to simulate console autoscaling: %v", err)
					}
				},
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "drift-console-add"
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
		},
	})
}

// TestAccReadDetectsCuSettingsDrift verifies that when autoscaling is changed
// via the console after Terraform configured it, Read() detects the drift and
// the next apply reverts to the HCL-specified values.
func TestAccReadDetectsCuSettingsDrift(t *testing.T) {
	t.Parallel()
	var clusterID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with cu_settings { min=2, max=4 }
			{
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "drift-console-change"
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
					extractClusterID("zillizcloud_cluster.test", &clusterID),
				),
			},
			// Step 2: Console changes autoscaling to { min=4, max=8 }.
			// Same HCL { min=2, max=4 } → Read detects drift → apply reverts.
			{
				PreConfig: func() {
					if err := simulateConsoleAutoscaling(clusterID, 4, 8); err != nil {
						t.Fatalf("failed to simulate console autoscaling change: %v", err)
					}
				},
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "drift-console-change"
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
					// After apply, values should be reverted to HCL values
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_settings.dynamic_scaling.min", "2"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_settings.dynamic_scaling.max", "4"),
				),
			},
		},
	})
}

// TestAccReadDetectsCuSettingsDriftThenUserUpdatesHCL verifies the scenario
// where autoscaling is changed via console, and then the user updates their
// HCL to a different (third) value. The apply should set the HCL value.
func TestAccReadDetectsCuSettingsDriftThenUserUpdatesHCL(t *testing.T) {
	t.Parallel()
	var clusterID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with cu_settings { min=2, max=4 }
			{
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "drift-user-update"
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
					extractClusterID("zillizcloud_cluster.test", &clusterID),
				),
			},
			// Step 2: Console changes to { min=4, max=8 }, user updates HCL to { min=2, max=8 }.
			{
				PreConfig: func() {
					if err := simulateConsoleAutoscaling(clusterID, 4, 8); err != nil {
						t.Fatalf("failed to simulate console autoscaling change: %v", err)
					}
				},
				Config: provider.ProviderConfig + `
data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "test" {
  cluster_name = "drift-user-update"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  cu_settings = {
    dynamic_scaling = {
      min = 2
      max = 8
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
					// Should reflect the user's HCL values, not the console values
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_settings.dynamic_scaling.min", "2"),
					resource.TestCheckResourceAttr("zillizcloud_cluster.test", "cu_settings.dynamic_scaling.max", "8"),
				),
			},
		},
	})
}
