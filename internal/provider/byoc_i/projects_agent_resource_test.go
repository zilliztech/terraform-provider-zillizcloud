package byoc_op_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccBYOCOpProjectAgentResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccBYOCOpProjectAgentResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_agent.test", "id"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_agent.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_agent.test", "data_plane_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_agent.test", "status"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_op_project_agent.test", "wait_until_ready", "false"),
				),
				Destroy: false,
			},
		},
	})
}

func testAccBYOCOpProjectAgentResourceConfig() string {
	return `
resource "zillizcloud_byoc_op_project_settings" "test" {
	cloud_provider = "aws"
	region = "aws-us-west-2"
	project_name = "byoc-zilliz-test-acc"
	instances = {
		core_vm        = "m6i.2xlarge"
		fundamental_vm = "m6i.2xlarge"
		search_vm      = "m6id.2xlarge"
	}
}

resource "zillizcloud_byoc_op_project_agent" "test" {
	project_id = zillizcloud_byoc_op_project_settings.test.project_id
	data_plane_id = zillizcloud_byoc_op_project_settings.test.data_plane_id
	wait_until_ready = false
}
`
}
