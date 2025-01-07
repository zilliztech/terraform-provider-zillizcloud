## Importing Existing Zilliz Cloud Clusters with Terraform

This tutorial guides you through importing existing Zilliz Cloud clusters into your Terraform state file. Importing allows you to manage these clusters using Terraform configurations, enabling infrastructure as code practices.

### Prerequisites

Before you begin, make sure you have completed the initial setup steps outlined in the [Getting Started with Zilliz Cloud Terraform Provider](./get-start.md) guide. Additionally, ensure that you have the necessary permissions and access credentials to interact with the Zilliz Cloud API.

### Understanding Cluster Importing

Importing allows you to bring existing Zilliz Cloud clusters under Terraform's management. This means Terraform will track the cluster's configuration and resources within its state file. By managing your clusters through Terraform, you can leverage infrastructure as code (IaC) practices for consistent and automated cluster provisioning and management.


### Steps to Import a Zilliz Cloud Cluster

Here's a step-by-step guide on importing an existing Zilliz Cloud cluster into your Terraform state:

1. **Identify Cluster Details**: Gather the following information about the Zilliz Cloud cluster you intend to import:
    * **Cluster ID**: A unique identifier assigned to the cluster within Zilliz Cloud. You can find this ID in the Zilliz Cloud dashboard (e.g., `"in01-1da9c8686b882d0"`).
    * **Region ID**: The region where the cluster is deployed (e.g., `aws-us-east-2`).

2. **Import zillizcloud_cluster resource**: In order to import a existing cluster, it's recommended to create a basic `zillizcloud_cluster` resource block in your Terraform configuration file. This block will serve as a placeholder for the imported cluster and can be further customized later. Here's an example:

   ```hcl
   resource "zillizcloud_cluster" "imported_cluster" {
     # ... attributes will be populated during import process ...
   }
   ```

3. **Import the Cluster**: Use the `terraform import` command to instruct Terraform to import the existing cluster. The general syntax is:

   ```bash
   terraform import RESOURCE_TYPE.RESOURCE_NAME CLUSTER_ID,REGION_ID
   ```

   Replace the placeholders with the following details:

     * `RESOURCE_TYPE`: Set this to `"zillizcloud_cluster"`.
     * `RESOURCE_NAME`: This is the name you'll assign to the imported cluster resource in Terraform (e.g., `"imported_cluster"`).
     * `RESOURCE_ID`: Provide the cluster ID obtained in step 1 (e.g., `"in01-1da9c8686b882d0"`).
     * `REGION_ID`: The region where the cluster is deployed(e.g., `"aws-us-east-2"`).


   Here's an example command assuming you named the imported cluster resource `"standard_plan_cluster"`:

   ```bash
   terraform import zillizcloud_cluster.imported_cluster in01-1da9c8686b882d0,aws-us-east-2
   ```

4. **Confirm Import**: Terraform will attempt to retrieve information about the cluster from the Zilliz Cloud API and populate the corresponding resource block in your Terraform configuration. If successful, you'll see confirmation messages indicating a successful import.

5. **Verify Import**: After a successful import, you can use the `terraform state show` command to view the details of the imported cluster within your Terraform state:

   ```bash
   terraform state show zillizcloud_cluster.imported_cluster

    resource "zillizcloud_cluster" "imported_cluster" {
        cluster_name    = "Cluster-02"
        cluster_type    = "Performance-optimized"
        connect_address = "https://in01-1da9c8686b882d0.aws-us-east-2.vectordb.zillizcloud.com:19539"
        create_time     = "2024-05-15T07:36:21Z"
        cu_size         = 1
        description     = "pre create instance"
        id              = "in01-1da9c8686b882d0"
        project_id      = "proj-4487580fcfe2c8a4391686"
        region_id       = "aws-us-east-2"
        status          = "RUNNING"
    }
   ```

   This command will display the attributes and current configuration of the imported cluster.

