package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccDatabasesDataSource(t *testing.T) {
	const connectAddress = "https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534"
	const dbName = "testdbds"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provider.ProviderConfig + fmt.Sprintf(`
resource "zillizcloud_database" "test" {
  connect_address = "%s"
  db_name         = "%s"
  properties      = jsonencode({
    "database.replica.number" = 1
    "database.max.collections" = 10
    "database.force.deny.writing" = false
    "database.force.deny.reading" = false
  })
}
data "zillizcloud_databases" "test" {
  connect_address = zillizcloud_database.test.connect_address
}
`, connectAddress, dbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.zillizcloud_databases.test", "items.0.db_name",
						regexp.MustCompile(dbName),
					),
				),
			},
		},
	})
}
