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

data "zilliz_cloud_providers" "example" {}

data "zilliz_cloud_regions" "example" {
  cloud_id = data.zilliz_cloud_providers.example.cloud_providers[0].cloud_id
}
