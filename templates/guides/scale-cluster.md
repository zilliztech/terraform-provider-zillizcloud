# Upgrading Zilliz Cloud Cluster Compute Unit Size with Terraform

This tutorial guides you through upgrading the compute unit (CU) size of an existing Zilliz Cloud cluster using Terraform. Scaling your cluster's CU size allows you to adjust its processing power to meet the demands of your workloads.

### Prerequisites

Before you begin, make sure you have completed the initial setup steps outlined in the [Getting Started with Zilliz Cloud Terraform Provider](./get-start.md) guide. Additionally, ensure that you have the necessary permissions and access credentials to interact with the Zilliz Cloud API.

### Understanding Compute Units (CUs)

Zilliz Cloud clusters are allocated resources in the form of compute units (CUs). Each CU represents a specific amount of processing power, memory, and storage. Upgrading the CU size of your cluster increases the overall resources available to handle more demanding workloads.

**Important Note:** The available CU sizes and pricing may vary depending on your Zilliz Cloud account, region, and chosen service plan. Refer to the Zilliz Cloud documentation for detailed information on CU sizes and pricing for your specific configuration.


### Upgrading the CU Size with Terraform

Here's how to leverage Terraform to upgrade the CU size for your Zilliz Cloud cluster:

1. **Modify Terraform Configuration**: Locate the Terraform configuration block defining your Zilliz Cloud cluster resource (typically named `zillizcloud_cluster`).

2. **Identify `cu_size` Attribute**: Within the cluster definition, identify the `cu_size` attribute. This attribute specifies the current number of CUs allocated to your cluster.

3. **Specify New CU Size**: Update the value of the `cu_size` attribute to reflect the desired increase in resources. Refer to the Zilliz Cloud documentation for valid CU size options within your region and service plan.

**Example:**

Assuming your current cluster configuration is similar to the following:

```hcl
resource "zillizcloud_cluster" "standard_plan_cluster" {
  cluster_name = "Cluster-02"
  # ... other fields ...
  cu_size       = 1  # Current CU size
  # ... other fields ...
}
```

To upgrade the cluster's CU size to 2, modify the configuration as follows:

```hcl
resource "zillizcloud_cluster" "standard_plan_cluster" {
  cluster_name = "Cluster-02"
  # ... other fields ...
  cu_size       = 2  # Upgraded CU size
  # ... other fields ...
}
```

4. **Save Changes**: Save the modifications made to your Terraform configuration file.


### Applying the Upgrade


**Apply Changes**: If the plan looks good, proceed with applying the changes to upgrade the CU size in Zilliz Cloud:

   ```bash
   terraform apply -auto-approve
   ```

**Note**: The `-auto-approve` flag avoids prompting for confirmation before applying the changes. Use caution, especially in production environments. It's recommended to thoroughly review the plan before applying.

Terraform will now initiate the upgrade process, scaling your Zilliz Cloud cluster to the specified CU size.


## Destroying the Cluster(Optional)
If you want to destroy the cluster, you can run the following command:
```
$ terraform destroy
```

### Verifying the Upgrade

Once Terraform finishes applying the changes, you can verify that the CU size of your cluster has been upgraded successfully via the Zilliz Cloud dashboard. Log in to your Zilliz Cloud account and navigate to the Clusters section in the dashboard. Select the upgraded cluster and view its details. The CU size should now reflect the value you specified in your Terraform configuration.

