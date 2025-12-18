# Configuring a Custom Bucket for BYOC Clusters

This guide demonstrates how to configure a custom storage bucket for your Bring Your Own Cloud (BYOC) Milvus clusters using Terraform. By specifying a custom bucket, you can control where your cluster's data is stored, enabling better data governance and compliance.

## Prerequisites

Before you begin, ensure you have:

- Completed the initial setup steps outlined in the [Getting Started with Zilliz Cloud Terraform Provider](./get-start.md) guide
- A BYOC project already set up in Zilliz Cloud
- A storage bucket created in your cloud provider (AWS S3, GCP Cloud Storage, etc.)
- Necessary permissions and access credentials

## Understanding Custom Bucket Configuration

### What is bucket_info?

The `bucket_info` attribute allows you to specify which storage bucket your BYOC Milvus cluster should use for data persistence. This is particularly useful when:

- You're managing multiple clusters and want to separate their data storage

### Bucket Information Components

The `bucket_info` block contains two attributes:

- **`bucket_name`** (Required): The name of the bucket
- **`prefix`** (Optional): A prefix path within the bucket for organizing cluster data. If not provided, the cluster will use the bucket's root directory.

### Important Constraints

- **BYOC Only**: The `bucket_info` attribute is only applicable to BYOC clusters (Standard and Enterprise plans in BYOC projects)
- **Immutable**: Once a cluster is created with a specific bucket configuration, you **cannot change** the bucket information. Any attempt to modify `bucket_info` after cluster creation will result in an error.
- **Create-time Configuration**: The bucket must be specified during cluster creation

## Creating a BYOC Cluster with Custom Bucket(example)




```hcl
data "zillizcloud_project" "byoc_project" {
  # Fetching your BYOC project information
}

resource "zillizcloud_cluster" "byoc_cluster_with_custom_bucket" {
  cluster_name = "production-cluster"
  region_id    = "aws-us-east-2"
  plan         = "Enterprise"
  cu_size      = 2
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.byoc_project.id

  bucket_info = {
    bucket_name = "my-production-data-bucket"  # Required: Your bucket name
    prefix      = "/a-prefix"                  # Optional: Path prefix within bucket
  }
}
```

#

## Common Errors and Troubleshooting

### Error: "Cannot change bucket info after cluster is created"

**Cause**: You attempted to modify the `bucket_info` block after the cluster was already created.

**Solution**: The bucket configuration is immutable. If you need to use a different bucket:

1. Destroy the existing cluster:
   ```bash
   terraform destroy -target=zillizcloud_cluster.your_cluster_name
   ```

2. Update the `bucket_info` configuration in your Terraform file

3. Recreate the cluster:
   ```bash
   terraform apply
   ```