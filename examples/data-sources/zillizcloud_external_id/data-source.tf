terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

data "zillizcloud_external_id" "current" {
}

output "external_id" {
  value = data.zillizcloud_external_id.current.external_id
}