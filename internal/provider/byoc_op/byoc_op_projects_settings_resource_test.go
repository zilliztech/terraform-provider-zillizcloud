package byoc_op_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccBYOCOpProjectSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_byoc_op_project_settings" "test" {
  project_name    = "TestProject"
  region         = "us-west-2"
  cloud_provider = "aws"
  
  instances = {
    core_vm        = "m6i.2xlarge"
    fundamental_vm = "m6i.2xlarge"
    search_vm      = "m6id.2xlarge"
  }
} 
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_byoc_op_project_settings.test", "project_name", "TestProject"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_op_project_settings.test", "region", "us-west-2"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_op_project_settings.test", "cloud_provider", "aws"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_settings.test", "id"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_settings.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_settings.test", "data_plane_id"),
					// Check computed node quotas
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_settings.test", "node_quotas.core.disk_size"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_settings.test", "node_quotas.core.min_size"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_settings.test", "node_quotas.core.max_size"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_settings.test", "node_quotas.core.desired_size"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_settings.test", "node_quotas.core.instance_types"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_settings.test", "node_quotas.core.capacity_type"),
					// Check op_config
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_settings.test", "op_config.token"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_settings.test", "op_config.agent_image_url"),
				),
			},
		},
	})
}
