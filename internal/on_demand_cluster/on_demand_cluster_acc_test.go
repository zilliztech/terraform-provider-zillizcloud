package cluster_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/provider"
)

func TestAccOnDemandClusterResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}

	name := "tf-acc-qc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: provider.ProviderConfig + `

resource "zillizcloud_on_demand_cluster" "test" {
  cluster_name = "` + name + `"
  project_id   = "proj-7fb9f1238c4ac89b70014f"
  region_id    = "aws-us-west-2"
  cu_size      = 8
  auto_suspend = 1800
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("zillizcloud_on_demand_cluster.test", "cluster_name", name),
					resource.TestCheckResourceAttrSet("zillizcloud_on_demand_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("zillizcloud_on_demand_cluster.test", "project_id"),
					resource.TestCheckResourceAttr("zillizcloud_on_demand_cluster.test", "region_id", "aws-us-west-2"),
					resource.TestCheckResourceAttr("zillizcloud_on_demand_cluster.test", "cu_size", "8"),
					resource.TestCheckResourceAttr("zillizcloud_on_demand_cluster.test", "auto_suspend", "1800"),
					resource.TestCheckResourceAttrSet("zillizcloud_on_demand_cluster.test", "endpoint"),
					resource.TestCheckResourceAttrSet("zillizcloud_on_demand_cluster.test", "ttl_seconds"),
				),
			},
		},
	})
}
