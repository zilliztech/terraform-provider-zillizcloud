terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}


resource "zillizcloud_byoc_op_project_agent" "this" {
  project_id    = "<project_id from zillizcloud_byoc_op_project_settings>"
  data_plane_id = "<data_plane_id from zillizcloud_byoc_op_project_settings>"
}
