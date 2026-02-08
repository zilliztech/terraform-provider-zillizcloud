# Configuring AWS Client-Side Encryption for BYOC Clusters

This guide demonstrates how to configure AWS Client-Side Encryption (CSE) for your Bring Your Own Cloud (BYOC) Milvus clusters using Terraform. By enabling CSE, you can encrypt your cluster data using your own AWS KMS keys, providing enhanced security and control over your encryption keys.

## Prerequisites

Before you begin, ensure you have:

- Completed the initial setup steps outlined in the [Getting Started with Zilliz Cloud Terraform Provider](./get-start.md) guide
- A BYOC project already set up in Zilliz Cloud
- An AWS KMS key created in your AWS account
- The ARN (Amazon Resource Name) of your AWS KMS key
- Necessary IAM permissions to use the KMS key

## Understanding AWS Client-Side Encryption

### What is Client-Side Encryption (CSE)?

Client-Side Encryption allows you to encrypt your data before it's sent to storage using your own AWS KMS (Key Management Service) keys. This provides:

- **Enhanced Security**: Data is encrypted using keys you control
- **Compliance**: Meet regulatory requirements for data encryption
- **Key Control**: Full ownership and management of encryption keys

### AWS CSE Key Configuration

The `aws_cse_key_arn` attribute accepts the ARN of an AWS KMS key. The format is:

```
arn:aws:kms:REGION:ACCOUNT_ID:key/KEY_ID
```

Example:
```
arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012
```

### Important Constraints

- **BYOC Only**: The `aws_cse_key_arn` attribute is only applicable to BYOC clusters
- **Immutable**: Once a cluster is created with a specific KMS key, you **cannot change** the key. Any attempt to modify `aws_cse_key_arn` after cluster creation will result in an error.
- **Create-time Configuration**: The KMS key ARN must be specified during cluster creation
- **Optional**: If not specified, standard encryption methods will be used

## Creating a BYOC Cluster with AWS CSE

### Basic Configuration

```hcl
data "zillizcloud_project" "byoc_project" {
  # Fetching your BYOC project information
}

resource "zillizcloud_cluster" "byoc_cluster_with_cse" {
  cluster_name    = "secure-production-cluster"
  region_id       = "aws-us-east-2"
  plan            = "Enterprise"
  cu_size         = 2
  cu_type         = "Performance-optimized"
  project_id      = data.zillizcloud_project.byoc_project.id
  aws_cse_key_arn = "arn:aws:kms:us-east-2:123456789012:key/12345678-1234-1234-1234-123456789012"
}
```

### Configuration with Custom Bucket

You can combine AWS CSE with custom bucket configuration for complete control over your data storage and encryption:

```hcl
data "zillizcloud_project" "byoc_project" {
  # Fetching your BYOC project information
}

resource "zillizcloud_cluster" "byoc_cluster_with_cse_and_bucket" {
  cluster_name    = "secure-production-cluster"
  region_id       = "aws-us-west-2"
  plan            = "Enterprise"
  cu_size         = 4
  cu_type         = "Performance-optimized"
  project_id      = data.zillizcloud_project.byoc_project.id
  aws_cse_key_arn = "arn:aws:kms:us-west-2:987654321098:key/abcdef12-3456-7890-abcd-ef1234567890"

  bucket_info = {
    bucket_name = "my-secure-data-bucket"
    prefix      = "/encrypted-data"
  }
}
```

## AWS KMS Key Setup

### Creating a KMS Key

Before using CSE with your cluster, you need to create a KMS key in AWS:

1. Navigate to AWS KMS in the AWS Console
2. Click "Create key"
3. Choose "Symmetric" key type
4. Configure key administrative and usage permissions
5. Note the ARN of the created key

### Required IAM Permissions

Ensure your Zilliz Cloud BYOC infrastructure has permissions to use the KMS key. The key policy should allow:

- `kms:Encrypt`
- `kms:Decrypt`
- `kms:GenerateDataKey`
- `kms:DescribeKey`

## Common Errors and Troubleshooting

### Error: "Cannot change AWS CSE key ARN after cluster is created"

**Cause**: You attempted to modify the `aws_cse_key_arn` attribute after the cluster was already created.

**Solution**: The AWS CSE key configuration is immutable. If you need to use a different KMS key:

1. Destroy the existing cluster:
   ```bash
   terraform destroy -target=zillizcloud_cluster.your_cluster_name
   ```

2. Update the `aws_cse_key_arn` in your Terraform configuration

3. Recreate the cluster:
   ```bash
   terraform apply
   ```

**Note**: This will result in data loss. Ensure you have backups before destroying the cluster.

### Error: "Invalid KMS key ARN format"

**Cause**: The provided KMS key ARN is not in the correct format.

**Solution**: Verify your KMS key ARN follows this format:
```
arn:aws:kms:REGION:ACCOUNT_ID:key/KEY_ID
```

### Error: "Access denied to KMS key"

**Cause**: The BYOC infrastructure doesn't have permission to use the specified KMS key.

**Solution**:
1. Check the KMS key policy in AWS Console
2. Ensure the BYOC service role has the required permissions
3. Update the key policy to grant necessary permissions

## Best Practices

1. **Key Management**: Implement proper key rotation policies in AWS KMS
2. **Access Control**: Use least privilege principle when granting KMS key permissions
3. **Monitoring**: Enable CloudTrail logging for KMS key usage auditing
4. **Backup**: Ensure your KMS key backup and disaster recovery procedures are in place
5. **Regional Alignment**: Use a KMS key in the same AWS region as your cluster for optimal performance
6. **Documentation**: Document which KMS keys are used by which clusters for operational clarity

## Related Resources

- [Configuring Custom Bucket for BYOC Clusters](./configure-custom-bucket-for-byoc-cluster.md)
- [Creating a BYOC Project](./create-a-byoc-project.md)
- [AWS KMS Documentation](https://docs.aws.amazon.com/kms/)
