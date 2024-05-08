terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

// default project
data "zillizcloud_project" "example01" {}

// specific project
data "zillizcloud_project" "example02" {
  name = "payments"
}

output "output_01" {
  value = data.zillizcloud_project.example01
}

output "output_02" {
  value = data.zillizcloud_project.example02
}
