package byoc_op_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccByocOpProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccByocOpProjectConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_settings.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project_settings.test", "data_plane_id"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_op_project_settings.test", "cloud_provider", "aws"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_op_project_settings.test", "region", "aws-us-west-2"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_op_project_settings.test", "project_name", "byoc-zilliz-test"),
					// asert status
					resource.TestCheckResourceAttr("zillizcloud_byoc_op_project.test", "status", "1"),

					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_op_project.test", "data_plane_id"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_op_project.test", "aws.region", "aws-us-west-2"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_op_project.test", "aws.network.vpc_id", "vpc-0a9a08f81e27c2b69"),
				),
			},
		},
	})
}

func testAccByocOpProjectConfig() string {
	return `
resource "zillizcloud_byoc_op_project_settings" "test" {
    cloud_provider = "aws"
    region = "aws-us-west-2"
    project_name = "byoc-zilliz-test"

    instances = {
        core_vm        = "m6i.2xlarge"
        fundamental_vm = "m6i.2xlarge"
        search_vm      = "m6id.2xlarge"
    }
}

resource "zillizcloud_byoc_op_project_agent" "test" {
    project_id = zillizcloud_byoc_op_project_settings.test.project_id
    data_plane_id = zillizcloud_byoc_op_project_settings.test.data_plane_id
}

resource "zillizcloud_byoc_op_project" "test" {
    lifecycle {
        ignore_changes = [data_plane_id, project_id, aws, ext_config]
    }

    data_plane_id = zillizcloud_byoc_op_project_settings.test.data_plane_id
    project_id = zillizcloud_byoc_op_project_settings.test.project_id
   
    aws = {
        region = "aws-us-west-2"
        network = {
            vpc_id = "vpc-0a9a08f81e27c2b69"
            subnet_ids = ["subnet-0d352e35a68f2f7c7", "subnet-03d0d894d05e0b87b", "subnet-08cf896411a229c8e"]
            security_group_ids = ["sg-02f41431f91303644"]
        }
        role_arn = {
            storage = "arn:aws:iam::999999999999:role/zilliz-stack-17d586-ZillizStorageRole-1jSYHHFWhGkz"
            eks = "arn:aws:iam::999999999999:role/zilliz-stack-17d586-ZillizEKSRole-D27XZP0XK5do"
            cross_account = "arn:aws:iam::999999999999:role/zilliz-stack-17d586-ZillizBootstrapRole-DAyuQSLZEN9g"
        }
        storage = {
            bucket_id = "zilliz-s3-0af21b"
        }

    }

    depends_on = [zillizcloud_byoc_op_project_settings.test, zillizcloud_byoc_op_project_agent.test]
}
`
}
