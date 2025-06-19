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

resource "zillizcloud_cluster" "example" {
  cluster_name = "Cluster-03"                        # The name of the cluster
  region_id    = "aws-us-east-2"                     # The region where the cluster will be deployed
  plan         = "Enterprise"                        # The service plan for the cluster
  cu_size      = "1"                                 # The size of the compute unit
  cu_type      = "Performance-optimized"             # The type of compute unit, optimized for performance
  project_id   = data.zillizcloud_project.default.id # Linking to the project ID fetched earlier
}

data "zillizcloud_users" "example" {
  connect_address = zillizcloud_cluster.example.connect_address
}

output "users" {
  value = data.zillizcloud_users.example.items
} 