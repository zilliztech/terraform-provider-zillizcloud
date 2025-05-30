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

resource "zillizcloud_database" "example" {
  connect_address = zillizcloud_cluster.example.connect_address
  db_name         = "mydb"
  properties = {
    "database.replica.number"     = "1"
    "database.max.collections"    = "10"
    "database.force.deny.writing" = "false"
    "database.force.deny.reading" = "false"
  }
}

resource "zillizcloud_collection" "example" {
  connect_address = zillizcloud_cluster.example.connect_address
  db_name         = zillizcloud_database.example.db_name
  collection_name = "mycollection"
  schema = {
    auto_id               = true
    enabled_dynamic_field = false
    fields = [
      {
        field_name = "id"
        data_type  = "Int64"
        is_primary = true
      },
      {
        field_name = "vector"
        data_type  = "FloatVector"
        element_type_params = {
          dim = "128"
        }
      }
    ]
  }
  params = {
    consistency_level = "Bounded"
  }
}

# List all aliases in the database
data "zillizcloud_aliases" "all_aliases" {
  connect_address = zillizcloud_cluster.example.connect_address
  db_name         = zillizcloud_database.example.db_name
  # collection_name is omitted to get all aliases in the database
}

# List aliases for a specific collection
data "zillizcloud_aliases" "collection_aliases" {
  connect_address = zillizcloud_cluster.example.connect_address
  db_name         = zillizcloud_database.example.db_name
  collection_name = zillizcloud_collection.example.collection_name
}

output "all_aliases" {
  description = "All aliases in the database"
  value       = data.zillizcloud_aliases.all_aliases.items
}

output "collection_aliases" {
  description = "Aliases for the specific collection"
  value       = data.zillizcloud_aliases.collection_aliases.items
} 