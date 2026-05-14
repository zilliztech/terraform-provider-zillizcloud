package provider_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccVolumeResourceManaged(t *testing.T) {
	projectID, regionID, _ := testAccVolumeConfigValues(t, false)
	volumeName := "tf-acc-managed-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	replacementName := volumeName + "-replacement"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeManagedConfig(projectID, regionID, volumeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_volume.test", "id", volumeName),
					resource.TestCheckResourceAttr("zillizcloud_volume.test", "project_id", projectID),
					resource.TestCheckResourceAttr("zillizcloud_volume.test", "region_id", regionID),
					resource.TestCheckResourceAttr("zillizcloud_volume.test", "volume_name", volumeName),
					resource.TestCheckResourceAttr("zillizcloud_volume.test", "type", "MANAGED"),
					resource.TestCheckResourceAttrSet("zillizcloud_volume.test", "status"),
					resource.TestCheckResourceAttrSet("zillizcloud_volume.test", "create_time"),
				),
			},
			{
				Config: testAccVolumeManagedConfig(projectID, regionID, replacementName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_volume.test", "id", replacementName),
					resource.TestCheckResourceAttr("zillizcloud_volume.test", "volume_name", replacementName),
					resource.TestCheckResourceAttr("zillizcloud_volume.test", "type", "MANAGED"),
				),
			},
		},
	})
}

func TestAccVolumeResourceExternal(t *testing.T) {
	projectID, regionID, storageIntegrationID := testAccVolumeConfigValues(t, true)
	volumeName := "tf-acc-external-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	path := fmt.Sprintf("terraform/%s/", volumeName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeExternalConfig(projectID, regionID, volumeName, storageIntegrationID, path),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_volume.test", "id", volumeName),
					resource.TestCheckResourceAttr("zillizcloud_volume.test", "project_id", projectID),
					resource.TestCheckResourceAttr("zillizcloud_volume.test", "region_id", regionID),
					resource.TestCheckResourceAttr("zillizcloud_volume.test", "volume_name", volumeName),
					resource.TestCheckResourceAttr("zillizcloud_volume.test", "type", "EXTERNAL"),
					resource.TestCheckResourceAttr("zillizcloud_volume.test", "storage_integration_id", storageIntegrationID),
					resource.TestCheckResourceAttr("zillizcloud_volume.test", "path", path),
					resource.TestCheckResourceAttrSet("zillizcloud_volume.test", "status"),
					resource.TestCheckResourceAttrSet("zillizcloud_volume.test", "create_time"),
				),
			},
		},
	})
}

func testAccVolumeConfigValues(t *testing.T, external bool) (string, string, string) {
	t.Helper()

	usesMockServer := testAccVolumeUsesMockServer()
	if os.Getenv("ZILLIZCLOUD_API_KEY") == "" {
		if !usesMockServer {
			t.Skip("ZILLIZCLOUD_API_KEY must be set for volume acceptance tests")
		}
		t.Setenv("ZILLIZCLOUD_API_KEY", "test-key")
	}

	projectID := os.Getenv("ZILLIZCLOUD_PROJECT_ID")
	if projectID == "" {
		if !usesMockServer {
			t.Skip("ZILLIZCLOUD_PROJECT_ID must be set for volume acceptance tests")
		}
		projectID = "proj-acceptance"
	}

	regionID := os.Getenv("ZILLIZCLOUD_REGION_ID")
	if regionID == "" {
		if !usesMockServer {
			t.Skip("ZILLIZCLOUD_REGION_ID must be set for volume acceptance tests")
		}
		regionID = "aws-us-west-2"
	}

	storageIntegrationID := os.Getenv("ZILLIZCLOUD_STORAGE_INTEGRATION_ID")
	if external && storageIntegrationID == "" {
		if !usesMockServer {
			t.Skip("ZILLIZCLOUD_STORAGE_INTEGRATION_ID must be set for external volume acceptance tests")
		}
		storageIntegrationID = "si-acceptance"
	}

	return projectID, regionID, storageIntegrationID
}

func testAccVolumeUsesMockServer() bool {
	host := os.Getenv("ZILLIZCLOUD_HOST_ADDRESS")
	return strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1")
}

func testAccVolumeManagedConfig(projectID, regionID, volumeName string) string {
	return provider.ProviderConfig + fmt.Sprintf(`
resource "zillizcloud_volume" "test" {
  project_id   = %[1]q
  region_id    = %[2]q
  volume_name  = %[3]q
}
`, projectID, regionID, volumeName)
}

func testAccVolumeExternalConfig(projectID, regionID, volumeName, storageIntegrationID, path string) string {
	return provider.ProviderConfig + fmt.Sprintf(`
resource "zillizcloud_volume" "test" {
  project_id              = %[1]q
  region_id               = %[2]q
  volume_name             = %[3]q
  type                    = "EXTERNAL"
  storage_integration_id  = %[4]q
  path                    = %[5]q
}
`, projectID, regionID, volumeName, storageIntegrationID, path)
}
