terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

data "zillizcloud_project" "default" {
}

resource "zillizcloud_cluster" "starter_cluster" {
  cluster_name = "Cluster-01"
  project_id   = data.zillizcloud_project.default.id
}

resource "zillizcloud_cluster" "standard_plan_cluster" {
  cluster_name = "Cluster-02"
  region_id    = "aws-us-east-2"
  plan         = "Standard"
  cu_size      = "1"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
}

