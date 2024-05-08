terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
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

output "aws_ouput" {
  value = data.zillizcloud_regions.aws_region.items
}


output "gcp_ouput" {
  value = data.zillizcloud_regions.gcp_region.items
}

output "azure_ouput" {
  value = data.zillizcloud_regions.azure_region.items
}
