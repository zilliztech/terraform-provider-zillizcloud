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

// default project
data "zilliz_projects" "example01" {}

// specific project
data "zilliz_projects" "example02" {
  name = "payments"
}

output "output_01" {
  value = data.zilliz_projects.example01.projects
}

output "output_02" {
  value = data.zilliz_projects.example02.projects
}
