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

data "zilliz_cloud_regions" "example" {
  cloud_id = "gcp"
}

output "ouput" {
  value = data.zilliz_cloud_regions.example.cloud_regions
}
