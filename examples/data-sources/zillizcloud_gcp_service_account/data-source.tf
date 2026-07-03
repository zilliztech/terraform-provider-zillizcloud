terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

data "zillizcloud_gcp_service_account" "current" {
}

output "service_account" {
  value = data.zillizcloud_gcp_service_account.current.service_account
}
