---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "zillizcloud_index Resource - zillizcloud"
subcategory: ""
description: |-
  Defines an index for a collection field.
  This resource can be attached to or detached from a collection independently.
  You can update or delete the index without affecting the collection resource itself.
---

# zillizcloud_index (Resource)

Defines an index for a collection field.
This resource can be attached to or detached from a collection independently.
You can update or delete the index without affecting the collection resource itself.

## Example Usage

```terraform
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
  properties = jsonencode({
    "database.replica.number"     = "1"
    "database.max.collections"    = "10"
    "database.force.deny.writing" = "false"
    "database.force.deny.reading" = "false"
  })
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
      }
    ]
  }
  params = {
    "mmap_enabled"      = true
    "ttl_seconds"       = 86400
    "consistency_level" = "Bounded"
  }
}

resource "zillizcloud_index" "myindex" {
  connect_address = zillizcloud_cluster.mycluster.connect_address
  db_name         = zillizcloud_database.mydb.db_name
  collection_name = zillizcloud_collection.mycollection.collection_name
  field_name      = "vector"
  metric_type     = "IP"
  index_name      = "testindex"
  index_type      = "HNSW"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `collection_name` (String) Required. The name of the collection to which the index belongs.
- `connect_address` (String) The connection address of the target Zilliz Cloud cluster.
You can obtain this value from the output of the `zillizcloud_cluster` resource, for example:
`zillizcloud_cluster.example.connect_address`

**Example:**
`https://in01-295cd02566647b7.aws-us-east-2.vectordb.zillizcloud.com:19534`

> **Note:** The address must include the protocol (e.g., `https://`).
- `db_name` (String) Required. The name of the database containing the collection.
- `field_name` (String) Required. The name of the field to be indexed.
- `index_name` (String) Required. The name of the index.
- `index_type` (String) Required. The type of the index (e.g., "IVF_FLAT", "HNSW", etc.).

### Optional

- `metric_type` (String) Optional. The metric type for the index (e.g., "L2", "IP", etc.).

### Read-Only

- `id` (String) The unique identifier for the index resource, generated by the service.
				
**Format:**`/connections/{db_name}/collections/{collection_name}/indexes/{index_name}`

**Example:**`/connections/mydb/collections/mycollection/indexes/myindex`

> **Note:** This value is automatically set and should not be manually specified.
