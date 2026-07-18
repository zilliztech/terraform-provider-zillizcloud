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

  cu_settings = {
    dynamic_scaling = {
      min = 4
      max = 8
    }
  }

  replica_settings = {
    dynamic_scaling = {
      min = 1
      max = 3
    }
  }

  cluster = [
    {
      cluster_name = "example-primary"
      region_id    = "aws-us-west-2"
      replica      = 2
    },
    {
      cluster_name = "example-secondary-eu"
      region_id    = "aws-eu-west-1"
      replica      = 1
    }
  ]
}
