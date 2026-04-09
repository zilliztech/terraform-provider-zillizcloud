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
  cu_size      = 1                                   # The size of the compute unit
  cu_type      = "Performance-optimized"             # The type of compute unit, optimized for performance
  project_id   = data.zillizcloud_project.default.id # Linking to the project ID fetched earlier
}


resource "zillizcloud_cluster" "enterprise_plan_cluster" {
  cluster_name = "Cluster-04"                        # The name of the cluster
  region_id    = "aws-us-east-2"                     # The region where the cluster will be deployed
  plan         = "Enterprise"                        # The service plan for the cluster
  cu_size      = 1                                   # The size of the compute unit
  cu_type      = "Performance-optimized"             # The type of compute unit, optimized for performance
  project_id   = data.zillizcloud_project.default.id # Linking to the project ID fetched earlier
}

# Create a cluster with CU autoscaling enabled from the start
resource "zillizcloud_cluster" "autoscaling_cluster" {
  cluster_name = "Cluster-05"
  region_id    = "aws-us-east-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  cu_settings = {
    dynamic_scaling = {
      min = 2
      max = 8
    }
  }
}

# Create a cluster with both CU and replica autoscaling
resource "zillizcloud_cluster" "full_autoscaling_cluster" {
  cluster_name = "Cluster-06"
  region_id    = "aws-us-east-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  cu_settings = {
    dynamic_scaling = {
      min = 2
      max = 8
    }
  }
  replica_settings = {
    dynamic_scaling = {
      min = 1
      max = 3
    }
  }
}
