package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccAliasResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create alias
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_database" "test" {
  connect_address = "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"
  db_name         = "testdb"
}
resource "zillizcloud_collection" "test" {
  connect_address = zillizcloud_database.test.connect_address
  db_name         = zillizcloud_database.test.db_name
  collection_name = "testcollection"
  schema = {
    auto_id = true
    enabled_dynamic_field = false
    fields = [
      {
        field_name = "id"
        data_type  = "Int64"
        is_primary = true
      },
      {
        field_name = "vector"
        data_type  = "FloatVector"
        element_type_params = {
          dim = "128"
        }
      }
    ]
  }
}
resource "zillizcloud_alias" "test" {
  connect_address = zillizcloud_database.test.connect_address
  db_name         = zillizcloud_database.test.db_name
  alias_name      = "testalias"
  collection_name = zillizcloud_collection.test.collection_name
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_alias.test", "alias_name", "testalias"),
					resource.TestCheckResourceAttr("zillizcloud_alias.test", "collection_name", "testcollection"),
					resource.TestCheckResourceAttrSet("zillizcloud_alias.test", "id"),
				),
			},
			// Step 2: Update alias_name (should update in place)
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_database" "test" {
  connect_address = "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"
  db_name         = "testdb"
}
resource "zillizcloud_collection" "test" {
  connect_address = zillizcloud_database.test.connect_address
  db_name         = zillizcloud_database.test.db_name
  collection_name = "testcollection"
  schema = {
    auto_id = true
    enabled_dynamic_field = false
    fields = [
      {
        field_name = "id"
        data_type  = "Int64"
        is_primary = true
      },
      {
        field_name = "vector"
        data_type  = "FloatVector"
        element_type_params = {
          dim = "128"
        }
      }
    ]
  }
}
resource "zillizcloud_alias" "test" {
  connect_address = zillizcloud_database.test.connect_address
  db_name         = zillizcloud_database.test.db_name
  alias_name      = "testalias2"
  collection_name = zillizcloud_collection.test.collection_name
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_alias.test", "alias_name", "testalias2"),
					resource.TestCheckResourceAttr("zillizcloud_alias.test", "collection_name", "testcollection"),
					resource.TestCheckResourceAttrSet("zillizcloud_alias.test", "id"),
				),
			},
			// Step 3: Import alias
			{
				ResourceName:            "zillizcloud_alias.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"collection_name"},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources["zillizcloud_alias.test"]
					if !ok {
						return "", fmt.Errorf("zillizcloud_alias.test not found")
					}
					connectAddress := rs.Primary.Attributes["connect_address"]
					dbName := rs.Primary.Attributes["db_name"]
					aliasName := rs.Primary.Attributes["alias_name"]
					connectAddress = connectAddress[len("https://"):]
					return fmt.Sprintf("/connections/%s/databases/%s/aliases/%s", connectAddress, dbName, aliasName), nil
				},
			},
		},
	})
}
