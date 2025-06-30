terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

resource "zillizcloud_byoc_project" "this" {
  name   = "MyProject"
  status = "RUNNING"

  aws = {
    region = "aws-us-west-2"

    network = {
      vpc_id = "vpc-0a9a08f81e27c2b69"

      subnet_ids         = ["subnet-0d352e35a68f2f7c7", "subnet-03d0d894d05e0b87b", "subnet-08cf896411a229c8e"]
      security_group_ids = ["sg-02f41431f91303644"]
      vpc_endpoint_id    = "vpce-12345678"
    }
    role_arn = {
      storage       = "arn:aws:iam::999999999999:role/zilliz-stack-17d586-ZillizStorageRole-1jSYHHFWhGkz"
      eks           = "arn:aws:iam::999999999999:role/zilliz-stack-17d586-ZillizEKSRole-D27XZP0XK5do"
      cross_account = "arn:aws:iam::999999999999:role/zilliz-stack-17d586-ZillizBootstrapRole-DAyuQSLZEN9g"
    }
    storage = {
      bucket_id = "zilliz-s3-0af21b"
    }
  }

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

}