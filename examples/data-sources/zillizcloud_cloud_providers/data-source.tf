terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

data "zillizcloud_cloud_providers" "example" {}

output "ouput" {
  value = data.zillizcloud_cloud_providers.example.cloud_providers
}
