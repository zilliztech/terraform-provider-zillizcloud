package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccCollectionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create collection
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_database" "test" {
  connect_address = "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"
  db_name         = "testdb"
  properties      = jsonencode({
    "database.replica.number" = 1
    "database.max.collections" = 10
    "database.force.deny.writing" = false
    "database.force.deny.reading" = false
  })
}
resource "zillizcloud_collection" "test" {
  connect_address = "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"
  db_name         = "testdb"
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
  params = jsonencode({
    "mmap.enabled" = true
    "ttlSeconds" = 86400
    "consistencyLevel" = "Bounded"
  })
  depends_on = [zillizcloud_database.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_collection.test", "connect_address", "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"),
					resource.TestCheckResourceAttr("zillizcloud_collection.test", "db_name", "testdb"),
					resource.TestCheckResourceAttr("zillizcloud_collection.test", "collection_name", "testcollection"),
					resource.TestCheckResourceAttrSet("zillizcloud_collection.test", "id"),
				),
			},
			// Step 2: Update collection (change schema)
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_database" "test" {
  connect_address = "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"
  db_name         = "testdb"
  properties      = jsonencode({
    "database.replica.number" = 1
    "database.max.collections" = 10
    "database.force.deny.writing" = false
    "database.force.deny.reading" = false
  })
}
resource "zillizcloud_collection" "test" {
  connect_address = "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"
  db_name         = "testdb"
  collection_name = "testcollection"
  schema = {
    auto_id = false
    enabled_dynamic_field = true
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
          dim = "256"
        }
      },
      {
        field_name = "tag"
        data_type  = "VarChar"
        element_type_params = {
          max_length = "256"
        }
      }
    ]
  }
  params = jsonencode({
    "mmap.enabled" = true
    "ttlSeconds" = 86400
    "consistencyLevel" = "Bounded"
  })
  depends_on = [zillizcloud_database.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_collection.test", "db_name", "testdb"),
					resource.TestCheckResourceAttr("zillizcloud_collection.test", "collection_name", "testcollection"),
					resource.TestCheckResourceAttrSet("zillizcloud_collection.test", "id"),
				),
			},
			// Step 3: Import collection
			{
				ResourceName:            "zillizcloud_collection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"schema"},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources["zillizcloud_collection.test"]
					if !ok {
						return "", fmt.Errorf("zillizcloud_collection.test not found")
					}
					connectAddress := rs.Primary.Attributes["connect_address"]
					dbName := rs.Primary.Attributes["db_name"]
					collectionName := rs.Primary.Attributes["collection_name"]
					connectAddress = connectAddress[len("https://"):]
					return fmt.Sprintf("/connections/%s/databases/%s/collections/%s", connectAddress, dbName, collectionName), nil
				},
			},
		},
	})
}
