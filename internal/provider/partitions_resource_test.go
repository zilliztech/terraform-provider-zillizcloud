package provider_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccPartitionsResource(t *testing.T) {
	timestamp := time.Now().Unix()
	dbName := fmt.Sprintf("testdb_%d", timestamp)
	collectionName := fmt.Sprintf("testcollection_%d", timestamp)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create partition
			{
				Config: provider.ProviderConfig + fmt.Sprintf(`
resource "zillizcloud_database" "test" {
  connect_address = "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"
  db_name         = "%s"
}
resource "zillizcloud_collection" "test" {
  connect_address = zillizcloud_database.test.connect_address
  db_name         = zillizcloud_database.test.db_name
  collection_name = "%s"
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
  params = {
    consistency_level = "Bounded"
  }
}
resource "zillizcloud_partitions" "test" {
  connect_address = zillizcloud_database.test.connect_address
  db_name         = zillizcloud_database.test.db_name
  collection_name = zillizcloud_collection.test.collection_name
  partition_name  = "testpartition"
}
`, dbName, collectionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_partitions.test", "partition_name", "testpartition"),
					resource.TestCheckResourceAttr("zillizcloud_partitions.test", "collection_name", collectionName),
					resource.TestCheckResourceAttrSet("zillizcloud_partitions.test", "id"),
				),
			},
			// Step 2: Update partition_name (should update in place)
			{
				Config: provider.ProviderConfig + fmt.Sprintf(`
resource "zillizcloud_database" "test" {
  connect_address = "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"
  db_name         = "%s"
}
resource "zillizcloud_collection" "test" {
  connect_address = zillizcloud_database.test.connect_address
  db_name         = zillizcloud_database.test.db_name
  collection_name = "%s"
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
  params = {
    consistency_level = "Bounded"
  }
}
resource "zillizcloud_partitions" "test" {
  connect_address = zillizcloud_database.test.connect_address
  db_name         = zillizcloud_database.test.db_name
  collection_name = zillizcloud_collection.test.collection_name
  partition_name  = "testpartition2"
}
`, dbName, collectionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_partitions.test", "partition_name", "testpartition2"),
					resource.TestCheckResourceAttr("zillizcloud_partitions.test", "collection_name", collectionName),
					resource.TestCheckResourceAttrSet("zillizcloud_partitions.test", "id"),
				),
			},
			// Step 3: Import partition
			{
				ResourceName:      "zillizcloud_partitions.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources["zillizcloud_partitions.test"]
					if !ok {
						return "", fmt.Errorf("zillizcloud_partitions.test not found")
					}
					connectAddress := rs.Primary.Attributes["connect_address"]
					dbName := rs.Primary.Attributes["db_name"]
					collectionName := rs.Primary.Attributes["collection_name"]
					partitionName := rs.Primary.Attributes["partition_name"]
					connectAddress = connectAddress[len("https://"):]
					return fmt.Sprintf("/connections/%s/databases/%s/collections/%s/partitions/%s", connectAddress, dbName, collectionName, partitionName), nil
				},
			},
		},
	})
}
