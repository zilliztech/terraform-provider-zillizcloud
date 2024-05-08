terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

data "zillizcloud_clusters" "example" {}

output "clusters" {
  value = data.zillizcloud_clusters.example.clusters
}
