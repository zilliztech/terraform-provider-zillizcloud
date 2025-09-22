terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

# Configure the Zilliz Cloud provider
provider "zillizcloud" {
}

# Fetch project information
data "zillizcloud_project" "default" {
  # Replace with your actual byoc project ID
  # id = "proj-example123456789"
}

# Create a cluster in byoc project
resource "zillizcloud_cluster" "example" {
  cluster_name = "example-cluster"
  project_id   = data.zillizcloud_project.default.id
  cu_size      = 1
  cu_type      = "Performance-optimized"
}

# Associate security groups with the cluster's load balancer
resource "zillizcloud_cluster_load_balancer_security_groups" "example" {
  cluster_id = zillizcloud_cluster.example.id

  security_group_ids = [
    "sg-0123456789abcdef0", # Replace with your actual security group ID
  ]
}

# Output the cluster connection details
output "cluster_endpoint" {
  description = "The cluster connection endpoint"
  value       = zillizcloud_cluster.example.connect_address
}

output "security_groups" {
  description = "The security groups associated with the cluster load balancer"
  value       = zillizcloud_cluster_load_balancer_security_groups.example.security_group_ids
}