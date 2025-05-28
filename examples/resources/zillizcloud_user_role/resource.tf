terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

data "zillizcloud_project" "default" {
  # Fetching the default project information to be used in cluster provisioning
}

resource "zillizcloud_cluster" "cluster" {
  cluster_name = "Cluster-03"                        # The name of the cluster
  region_id    = "aws-us-east-2"                     # The region where the cluster will be deployed
  plan         = "Enterprise"                        # The service plan for the cluster
  cu_size      = "1"                                 # The size of the compute unit
  cu_type      = "Performance-optimized"             # The type of compute unit, optimized for performance
  project_id   = data.zillizcloud_project.default.id # Linking to the project ID fetched earlier
}

resource "random_password" "password" {
  length           = 16
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

resource "zillizcloud_user" "myuser" {
  connect_address = zillizcloud_cluster.cluster.connect_address
  username        = "test111"
  password        = random_password.password.result
}

resource "zillizcloud_user_role" "myuser_role" {
  connect_address = zillizcloud_cluster.standard_plan_cluster.connect_address
  username        = zillizcloud_user.myuser.username
  roles           = ["db_rw"]
}
