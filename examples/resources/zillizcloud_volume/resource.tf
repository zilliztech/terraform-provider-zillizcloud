terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

resource "zillizcloud_volume" "managed" {
  project_id  = "proj-example"
  region_id   = "aws-us-west-2"
  volume_name = "terraform-managed-volume"
}

resource "zillizcloud_volume" "external" {
  project_id             = "proj-example"
  region_id              = "aws-us-west-2"
  volume_name            = "terraform-external-volume"
  type                   = "EXTERNAL"
  storage_integration_id = "storage-integration-example"
  path                   = "terraform/external-volume/"
}
