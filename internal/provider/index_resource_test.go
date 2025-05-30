package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccIndexResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create database, collection, and index
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
resource "zillizcloud_index" "test" {
  connect_address  = zillizcloud_collection.test.connect_address
  db_name          = zillizcloud_collection.test.db_name
  collection_name  = zillizcloud_collection.test.collection_name
  field_name       = "vector"
  metric_type      = "L2"
  index_name       = "testindex"
  index_type       = "IVF_FLAT"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_index.test", "connect_address", "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"),
					resource.TestCheckResourceAttr("zillizcloud_index.test", "db_name", "testdb"),
					resource.TestCheckResourceAttr("zillizcloud_index.test", "collection_name", "testcollection"),
					resource.TestCheckResourceAttr("zillizcloud_index.test", "field_name", "vector"),
					resource.TestCheckResourceAttr("zillizcloud_index.test", "metric_type", "L2"),
					resource.TestCheckResourceAttr("zillizcloud_index.test", "index_name", "testindex"),
					resource.TestCheckResourceAttr("zillizcloud_index.test", "index_type", "IVF_FLAT"),
					resource.TestCheckResourceAttrSet("zillizcloud_index.test", "id"),
				),
			},
			// Step 2: Update index (change metric_type and index_type)
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
resource "zillizcloud_index" "test" {
  connect_address  = zillizcloud_collection.test.connect_address
  db_name          = zillizcloud_collection.test.db_name
  collection_name  = zillizcloud_collection.test.collection_name
  field_name       = "vector"
  metric_type      = "IP"
  index_name       = "testindex"
  index_type       = "HNSW"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_index.test", "metric_type", "IP"),
					resource.TestCheckResourceAttr("zillizcloud_index.test", "index_type", "HNSW"),
				),
			},
			// Step 3: Import index
			{
				ResourceName:            "zillizcloud_index.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"field_name", "metric_type", "index_type"},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources["zillizcloud_index.test"]
					if !ok {
						return "", fmt.Errorf("zillizcloud_index.test not found")
					}
					connectAddress := rs.Primary.Attributes["connect_address"]
					dbName := rs.Primary.Attributes["db_name"]
					collectionName := rs.Primary.Attributes["collection_name"]
					indexName := rs.Primary.Attributes["index_name"]
					connectAddress = connectAddress[len("https://"):]
					return fmt.Sprintf("/connections/%s/databases/%s/collections/%s/indexes/%s", connectAddress, dbName, collectionName, indexName), nil
				},
			},
		},
	})
}
