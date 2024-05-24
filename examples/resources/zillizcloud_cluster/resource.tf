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

resource "zillizcloud_cluster" "free_plan_cluster" {
  cluster_name = "Cluster-01"                        # The name of the cluster
  plan         = "Free"                              # The service plan for the cluster The name of the cluster
  project_id   = data.zillizcloud_project.default.id # Linking to the project ID fetched earlier
}

resource "zillizcloud_cluster" "serverless_plan_cluster" {
  cluster_name = "Cluster-02"                        # The name of the cluster
  plan         = "Serverless"                        # The service plan for the cluster# The name of the cluster
  project_id   = data.zillizcloud_project.default.id # Linking to the project ID fetched earlier
}


resource "zillizcloud_cluster" "standard_plan_cluster" {
  cluster_name = "Cluster-03"                        # The name of the cluster
  region_id    = "aws-us-east-2"                     # The region where the cluster will be deployed
  plan         = "Standard"                          # The service plan for the cluster
  cu_size      = "1"                                 # The size of the compute unit
  cu_type      = "Performance-optimized"             # The type of compute unit, optimized for performance
  project_id   = data.zillizcloud_project.default.id # Linking to the project ID fetched earlier
}
