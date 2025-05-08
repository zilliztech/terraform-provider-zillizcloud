terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}


data "zillizcloud_byoc_i_project_settings" "test" {
  project_id    = "<project_id from zillizcloud console>"
  data_plane_id = "<data_plane_id from zillizcloud console>"
}

output "agent_config" {
  value = data.zillizcloud_byoc_i_project_settings.this.op_config
}

output "node_groups" {
  value = data.zillizcloud_byoc_i_project_settings.this.node_quotas
}
