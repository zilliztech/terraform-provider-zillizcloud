package byoc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccGCPServiceAccountDataSource(t *testing.T) {

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provider.ProviderConfig + `data "zillizcloud_gcp_service_account" "current" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.zillizcloud_gcp_service_account.current", "id"),
					resource.TestCheckResourceAttrSet("data.zillizcloud_gcp_service_account.current", "service_account"),
				),
			},
		},
	})
}
