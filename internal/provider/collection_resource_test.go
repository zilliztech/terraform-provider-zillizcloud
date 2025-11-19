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
          dim = "256"
        }
      },
      {
        field_name = "tags"
        data_type  = "Array"
        element_data_type = "VarChar"
        element_type_params = {
          max_length   = "128"
          max_capacity = "100"
        }
      }
    ]
  }
  params = {
    mmap_enabled = true
    ttl_seconds = 86400
    consistency_level = "Bounded"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_collection.test", "connect_address", "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"),
					resource.TestCheckResourceAttr("zillizcloud_collection.test", "db_name", "testdb"),
					resource.TestCheckResourceAttr("zillizcloud_collection.test", "collection_name", "testcollection"),
					resource.TestCheckResourceAttr("zillizcloud_collection.test", "schema.fields.2.field_name", "tags"),
					resource.TestCheckResourceAttr("zillizcloud_collection.test", "schema.fields.2.data_type", "Array"),
					resource.TestCheckResourceAttr("zillizcloud_collection.test", "schema.fields.2.element_data_type", "VarChar"),
					resource.TestCheckResourceAttr("zillizcloud_collection.test", "schema.fields.2.element_type_params.max_length", "128"),
					resource.TestCheckResourceAttrSet("zillizcloud_collection.test", "id"),
				),
			},
			// Step 2: Update collection (change params only)
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
          dim = "256"
        }
      },
      {
        field_name = "tags"
        data_type  = "Array"
        element_data_type = "VarChar"
        element_type_params = {
          max_length   = "128"
          max_capacity = "100"
        }
      }
    ]
  }
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				//			Check: resource.ComposeAggregateTestCheckFunc(
				//				resource.TestCheckResourceAttr("zillizcloud_collection.test", "schema.auto_id", "false"),
				//			),
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
