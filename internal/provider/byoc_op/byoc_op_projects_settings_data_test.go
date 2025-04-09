package byoc_op_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccBYOCOpProjectSettingsData(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_byoc_op_project_settings" "resource" {
  project_name    = "TestProject"
  region         = "us-west-2"
  cloud_provider = "aws"
  
  instances = {
    core_vm        = "m6i.2xlarge"
    fundamental_vm = "m6i.2xlarge"
    search_vm      = "m6id.2xlarge"
    search_vm_min_count = 1
    fundamental_vm_min_count = 1
    core_vm_min_count = 1
  }
	private_link_enabled = true
} 

data "zillizcloud_byoc_op_project_settings" "test" {
  project_id = zillizcloud_byoc_op_project_settings.resource.project_id
  data_plane_id = zillizcloud_byoc_op_project_settings.resource.data_plane_id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.zillizcloud_byoc_op_project_settings.test", "project_name", "TestProject"),
					resource.TestCheckResourceAttr("data.zillizcloud_byoc_op_project_settings.test", "region", "us-west-2"),
					resource.TestCheckResourceAttr("data.zillizcloud_byoc_op_project_settings.test", "cloud_provider", "aws"),
					resource.TestCheckResourceAttrSet("data.zillizcloud_byoc_op_project_settings.test", "id"),
					resource.TestCheckResourceAttrSet("data.zillizcloud_byoc_op_project_settings.test", "project_id"),
					resource.TestCheckResourceAttrSet("data.zillizcloud_byoc_op_project_settings.test", "data_plane_id"),
					// Check computed node quotas
					resource.TestCheckResourceAttrSet("data.zillizcloud_byoc_op_project_settings.test", "node_quotas.core.disk_size"),
					resource.TestCheckResourceAttrSet("data.zillizcloud_byoc_op_project_settings.test", "node_quotas.core.min_size"),
					resource.TestCheckResourceAttrSet("data.zillizcloud_byoc_op_project_settings.test", "node_quotas.core.max_size"),
					resource.TestCheckResourceAttrSet("data.zillizcloud_byoc_op_project_settings.test", "node_quotas.core.desired_size"),
					resource.TestCheckResourceAttrSet("data.zillizcloud_byoc_op_project_settings.test", "node_quotas.core.instance_types"),
					resource.TestCheckResourceAttrSet("data.zillizcloud_byoc_op_project_settings.test", "node_quotas.core.capacity_type"),
					// Check op_config
					resource.TestCheckResourceAttrSet("data.zillizcloud_byoc_op_project_settings.test", "op_config.token"),
					resource.TestCheckResourceAttrSet("data.zillizcloud_byoc_op_project_settings.test", "op_config.agent_image_url"),
					resource.TestCheckResourceAttr("data.zillizcloud_byoc_op_project_settings.test", "private_link_enabled", "true"),
				),
			},
		},
	})
}
