terraform {
  required_providers {
    zilliz = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zilliz" {
  cloud_region_id = "gcp-us-west1"
}

data "zilliz_clusters" "example" {}

output "clusters" {
  value = data.zilliz_clusters.example.clusters
}

