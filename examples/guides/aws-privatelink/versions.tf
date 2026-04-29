terraform {
  required_version = ">= 1.3"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.100.0"
    }
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

provider "zillizcloud" {
  api_key = var.zilliz_api_key
}
