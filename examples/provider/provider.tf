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
