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
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_i_project_settings.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_i_project_settings.test", "data_plane_id"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_i_project_settings.test", "cloud_provider", "aws"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_i_project_settings.test", "region", "aws-us-west-2"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_i_project_settings.test", "project_name", "byoc-zilliz-test"),
					// asert status
					resource.TestCheckResourceAttr("zillizcloud_byoc_i_project.test", "status", "0"),

					resource.TestCheckResourceAttrSet("zillizcloud_byoc_i_project.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_i_project.test", "data_plane_id"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_i_project.test", "aws.region", "aws-us-west-2"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_i_project.test", "aws.network.vpc_id", "vpc-0a9a08f81e27c2b69"),
				),
			},
		},
	})
}

func testAccByocOpProjectConfig() string {
	return `
resource "zillizcloud_byoc_i_project_settings" "test" {
    cloud_provider = "aws"
    region = "aws-us-west-2"
    project_name = "byoc-zilliz-test"

    instances = {
        core = {
            vm    = "m6i.2xlarge"
            count = 3
        }

        fundamental = {
            vm        = "m6i.2xlarge"
            min_count = 1
            max_count = 1
        }

        search = {
            vm        = "m6id.4xlarge"
            min_count = 1
            max_count = 1
        }

        index = {
            vm        = "m6i.2xlarge"
            min_count = 2
            max_count = 2
        }

        auto_scaling = true
        arch         = "X86"
    }
}

resource "zillizcloud_byoc_i_project_agent" "test" {
    project_id = zillizcloud_byoc_i_project_settings.test.project_id
    data_plane_id = zillizcloud_byoc_i_project_settings.test.data_plane_id
}

resource "zillizcloud_byoc_i_project" "test" {
    lifecycle {
        ignore_changes = [data_plane_id, project_id, aws, ext_config]
    }

    data_plane_id = zillizcloud_byoc_i_project_settings.test.data_plane_id
    project_id = zillizcloud_byoc_i_project_settings.test.project_id
   
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

    depends_on = [zillizcloud_byoc_i_project_settings.test, zillizcloud_byoc_i_project_agent.test]
}
`
}

func TestAccByocOpProjectResource_Azure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccByocOpProjectConfig_Azure(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_i_project_settings.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_i_project_settings.test", "data_plane_id"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_i_project_settings.test", "cloud_provider", "azure"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_i_project_settings.test", "region", "eastus"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_i_project_settings.test", "project_name", "byoc-zilliz-azure-test"),
					// assert status
					resource.TestCheckResourceAttr("zillizcloud_byoc_i_project.test", "status", "0"),

					resource.TestCheckResourceAttrSet("zillizcloud_byoc_i_project.test", "project_id"),
					resource.TestCheckResourceAttrSet("zillizcloud_byoc_i_project.test", "data_plane_id"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_i_project.test", "azure.region", "eastus"),
					resource.TestCheckResourceAttr("zillizcloud_byoc_i_project.test", "azure.network.vnet_id", "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.Network/virtualNetworks/my-vnet"),
				),
			},
		},
	})
}

func testAccByocOpProjectConfig_Azure() string {
	return `
resource "zillizcloud_byoc_i_project_settings" "test" {
    cloud_provider = "azure"
    region = "eastus"
    project_name = "byoc-zilliz-azure-test"

    instances = {
        core = {
            vm    = "Standard_D8s_v5"
            count = 3
        }

        fundamental = {
            vm        = "Standard_D8s_v5"
            min_count = 1
            max_count = 1
        }

        search = {
            vm        = "Standard_D16ds_v5"
            min_count = 1
            max_count = 1
        }

        index = {
            vm        = "Standard_D8s_v5"
            min_count = 2
            max_count = 2
        }

        auto_scaling = true
        arch         = "X86"
    }
}

resource "zillizcloud_byoc_i_project_agent" "test" {
    project_id = zillizcloud_byoc_i_project_settings.test.project_id
    data_plane_id = zillizcloud_byoc_i_project_settings.test.data_plane_id
}

resource "zillizcloud_byoc_i_project" "test" {
    lifecycle {
        ignore_changes = [data_plane_id, project_id, azure, ext_config]
    }

    data_plane_id = zillizcloud_byoc_i_project_settings.test.data_plane_id
    project_id = zillizcloud_byoc_i_project_settings.test.project_id

    azure = {
        region = "eastus"

        network = {
            vnet_id = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.Network/virtualNetworks/my-vnet"
            subnet_ids = [
                "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/subnet-1",
                "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/subnet-2",
                "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/subnet-3"
            ]
            nsg_ids = [
                "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.Network/networkSecurityGroups/my-nsg"
            ]
            private_endpoint_id = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.Network/privateEndpoints/my-private-endpoint"
        }

        identity = {
            storage = {
                client_id   = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
                resource_id = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/storage-identity"
            }
            kubelet = {
                client_id   = "b2c3d4e5-f6a7-8901-bcde-f12345678901"
                resource_id = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/kubelet-identity"
            }
            maintenance = {
                client_id   = "c3d4e5f6-a7b8-9012-cdef-123456789012"
                resource_id = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/maintenance-identity"
            }
        }

        storage = {
            storage_account_name = "zilliz-storage-0af21b"
            container_name       = "zilliz-container-i9fip2"
        }
    }

    depends_on = [zillizcloud_byoc_i_project_settings.test, zillizcloud_byoc_i_project_agent.test]
}
`
}
