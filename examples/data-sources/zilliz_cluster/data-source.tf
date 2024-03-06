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
  id = zilliz_cluster.test.id
  id = "in03-bd4013ae76e3a72"
}

output "output" {
  value = data.zilliz_cluster.test
}
