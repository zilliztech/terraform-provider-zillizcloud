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

resource "zillizcloud_cluster" "mycluster" {
  cluster_name = "Cluster-03"                        # The name of the cluster
  region_id    = "aws-us-east-2"                     # The region where the cluster will be deployed
  plan         = "Enterprise"                        # The service plan for the cluster
  cu_size      = "1"                                 # The size of the compute unit
  cu_type      = "Performance-optimized"             # The type of compute unit, optimized for performance
  project_id   = data.zillizcloud_project.default.id # Linking to the project ID fetched earlier
}

resource "zillizcloud_database" "mydb" {
  connect_address = zillizcloud_cluster.mycluster.connect_address
  db_name         = "mydb"
  properties = {
    "database.replica.number"     = "1"
    "database.max.collections"    = "10"
    "database.force.deny.writing" = "false"
    "database.force.deny.reading" = "false"
  }
}

resource "zillizcloud_collection" "mycollection" {
  connect_address = zillizcloud_cluster.mycluster.connect_address
  db_name         = zillizcloud_database.mydb.db_name
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
      },
      {
        field_name        = "tags"
        data_type         = "Array"
        element_data_type = "VarChar"
        element_type_params = {
          max_length   = "128"
          max_capacity = "100"
        }
      }
    ]
  }
  params = {
    "mmap_enabled"      = true
    "ttl_seconds"       = 86400
    "consistency_level" = "Bounded"
  }
}
