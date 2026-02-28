terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}


# Basic backup policy - backup on weekdays
resource "zillizcloud_backup_policy" "basic" {
  cluster_id     = "in01-your-cluster-id"
  region_id      = "aws-us-east-2"
  frequency      = "1,2,3,4,5" # Monday through Friday
  start_time     = "02:00-04:00"
  retention_days = 7
  enabled        = true
}

# Advanced backup policy with cross-region copies
resource "zillizcloud_backup_policy" "with_cross_region" {
  cluster_id     = "in01-your-cluster-id"
  region_id      = "aws-us-east-2"
  frequency      = "1,3,5,7" # Monday, Wednesday, Friday, Sunday
  start_time     = "03:00-05:00"
  retention_days = 14
  enabled        = true
  cross_region_copies = [
    {
      region_id      = "aws-us-west-2"
      retention_days = 7
    },
    {
      region_id      = "aws-eu-west-1"
      retention_days = 5
    }
  ]
}
