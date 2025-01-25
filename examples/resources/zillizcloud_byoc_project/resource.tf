terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
  api_key   = "xxx"
  byoc_mode = true
}

resource "zillizcloud_byoc_project" "this" {
  name = "MyProject"

  aws = {
    region = "aws-us-west-2"

    network = {
      vpc_id = "vpc-0a9a08f81e27c2b69"

      subnet_ids         = ["subnet-0d352e35a68f2f7c7", "subnet-03d0d894d05e0b87b", "subnet-08cf896411a229c8e"]
      security_group_ids = ["sg-02f41431f91303644"]
      # vpc_endpoint_id    = "vpce-12345678"
    }
    role_arn = {
      storage       = "arn:aws:iam::306787409409:role/zilliz-stack-17d586-ZillizStorageRole-1jSYHHFWhGkz"
      eks           = "arn:aws:iam::306787409409:role/zilliz-stack-17d586-ZillizEKSRole-D27XZP0XK5do"
      cross_account = "arn:aws:iam::306787409409:role/zilliz-stack-17d586-ZillizBootstrapRole-DAyuQSLZEN9g"
    }
    storage = {
      bucket_id = "zilliz-s3-0af21b"
    }

    instances = {
      core_vm        = "m6i.2xlarge"
      fundamental_vm = "m6i.2xlarge"
      search_vm      = "m6id.2xlarge"
    }
  }
}