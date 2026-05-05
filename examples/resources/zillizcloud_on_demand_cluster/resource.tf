terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

# The project must have the on-demand cluster feature enabled before creating an on-demand cluster.
data "zillizcloud_project" "default" {
  id = "proj-xxxxxxxxxxxxxxxxxxxxxxx"
}

resource "zillizcloud_on_demand_cluster" "example" {
  cluster_name = "query-cluster-dev"
  project_id   = data.zillizcloud_project.default.id
  region_id    = "aws-us-west-2"
  cu_size      = 8
  auto_suspend = 1800
}
