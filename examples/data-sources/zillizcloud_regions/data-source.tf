terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

data "zillizcloud_regions" "all_regions" {
}

data "zillizcloud_regions" "aws_region" {
  cloud_id = "aws"
}

data "zillizcloud_regions" "gcp_region" {
  cloud_id = "gcp"
}

data "zillizcloud_regions" "azure_region" {
  cloud_id = "azure"
}

output "all_output" {
  value = data.zillizcloud_regions.all_regions.items
}

output "aws_output" {
  value = data.zillizcloud_regions.aws_region.items
}

output "gcp_output" {
  value = data.zillizcloud_regions.gcp_region.items
}

output "azure_output" {
  value = data.zillizcloud_regions.azure_region.items
}
