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

data "zilliz_cloud_regions" "region01" {
  cloud_id = "aws"
}

data "zilliz_cloud_regions" "region02" {
  cloud_id = "gcp"
}

data "zilliz_cloud_regions" "region03" {
  cloud_id = "azure"
}

output "aws_ouput" {
  value = data.zilliz_cloud_regions.region01.cloud_regions
}


output "gcp_ouput" {
  value = data.zilliz_cloud_regions.region02.cloud_regions
}

output "azure_ouput" {
  value = data.zilliz_cloud_regions.region03.cloud_regions
}
