package byoc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccBYOCProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_byoc_project" "test" {
  name = "TestProject"

  aws = {
    region = "aws-us-west-2"

    network = {
      vpc_id = "vpc-06d74ec11c83c2da2"
      subnet_ids = [
        "subnet-01c2a9d595eb577ff",
        "subnet-0ef457de4d79e98b6",
        "subnet-0fb9665409f2a96f5",
      ]
      security_group_ids = ["sg-005f7dd3e825ad555"]
      vpc_endpoint_id    = "vpce-12345678"
    }
    role_arn = {
      storage       = "arn:aws:iam::041623484421:role/test-storage-role"
      eks           = "arn:aws:iam::041623484421:role/test-eks-role"
      cross_account = "arn:aws:iam::041623484421:role/test-cross-account-role"
    }
    storage = {
      bucket_id = "test-bucket"
    }

    instances = {
      core_vm        = "m6i.2xlarge"
      fundamental_vm = "m6i.2xlarge"
      search_vm      = "m6id.2xlarge"
    }
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_byoc_project.test", "name", "TestProject"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_project.test", "aws.region", "aws-us-west-2"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_project.test", "aws.network.vpc_id", "vpc-06d74ec11c83c2da2"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_project.test", "aws.network.vpc_endpoint_id", "vpce-12345678"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_project.test", "aws.storage.bucket_id", "test-bucket"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_project.test", "aws.instances.core_vm", "m6i.2xlarge"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_project.test", "aws.instances.fundamental_vm", "m6i.2xlarge"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_project.test", "aws.instances.search_vm", "m6id.2xlarge"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_project.test", "id"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_project.test", "data_plane_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_project.test", "status"),
				),
				PreventPostDestroyRefresh: true,
			},
		},
	})
}
