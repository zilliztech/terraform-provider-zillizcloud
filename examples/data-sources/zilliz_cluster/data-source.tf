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


data "zilliz_cluster" "test" {
  id = "in03-bd4013ae76exxx"
}

output "output" {
  value = data.zilliz_cluster.test
}
