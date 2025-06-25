terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

resource "zillizcloud_byoc_i_project_settings" "this" {

  cloud_provider = "aws"
  region         = "aws-us-west-2"
  project_name   = "byoc-zilliz-test"

  instances = {
    core = {
      vm    = "m6i.2xlarge"
      count = 3
    }

    fundamental = {
      vm        = "m6i.2xlarge"
      min_count = 1
      max_count = 1
    }

    search = {
      vm        = "m6id.4xlarge"
      min_count = 1
      max_count = 1
    }

    index = {
      vm        = "m6i.2xlarge"
      min_count = 2
      max_count = 2
    }

    auto_scaling = true
    arch         = "X86"
  }

  private_link_enabled = true

}

