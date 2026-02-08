# Configuring AWS Client-Side Encryption for BYOC

This guide demonstrates how to configure AWS Client-Side Encryption (CSE) for your Bring Your Own Cloud (BYOC) infrastructure using Terraform. CSE can be configured at two levels:

1. **Project-level CSE** - Default encryption settings for all clusters in a BYOC project
2. **Cluster-level CSE** - Per-cluster encryption using a specific KMS key

By enabling CSE, you can encrypt your data using your own AWS KMS keys, providing enhanced security and control over your encryption keys.

## Prerequisites

Before you begin, ensure you have:

- Completed the initial setup steps outlined in the [Getting Started with Zilliz Cloud Terraform Provider](./get-start.md) guide
- An AWS KMS key created in your AWS account
- The ARN (Amazon Resource Name) of your AWS KMS key
- An IAM role with permissions to use the KMS key
- For cross-account scenarios: An external ID for secure role assumption

## Understanding AWS Client-Side Encryption

### What is Client-Side Encryption (CSE)?

Client-Side Encryption allows you to encrypt your data before it's sent to storage using your own AWS KMS (Key Management Service) keys. This provides:

- **Enhanced Security**: Data is encrypted using keys you control
- **Compliance**: Meet regulatory requirements for data encryption
- **Key Control**: Full ownership and management of encryption keys
- **Audit Trail**: Track KMS key usage through AWS CloudTrail

### CSE Configuration Levels

AWS CSE can be configured at two different levels in Zilliz Cloud BYOC:

#### 1. Project-Level CSE (Default for All Clusters)

When you configure CSE at the project level, it sets default encryption parameters for all clusters created within that project. This includes:

- **CSE Role ARN**: IAM role used for encryption/decryption operations
- **Default CSE Key ARN**: Default KMS key for encrypting data
- **External ID**: Security token for cross-account access

#### 2. Cluster-Level CSE (Per-Cluster Override)

Individual clusters can specify their own KMS key, overriding the project default. This is useful when:

- Different clusters require different encryption keys
- You need to segregate encryption keys by environment or team
- Compliance requires separate keys for different data classifications

### AWS KMS Key ARN Format

KMS key ARNs follow this format:

```
arn:aws:kms:REGION:ACCOUNT_ID:key/KEY_ID
```

Example:
```
arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012
```

### Important Constraints

- **BYOC Only**: CSE configuration is only applicable to BYOC projects and clusters
- **Immutable**: Once configured, CSE settings cannot be changed. Any attempt to modify CSE configuration after creation will require resource replacement.
- **Create-time Configuration**: CSE must be specified during resource creation
- **Optional**: If not specified, standard encryption methods will be used

## Configuration Examples

### Option 1: Project-Level CSE (Recommended for Multiple Clusters)

Configure CSE at the project level to set default encryption for all clusters:

```hcl
resource "zillizcloud_byoc_i_project" "aws_project_with_cse" {
  project_name = "production-byoc-project"
  cloud_id     = "aws"
  region       = "us-west-2"

  aws {
    # Network configuration
    network = {
      vpc_id             = "vpc-0123456789abcdef0"
      subnet_ids         = ["subnet-111", "subnet-222", "subnet-333"]
      security_group_ids = ["sg-0123456789abcdef0"]
    }

    # Storage configuration
    storage = {
      bucket_id = "my-production-bucket"
    }

    # IAM role configuration
    role_arn = {
      storage   = "arn:aws:iam::123456789012:role/ZillizStorageRole"
      eks       = "arn:aws:iam::123456789012:role/ZillizEKSRole"
      bootstrap = "arn:aws:iam::123456789012:role/ZillizBootstrapRole"
    }

    # CSE (Client-Side Encryption) configuration
    cse = {
      aws_cse_role_arn        = "arn:aws:iam::123456789012:role/ZillizCSERole"
      default_aws_cse_key_arn = "arn:aws:kms:us-west-2:123456789012:key/default-cse-key-id"
      external_id             = "unique-external-id-for-security"
    }
  }
}

# Clusters created in this project will use the default CSE configuration
resource "zillizcloud_cluster" "cluster_with_default_cse" {
  cluster_name = "production-cluster-1"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_size      = 2
  cu_type      = "Performance-optimized"
  project_id   = zillizcloud_byoc_i_project.aws_project_with_cse.id
}
```

### Option 2: Cluster-Level CSE (Per-Cluster Override)

Configure a specific KMS key for an individual cluster:

```hcl
data "zillizcloud_project" "byoc_project" {
  # Fetching your BYOC project information
}

resource "zillizcloud_cluster" "cluster_with_specific_key" {
  cluster_name    = "secure-production-cluster"
  region_id       = "aws-us-east-2"
  plan            = "Enterprise"
  cu_size         = 2
  cu_type         = "Performance-optimized"
  project_id      = data.zillizcloud_project.byoc_project.id
  aws_cse_key_arn = "arn:aws:kms:us-east-2:123456789012:key/cluster-specific-key-id"
}
```

### Option 3: Combined Configuration

Combine project-level defaults with cluster-level overrides and custom bucket:

```hcl
resource "zillizcloud_byoc_i_project" "aws_project" {
  project_name = "production-byoc-project"
  cloud_id     = "aws"
  region       = "us-west-2"

  aws {
    network = {
      vpc_id             = "vpc-0123456789abcdef0"
      subnet_ids         = ["subnet-111", "subnet-222", "subnet-333"]
      security_group_ids = ["sg-0123456789abcdef0"]
    }

    storage = {
      bucket_id = "my-production-bucket"
    }

    role_arn = {
      storage   = "arn:aws:iam::123456789012:role/ZillizStorageRole"
      eks       = "arn:aws:iam::123456789012:role/ZillizEKSRole"
      bootstrap = "arn:aws:iam::123456789012:role/ZillizBootstrapRole"
    }

    # Default CSE for all clusters
    cse = {
      aws_cse_role_arn        = "arn:aws:iam::123456789012:role/ZillizCSERole"
      default_aws_cse_key_arn = "arn:aws:kms:us-west-2:123456789012:key/default-key"
      external_id             = "secure-external-id"
    }
  }
}

# Cluster with custom KMS key and bucket
resource "zillizcloud_cluster" "high_security_cluster" {
  cluster_name    = "high-security-cluster"
  region_id       = "aws-us-west-2"
  plan            = "Enterprise"
  cu_size         = 4
  cu_type         = "Performance-optimized"
  project_id      = zillizcloud_byoc_i_project.aws_project.id

  # Override with cluster-specific KMS key
  aws_cse_key_arn = "arn:aws:kms:us-west-2:123456789012:key/high-security-key"

  # Custom storage bucket
  bucket_info = {
    bucket_name = "high-security-data-bucket"
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
