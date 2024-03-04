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

data "zilliz_projects" "test" {
}

resource "zilliz_cluster" "test" {
  plan         = "Standard"
  cluster_name = "Cluster-Test1"
  cu_size      = 1
  cu_type      = "Performance-optimized"
  project_id   = data.zilliz_projects.test.projects[0].project_id
}

data "zilliz_cluster" "test" {
  id = zilliz_cluster.test.id
}

