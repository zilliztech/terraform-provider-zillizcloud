package provider_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_user" "test" {
  connect_address = "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"
  username        = "testuser"
  password        = "LZ0lS#FRU5V49$2q"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_user.test", "connect_address", "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"),
					resource.TestCheckResourceAttr("zillizcloud_user.test", "username", "testuser"),
					resource.TestCheckResourceAttrSet("zillizcloud_user.test", "id"),
				),
			},
			{
				ResourceName:            "zillizcloud_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources["zillizcloud_user.test"]
					if !ok {
						return "", fmt.Errorf("zillizcloud_user.test not found")
					}
					connectAddress := rs.Primary.Attributes["connect_address"]
					username := rs.Primary.Attributes["username"]
					connectionID := connectAddress
					if strings.HasPrefix(connectionID, "https://") {
						connectionID = connectionID[len("https://"):]
					}
					return fmt.Sprintf("/connection/%s/user/%s", connectionID, username), nil
				},
			},
		},
	})
}
