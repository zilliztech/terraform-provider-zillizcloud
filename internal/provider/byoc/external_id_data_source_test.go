package byoc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccExternalIdDataSource(t *testing.T) {

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provider.ProviderConfig + `data "zillizcloud_external_id" "current" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.zillizcloud_external_id.current", "external_id", "cid-c88368a7164f15ad9e1fa9068"),
				),
			},
		},
	})
}
