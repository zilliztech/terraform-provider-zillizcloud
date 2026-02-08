# Configuring AWS Client-Side Encryption for BYOC

This guide demonstrates how to configure AWS Client-Side Encryption (CSE) for your Bring Your Own Cloud (BYOC) infrastructure using Terraform.

CSE requires two steps:

1. **Configure CSE infrastructure in the BYOC project** - Set up IAM roles and default KMS key that enables CSE capability
2. **Opt-in individual clusters to use CSE** - Specify `aws_cse_key_arn` on clusters that should use encryption

**Important**: Clusters do NOT use CSE by default. Each cluster must explicitly opt-in by specifying an `aws_cse_key_arn`. Different clusters within the same project can choose whether to use CSE or not.

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

### How CSE Works in BYOC

CSE configuration in Zilliz Cloud BYOC involves two components:

#### 1. Project CSE Infrastructure (Prerequisites)

First, configure the CSE infrastructure at the BYOC project level. This sets up the necessary AWS resources:

- **CSE Role ARN**: IAM role used for encryption/decryption operations
- **Default CSE Key ARN**: Default KMS key available for clusters to use
- **External ID**: Security token for secure cross-account access

**Important**: Configuring CSE at the project level does NOT automatically enable encryption for clusters. It only makes CSE capability available.

#### 2. Cluster CSE Opt-In (Per-Cluster)

Each cluster must explicitly opt-in to use CSE by specifying an `aws_cse_key_arn`. This allows flexibility:

- **Selective Encryption**: Only clusters that need encryption use it
- **Different Keys**: Different clusters can use different KMS keys for data segregation
- **Mixed Mode**: Within the same project, some clusters use CSE while others don't
- **No Default Encryption**: Clusters without `aws_cse_key_arn` will NOT use CSE

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

### Step 1: Configure CSE Infrastructure in BYOC Project

First, set up the CSE infrastructure in your BYOC project. This makes CSE capability available but doesn't automatically encrypt any clusters:

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

    # CSE infrastructure - enables CSE capability for the project
    # Note: This does NOT automatically encrypt clusters
    cse = {
      aws_cse_role_arn        = "arn:aws:iam::123456789012:role/ZillizCSERole"
      default_aws_cse_key_arn = "arn:aws:kms:us-west-2:123456789012:key/default-cse-key-id"
      external_id             = "unique-external-id-for-security"
    }
  }
}
```

### Step 2: Create Clusters with CSE Opt-In

Now you can create clusters. Each cluster independently decides whether to use CSE:

#### Example A: Cluster WITHOUT CSE (Default Behavior)

This cluster will NOT use encryption, even though the project has CSE configured:

```hcl
resource "zillizcloud_cluster" "cluster_without_cse" {
  cluster_name = "development-cluster"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_size      = 2
  cu_type      = "Performance-optimized"
  project_id   = zillizcloud_byoc_i_project.aws_project_with_cse.id

  # No aws_cse_key_arn specified = No CSE encryption
}
```

#### Example B: Cluster WITH CSE Using Default Key

This cluster opts-in to CSE and uses the default key from the project:

```hcl
resource "zillizcloud_cluster" "cluster_with_default_cse" {
  cluster_name    = "production-cluster-1"
  region_id       = "aws-us-west-2"
  plan            = "Enterprise"
  cu_size         = 2
  cu_type         = "Performance-optimized"
  project_id      = zillizcloud_byoc_i_project.aws_project_with_cse.id

  # Opt-in to CSE using the project's default key
  aws_cse_key_arn = "arn:aws:kms:us-west-2:123456789012:key/default-cse-key-id"
}
```

#### Example C: Cluster WITH CSE Using Custom Key

This cluster opts-in to CSE but uses a different KMS key for data segregation:

```hcl
resource "zillizcloud_cluster" "cluster_with_custom_cse" {
  cluster_name    = "high-security-cluster"
  region_id       = "aws-us-west-2"
  plan            = "Enterprise"
  cu_size         = 4
  cu_type         = "Performance-optimized"
  project_id      = zillizcloud_byoc_i_project.aws_project_with_cse.id

  # Opt-in to CSE using a different KMS key
  aws_cse_key_arn = "arn:aws:kms:us-west-2:123456789012:key/high-security-key-id"

  # Can also combine with custom bucket
  bucket_info = {
    bucket_name = "high-security-data-bucket"
    prefix      = "/encrypted-data"
  }
}
```

### Complete Example: Mixed Mode Project

This example shows a project with CSE infrastructure and multiple clusters with different encryption choices:

```hcl
# 1. Create BYOC project with CSE infrastructure
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

    # CSE infrastructure configuration
    cse = {
      aws_cse_role_arn        = "arn:aws:iam::123456789012:role/ZillizCSERole"
      default_aws_cse_key_arn = "arn:aws:kms:us-west-2:123456789012:key/default-key"
      external_id             = "secure-external-id"
    }
  }
}

# 2. Development cluster - NO encryption
resource "zillizcloud_cluster" "dev_cluster" {
  cluster_name = "dev-cluster"
  region_id    = "aws-us-west-2"
  plan         = "Standard"
  cu_size      = 1
  cu_type      = "Performance-optimized"
  project_id   = zillizcloud_byoc_i_project.aws_project.id
  # No aws_cse_key_arn = No encryption
}

# 3. Staging cluster - uses default CSE key
resource "zillizcloud_cluster" "staging_cluster" {
  cluster_name    = "staging-cluster"
  region_id       = "aws-us-west-2"
  plan            = "Enterprise"
  cu_size         = 2
  cu_type         = "Performance-optimized"
  project_id      = zillizcloud_byoc_i_project.aws_project.id
  aws_cse_key_arn = "arn:aws:kms:us-west-2:123456789012:key/default-key"
}

# 4. Production cluster - uses dedicated CSE key
resource "zillizcloud_cluster" "prod_cluster" {
  cluster_name    = "production-cluster"
  region_id       = "aws-us-west-2"
  plan            = "Enterprise"
  cu_size         = 8
  cu_type         = "Performance-optimized"
  project_id      = zillizcloud_byoc_i_project.aws_project.id
  aws_cse_key_arn = "arn:aws:kms:us-west-2:123456789012:key/production-key"

  bucket_info = {
    bucket_name = "production-data-bucket"
    prefix      = "/prod"
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
