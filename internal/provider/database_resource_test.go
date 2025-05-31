package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccDatabaseResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_database" "test" {
  connect_address = "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"
  db_name         = "testdb"
  properties      = {
    "database.replica.number" = "1"
    "database.max.collections" = "10"
    "database.force.deny.writing" = "false"
    "database.force.deny.reading" = "false"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_database.test", "connect_address", "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"),
					resource.TestCheckResourceAttr("zillizcloud_database.test", "db_name", "testdb"),
					resource.TestCheckResourceAttrSet("zillizcloud_database.test", "id"),
				),
			},
			{
				Config: provider.ProviderConfig + `
resource "zillizcloud_database" "test" {
  connect_address = "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"
  db_name         = "testdb1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_database.test", "connect_address", "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"),
					resource.TestCheckResourceAttr("zillizcloud_database.test", "db_name", "testdb1"),
					resource.TestCheckResourceAttrSet("zillizcloud_database.test", "id"),
				),
			},
			// Test update
			// TODO: permission bug not work
			/*
							{
								Config: provider.ProviderConfig + `
				resource "zillizcloud_database" "test" {
				  connect_address = "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"
				  db_name         = "testdb"
				  properties      = jsonencode({
				    "database.replica.number" = 2
				    "database.max.collections" = 20
				    "database.force.deny.writing" = true
				    "database.force.deny.reading" = true
				  })
				}
				`,
								Check: resource.ComposeAggregateTestCheckFunc(
									resource.TestCheckResourceAttr("zillizcloud_database.test", "db_name", "testdb"),
									resource.TestCheckResourceAttrSet("zillizcloud_database.test", "properties"),
									resource.TestCheckFunc(func(s *terraform.State) error {
										rs, ok := s.RootModule().Resources["zillizcloud_database.test"]
										if !ok {
											return fmt.Errorf("zillizcloud_database.test not found")
										}
										val := rs.Primary.Attributes["properties"]
										var m map[string]any
										if err := json.Unmarshal([]byte(val), &m); err != nil {
											return err
										}
										if m["database.replica.number"] != float64(2) {
											return fmt.Errorf("expected database.replica.number=2, got %v", m["database.replica.number"])
										}
										return nil
									}),
								),
							},
			*/
			// Test import
			{
				ResourceName:            "zillizcloud_database.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"properties"},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources["zillizcloud_database.test"]
					if !ok {
						return "", fmt.Errorf("zillizcloud_database.test not found")
					}
					connectAddress := rs.Primary.Attributes["connect_address"]
					dbName := rs.Primary.Attributes["db_name"]
					connectAddress = connectAddress[len("https://"):]
					return fmt.Sprintf("/connections/%s/databases/%s", connectAddress, dbName), nil
				},
			},
		},
	})
}
