package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccUserRoleResource(t *testing.T) {
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

resource "zillizcloud_user_role" "test" {
  connect_address = zillizcloud_user.test.connect_address
  username        = zillizcloud_user.test.username
  roles           = ["db_ro","db_rw"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_user_role.test", "connect_address", "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"),
					resource.TestCheckResourceAttr("zillizcloud_user_role.test", "username", "testuser"),
					resource.TestCheckResourceAttr("zillizcloud_user_role.test", "roles.0", "db_ro"),
					resource.TestCheckResourceAttr("zillizcloud_user_role.test", "roles.1", "db_rw"),
					resource.TestCheckResourceAttrSet("zillizcloud_user_role.test", "id"),
				),
			},
			{
				ResourceName:      "zillizcloud_user_role.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources["zillizcloud_user_role.test"]
					if !ok {
						return "", fmt.Errorf("zillizcloud_user_role.test not found")
					}
					connectAddress := rs.Primary.Attributes["connect_address"]
					connectAddress = connectAddress[len("https://"):]
					username := rs.Primary.Attributes["username"]
					return fmt.Sprintf("/connections/%s/users/%s/roles", connectAddress, username), nil
				},
			},
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_user" "test" {
  connect_address = "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"
  username        = "testuser"
  password        = "LZ0lS#FRU5V49$2q"
}

resource "zillizcloud_user_role" "test" {
  connect_address = zillizcloud_user.test.connect_address
  username        = zillizcloud_user.test.username
  roles           = ["db_ro"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_user_role.test", "connect_address", "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"),
					resource.TestCheckResourceAttr("zillizcloud_user_role.test", "username", "testuser"),
					resource.TestCheckResourceAttr("zillizcloud_user_role.test", "roles.0", "db_ro"),
					resource.TestCheckResourceAttr("zillizcloud_user_role.test", "roles.#", "1"),
					resource.TestCheckResourceAttrSet("zillizcloud_user_role.test", "id"),
				),
			},
		},
	})
}
