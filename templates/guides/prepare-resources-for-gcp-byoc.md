# Prepare Resources for GCP BYOC Project Setup Guide

## 1. Introduction

Zilliz Cloud Bring Your Own Cloud (BYOC) allows you to deploy Zilliz's vector database services within your own Google Cloud Platform (GCP) environment. This approach provides greater control over your infrastructure, data residency, and security while still benefiting from Zilliz's managed service capabilities.

This guide walks you through the process of using Terraform to provision the necessary GCP infrastructure for a Zilliz BYOC deployment. Terraform automates the creation of all required resources, ensuring consistency and reproducibility while significantly reducing the manual effort involved in setting up your environment.

By following this guide, you will:
- Set up the required GCP infrastructure components
- Configure networking, storage, and service accounts
- Prepare your environment for Zilliz BYOC deployment
- Learn how to customize the configuration to meet your specific requirements

## 2. Prerequisites

Before you begin, ensure you have the following prerequisites in place:

### GCP Account and Project

- An active GCP account with billing enabled
- A GCP project with sufficient permissions to create resources
- The project ID of your GCP project

### Software Requirements

- **Terraform**: Version 1.0.0 or later installed on your local machine
  - Installation guide: [Terraform Installation](https://learn.hashicorp.com/tutorials/terraform/install-cli)
- **Google Cloud SDK**: Installed and configured
  - Installation guide: [Google Cloud SDK Installation](https://cloud.google.com/sdk/docs/install)

### GCP API Enablement

The following GCP APIs must be enabled in your project:
- Compute Engine API (`compute.googleapis.com`)
- Kubernetes Engine API (`container.googleapis.com`)
- Identity and Access Management (IAM) API (`iam.googleapis.com`)
- Cloud Resource Manager API (`cloudresourcemanager.googleapis.com`)
- Cloud Storage API (`storage.googleapis.com`)

You can enable these APIs using the following gcloud command:

```bash
gcloud services enable compute.googleapis.com container.googleapis.com iam.googleapis.com cloudresourcemanager.googleapis.com storage.googleapis.com
```

### Authentication Setup

Authenticate with GCP using one of the following methods:

1. **Service Account Key (recommended for automation)**:
   - Create a service account with the required permissions
   - Generate and download a JSON key file
   - Set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable:
     ```bash
     export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your-service-account-key.json"
     ```

2. **User Account (recommended for local development)**:
   - Run `gcloud auth application-default login` and follow the prompts

## 3. Repository Structure

The `terraform-zilliz-examples` repository contains various examples for deploying Zilliz BYOC on different cloud providers. For GCP deployment, we'll focus on the `gcp-project-byoc-manual` directory.

### Key Files

- **main.tf**: Defines the main infrastructure components including VPC, GCS, IAM, and GKE resources
- **provider.tf**: Configures the Google Cloud provider for Terraform
- **variables.tf**: Declares all variables used in the configuration
- **terraform.tfvars.json**: Default input values for required variables
- **terraform.examples.tfvars.json**: Example of a more comprehensive configuration with custom values

### Module Structure

The configuration uses several modules to organize resources:
- **VPC Module**: Creates the VPC network and subnets
- **GCS Module**: Sets up Google Cloud Storage buckets
- **Private Link Module**: Configures private connectivity
- **IAM Module**: Sets up service accounts and permissions

## 4. Configuration Setup

### Cloning the Repository

First, clone the Terraform examples repository to your local machine:

```bash
git clone https://github.com/zilliztech/terraform-zilliz-examples.git
cd terraform-zilliz-examples/examples/gcp-project-byoc-manual
```

### Understanding Configuration Files

The repository provides two JSON configuration files:

1. **terraform.tfvars.json**: Contains the minimum required parameters
2. **terraform.examples.tfvars.json**: Provides a more comprehensive configuration with custom values

You can use either file as a starting point, depending on your requirements.

### Required Parameters

At a minimum, you must specify the following parameters in your configuration:

- `gcp_project_id`: Your GCP project ID
- `gcp_vpc_cidr`: CIDR block for the VPC (e.g., "10.7.0.0/16")
- `zilliz_service_account`: The Zilliz service account that will be granted access to your resources

### Optional Parameters with Default Values

Many parameters have default values and can be omitted if the defaults are acceptable:

- `gcp_region`: Defaults to "us-west1"
- `customer_vpc_name`: Defaults to "zilliz-byoc-vpc"
- `customer_primary_subnet_name`: Defaults to "primary-subnet"
- `enable_private_link`: Defaults to true

## 5. Basic Configuration (Using Default Values)

For a basic setup, you can use the provided `terraform.tfvars.json` file with minimal modifications. This approach uses default values for most parameters, simplifying the initial setup.

### Step 1: Create or Modify terraform.tfvars.json

Create a `terraform.tfvars.json` file with the following content, replacing the placeholder values with your actual information:

```json
{
  "gcp_project_id": "your-gcp-project-id",
  "gcp_vpc_cidr": "10.7.0.0/16",
  "customer_vpc_name": "zilliz-byoc",
  "zilliz_service_account": "zilliz-service-account@zilliz-project.iam.gserviceaccount.com"
}
```

### Step 2: Initialize Terraform

Initialize the Terraform working directory:

```bash
terraform init
```

This command downloads the necessary provider plugins and sets up the backend.

### Step 3: Review the Execution Plan

Generate and review the execution plan:

```bash
terraform plan
```

This command shows what resources will be created, modified, or destroyed. Review the plan carefully to ensure it aligns with your expectations.

### Step 4: Apply the Configuration

Apply the configuration to create the resources:

```bash
terraform apply
```

When prompted, type `yes` to confirm the resource creation. The process may take several minutes to complete.

### Step 5: Verify the Deployment

After the deployment completes, Terraform will display the outputs, including:
- VPC ID
- Subnet IDs
- GCS bucket name
- Service account details

You can also verify the resources in the GCP Console.

## 6. Advanced Configuration (Using Custom Values)

For more control over your infrastructure, you can use the `terraform.examples.tfvars.json` file as a template for a customized configuration.

### Customizing Network Settings

You can customize the network configuration by specifying values for:

```json
{
  "gcp_project_id": "your-gcp-project-id",
  "gcp_region": "us-west1",
  "gcp_zones": ["us-west1-a", "us-west1-b", "us-west1-c"],
  "customer_vpc_name": "zilliz-byoc-gke-vpc",
  "gcp_vpc_cidr": "10.7.0.0/16",
  "customer_primary_subnet_name": "zilliz-primary-subnet",
  "customer_primary_subnet_cidr": "10.7.0.0/18",
  "customer_pod_subnet_name": "zilliz-pod-subnet",
  "customer_pod_subnet_cidr": "10.7.64.0/18",
  "customer_service_subnet_name": "zilliz-service-subnet",
  "customer_service_subnet_cidr": "10.7.128.0/18",
  "customer_lb_subnet_name": "zilliz-lb-subnet",
  "customer_lb_subnet_cidr": "10.7.192.0/20"
}
```

### Configuring Service Accounts

You can customize the service account names:

```json
{
  "customer_storage_service_account_name": "zilliz-byoc-gcp-storage-sa",
  "customer_management_service_account_name": "zilliz-byoc-management-sa",
  "customer_gke_node_service_account_name": "zilliz-byoc-gke-node-sa"
}
```

### GKE Cluster Configuration

Customize the GKE cluster name:

```json
{
  "customer_gke_cluster_name": "zilliz-byoc-gke"
}
```

### Storage Configuration

Customize the GCS bucket name:

```json
{
  "customer_bucket_name": "zilliz-byoc-gcp-storage"
}
```

### Private Link Configuration

Enable or disable private link connectivity:

```json
{
  "enable_private_link": true
}
```

### Complete Custom Configuration Example

A complete custom configuration might look like this:

```json
{
  "gcp_project_id": "your-gcp-project-id",
  "gcp_region": "us-west1",
  "gcp_zones": ["us-west1-a", "us-west1-b", "us-west1-c"],
  "customer_vpc_name": "zilliz-byoc-gke-vpc",
  "gcp_vpc_cidr": "10.7.0.0/16",
  "customer_primary_subnet_name": "zilliz-primary-subnet",
  "customer_primary_subnet_cidr": "10.7.0.0/18",
  "customer_pod_subnet_name": "zilliz-pod-subnet",
  "customer_pod_subnet_cidr": "10.7.64.0/18",
  "customer_service_subnet_name": "zilliz-service-subnet",
  "customer_service_subnet_cidr": "10.7.128.0/18",
  "customer_lb_subnet_name": "zilliz-lb-subnet",
  "customer_lb_subnet_cidr": "10.7.192.0/20",
  "enable_private_link": true,
  "customer_bucket_name": "zilliz-byoc-gcp-storage",
  "customer_gke_cluster_name": "zilliz-byoc-gke",
  "customer_storage_service_account_name": "zilliz-byoc-gcp-storage-sa",
  "customer_management_service_account_name": "zilliz-byoc-management-sa",
  "customer_gke_node_service_account_name": "zilliz-byoc-gke-node-sa",
  "zilliz_service_account": "zilliz-service-account@zilliz-project.iam.gserviceaccount.com"
}
```

## 7. Execution Steps

Regardless of whether you're using the basic or advanced configuration, the execution steps remain the same.

### Step 1: Initialize Terraform

Initialize the Terraform working directory:

```bash
terraform init
```

This command downloads the necessary provider plugins and sets up the backend.

### Step 2: Validate the Configuration

Validate the configuration files:

```bash
terraform validate
```

This command checks the syntax and internal consistency of the configuration.

### Step 3: Review the Execution Plan

Generate and review the execution plan:

```bash
terraform plan
```

This command shows what resources will be created, modified, or destroyed.

### Step 4: Apply the Configuration

Apply the configuration to create the resources:

```bash
terraform apply 
```

The process may take several minutes to complete. Terraform will display progress updates as resources are created.

### Step 5: Review the Outputs

After the deployment completes, Terraform will display the outputs, including:
- VPC ID and CIDR
- Subnet IDs and CIDRs
- GCS bucket name
- Service account details
- GKE cluster information

These outputs provide important information that you'll need to provide to Zilliz for the BYOC setup.

## 8. Resource Management

### Resources Created

The Terraform configuration creates the following resources:

1. **Networking Resources**:
   - VPC network
   - Subnets (primary, pod, service, and load balancer)
   - Firewall rules

2. **Storage Resources**:
   - GCS bucket for data storage

3. **Identity and Access Management**:
   - Service accounts for storage, management, and GKE nodes
   - IAM roles and permissions

4. **Connectivity Resources**:
   - Private Service Connect endpoints (if private link is enabled)

### Resource Dependencies

The resources are created in a specific order based on their dependencies:
- VPC network is created first
- Subnets are created after the VPC
- Service accounts are created in parallel with the network resources
- GCS bucket is created after the storage service account

### Accessing Created Resources

You can access and manage the created resources through:

1. **GCP Console**:
   - Navigate to the respective services (VPC Networks, IAM & Admin, Storage, etc.)
   - Filter by the names specified in your configuration

2. **Google Cloud SDK**:
   - Use `gcloud` commands to list and manage resources
   - Example: `gcloud compute networks list --filter="name=zilliz-byoc-vpc"`

3. **Terraform State**:
   - Use `terraform state list` to see all managed resources
   - Use `terraform state show [resource]` to see details of a specific resource

## 9. Troubleshooting

### Common Issues and Solutions

#### Insufficient Permissions

**Issue**: Error messages about insufficient permissions to create resources.

**Solution**:
- Verify that the account used has the required roles mentioned in the prerequisites
- Check for organization policies that might restrict resource creation
- Use `gcloud projects get-iam-policy [PROJECT_ID]` to check current permissions

#### API Not Enabled

**Issue**: Error messages about APIs not being enabled.

**Solution**:
- Enable the required APIs as mentioned in the prerequisites
- Use `gcloud services list --available` to see available APIs
- Use `gcloud services list --enabled` to see enabled APIs

#### Resource Quota Exceeded

**Issue**: Error messages about exceeding resource quotas.

**Solution**:
- Check your GCP quotas in the GCP Console
- Request quota increases if necessary
- Consider using a different region with higher quotas

#### Network Conflicts

**Issue**: Error messages about CIDR range conflicts.

**Solution**:
- Ensure the specified CIDR ranges don't overlap with existing networks
- Modify the CIDR ranges in your configuration
- Use `gcloud compute networks list --format="table(name,IPv4Range)"` to check existing networks

### Debugging Techniques

#### Increase Terraform Logging

Set the `TF_LOG` environment variable to get more detailed logs:

```bash
export TF_LOG=DEBUG
terraform apply
```

#### Check GCP Activity Logs

Review the GCP Activity Logs for error details:
1. Go to the GCP Console
2. Navigate to "Logging" > "Logs Explorer"
3. Filter by resource type and time range

#### Use Terraform Console

Use the Terraform console to inspect variables and expressions:

```bash
terraform console
```

### Error Messages and Their Meanings

#### "Error creating Network"

This usually indicates issues with the VPC configuration or permissions. Check that:
- The specified CIDR range is valid
- There are no conflicting networks
- The service account has the `compute.networks.create` permission

#### "Error creating Bucket"

This usually indicates issues with the GCS bucket configuration or permissions. Check that:
- The bucket name is globally unique
- The service account has the `storage.buckets.create` permission
- There are no organization policies restricting bucket creation

#### "Error setting IAM policy"

This usually indicates issues with IAM permissions. Check that:
- The service account has the `iam.serviceAccounts.setIamPolicy` permission
- The specified service accounts exist
- There are no organization policies restricting IAM changes

## 10. Cleanup and Maintenance

### Updating the Configuration

To update your configuration:

1. Modify the `terraform.tfvars.json` file with your changes
2. Run `terraform plan` to see what will change
3. Run `terraform apply` to apply the changes

Terraform will only modify resources that need to be changed, preserving the rest of your infrastructure.

### Destroying Resources

To remove all resources created by Terraform:

```bash
terraform destroy
```

When prompted, type `yes` to confirm. This will remove all resources managed by Terraform in the current configuration.

**Warning**: This action is irreversible and will delete all resources, including any data stored in the GCS bucket.

### Best Practices for Maintenance

1. **Version Control**:
   - Store your Terraform configuration in a version control system like Git
   - Document changes with meaningful commit messages

2. **State Management**:
   - Consider using remote state storage (e.g., GCS bucket)
   - Implement state locking to prevent concurrent modifications

3. **Regular Updates**:
   - Keep Terraform and provider versions updated
   - Regularly apply security patches and updates to your infrastructure

4. **Monitoring**:
   - Set up monitoring for the created resources
   - Configure alerts for potential issues

5. **Backup**:
   - Regularly backup your Terraform state file
   - Consider implementing disaster recovery procedures

## 11. Reference

### Variable Reference Table

| Variable Name | Description | Type | Default Value | Required |
|---------------|-------------|------|---------------|----------|
| gcp_project_id | The ID of the Google Cloud Platform project | string | - | Yes |
| gcp_region | The GCP region of the Google Cloud Platform project | string | "us-west1" | No |
| gcp_zones | The GCP zones for the GKE cluster | list(string) | null | No |
| gcp_vpc_cidr | The CIDR block for the customer VPC, cidr x/16 | string | - | Yes |
| customer_vpc_name | The VPC name of the Google Cloud Platform project | string | "zilliz-byoc-vpc" | No |
| customer_primary_subnet_name | The name of the primary subnet | string | "primary-subnet" | No |
| customer_primary_subnet_cidr | The CIDR block for the primary subnet | string | "" | No |
| customer_pod_subnet_name | The name of the pod subnet | string | "" | No |
| customer_pod_subnet_cidr | The CIDR block for the pod subnet | string | "" | No |
| customer_service_subnet_name | The name of the service subnet | string | "" | No |
| customer_service_subnet_cidr | The CIDR block for the service subnet | string | "" | No |
| customer_lb_subnet_name | The name of the load balancer subnet | string | "" | No |
| customer_lb_subnet_cidr | The CIDR block for the load balancer subnet | string | "" | No |
| customer_bucket_name | The name of the GCS bucket | string | "" | No |
| customer_gke_cluster_name | The name of the GKE cluster | string | "" | No |
| customer_storage_service_account_name | The name of the storage service account | string | "" | No |
| customer_management_service_account_name | The name of the management service account | string | "" | No |
| customer_gke_node_service_account_name | The name of the gke node service account | string | "" | No |
| zilliz_service_account | The service account that can impersonate the customer service account | string | - | Yes |
| enable_private_link | Whether to enable private link | bool | true | No |

### Output Reference Table

| Output Name | Description |
|-------------|-------------|
| gcp_project_id | The GCP project ID |
| vpc_id | The ID of the created VPC |
| vpc_name | The name of the created VPC |
| primary_subnet_id | The ID of the primary subnet |
| pod_subnet_id | The ID of the pod subnet |
| service_subnet_id | The ID of the service subnet |
| lb_subnet_id | The ID of the load balancer subnet |
| gcs_bucket_name | The name of the created GCS bucket |
| storage_service_account_email | The email of the storage service account |
| management_service_account_email | The email of the management service account |
| gke_node_service_account_email | The email of the GKE node service account |
| private_link_endpoint | The endpoint for private link connectivity (if enabled) |

### Useful GCP and Terraform Commands

#### Terraform Commands

```bash
# Initialize Terraform
terraform init

# Validate configuration
terraform validate

# Plan changes
terraform plan

# Apply changes
terraform apply

# Destroy resources
terraform destroy

# List resources in state
terraform state list

# Show resource details
terraform state show [resource]

# Import existing resources
terraform import [resource_address] [resource_id]
```

#### GCP Commands

```bash
# List VPC networks
gcloud compute networks list

# List subnets
gcloud compute networks subnets list

# List service accounts
gcloud iam service-accounts list

# List GCS buckets
gsutil ls

# Check IAM policy
gcloud projects get-iam-policy [PROJECT_ID]

# Enable APIs
gcloud services enable [API_NAME]
```

### Additional Resources

- [Terraform Documentation](https://www.terraform.io/docs)
- [Google Cloud Provider Documentation](https://registry.terraform.io/providers/hashicorp/google/latest/docs)
- [Zilliz Cloud Documentation](https://docs.zilliz.com/)
- [GCP Documentation](https://cloud.google.com/docs)
- [Terraform Best Practices](https://www.terraform-best-practices.com/)

By following this guide, you should now have a fully configured GCP environment ready for Zilliz BYOC deployment. The infrastructure created provides the necessary resources for running Zilliz's vector database services within your own GCP environment, giving you greater control over your data while still benefiting from Zilliz's managed service capabilities.
