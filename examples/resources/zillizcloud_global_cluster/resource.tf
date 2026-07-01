terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

resource "zillizcloud_global_cluster" "example" {
  global_cluster_name = "example-global-cluster"
  project_id          = "proj-ebc5ac7f430702aec8c57b"
  cu_type             = "Performance-optimized"
  cu_size             = 1

  cluster = [
    {
      cluster_name = "example-primary"
      region_id    = "aws-us-west-2"
    },
    {
      cluster_name = "example-secondary-eu"
      region_id    = "aws-eu-west-1"
    }
  ]
}
