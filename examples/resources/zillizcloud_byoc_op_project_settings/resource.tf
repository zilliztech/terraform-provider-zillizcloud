terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

resource "zillizcloud_byoc_op_project_settings" "this" {

  cloud_provider = "aws"
  region         = "aws-us-west-2"
  project_name   = "byoc-zilliz-test"

  instances = {
    core_vm                  = "m6i.2xlarge"
    core_vm_min_count        = 3
    fundamental_vm           = "m6i.2xlarge"
    fundamental_vm_min_count = 0
    search_vm                = "m6id.4xlarge"
    search_vm_min_count      = 0
  }

  private_link_enabled = true

}

