terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
  region_id = "gcp-us-west1"
}


data "zillizcloud_cluster" "test" {
  id = "in03-bd4013ae76exxx"
}

output "output" {
  value = data.zillizcloud_cluster.test
}
