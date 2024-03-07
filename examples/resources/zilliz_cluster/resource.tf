terraform {
  required_providers {
    zilliz = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zilliz" {
  api_key         = "fake-api-key"
  cloud_region_id = "gcp-us-west1"
}

data "zilliz_project" "default" {
}

resource "zilliz_cluster" "standard_plan_cluster" {
  plan         = "Standard"
  cluster_name = "Cluster-01"
  cu_size      = "1"
  cu_type      = "Performance-optimized"
  project_id   = data.zilliz_project.default.id
}

resource "zilliz_cluster" "serverless_cluster" {
  cluster_name = "Cluster-02"
  project_id   = data.zilliz_project.default.id
}
