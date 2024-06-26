## Tutorial: Creating a Standard Plan Cluster with Terraform

This tutorial guides you through managing a **Standard Plan** Cluster using the `zillizcloud_cluster` resource within Terraform.  Standard Plan Cluster have more resources, compared to Free Plan Cluster and are suitable for production workloads.

Check out the [Select the Right Cluster Plan](https://docs.zilliz.com/docs/select-zilliz-cloud-service-plans) for more information on the available plans.

**You'll learn how to:**
- Retrieve project and region IDs for cluster creation.
- Define clusters with various configurations in Terraform.
- Plan and apply changes to provision clusters in Zilliz Cloud.
- Verify the creation of your Zilliz Cloud clusters.

### Prerequisites

Before you begin, make sure you have completed the initial setup steps outlined in the [Getting Started with Zilliz Cloud Terraform Provider](./get-start.md) guide. Additionally, ensure that you have the necessary permissions and access credentials to interact with the Zilliz Cloud API.


## Retrieving Your Zilliz Cloud Project ID

There are two ways to find your Zilliz Cloud project ID:

**1. Zilliz Cloud Console:**

Refer to the Zilliz Cloud documentation for details on obtaining your project ID through the console: [How Can I Obtain the Project ID?](https://support.zilliz.com/hc/en-us/articles/22048954409755-How-Can-I-Obtain-the-Project-ID)

**2. Zilliz Cloud Data Source (Optional):**

For convenience, you can use the `zillizcloud_project` data source to retrieve the ID of the default project associated with your Zilliz Cloud account.

## Selecting a Zilliz Cloud Region

This tutorial will use the `aws-us-east-2` region.  However, you can choose a different region based on your needs.

For a full list of available cloud providers and regions, refer to the Zilliz Cloud documentation: [Cloud Providers & Regions](https://docs.zilliz.com/docs/cloud-providers-and-regions)
 
### Creating a Cluster

With the `project ID` and `region ID` at hand, creating Zilliz Cloud clusters is straightforward. Below is an illustrative example defining zillizcloud_cluster resources within Terraform configuration:

Append the following code to your `main.tf` file mentioned in the [Getting Started with Zilliz Cloud Terraform Provider](./get-start.md) guide.

```hcl

data "zillizcloud_project" "default" {
  # Fetching the default project information to be used in cluster provisioning
}

resource "zillizcloud_cluster" "standard_plan_cluster" {
  cluster_name = "Cluster-02"                        # The name of the cluster
  region_id    = "aws-us-east-2"                     # The region where the cluster will be deployed
  plan         = "Standard"                          # The service plan for the cluster
  cu_size      = "1"                                 # The size of the compute unit
  cu_type      = "Performance-optimized"             # The type of compute unit, optimized for performance
  project_id   = data.zillizcloud_project.default.id # Linking to the project ID fetched earlier
}


```
**Explanation of the Configuration:**

This code creates a Zilliz Cloud cluster named "Cluster-02" in the "aws-us-east-2" region. It serves the "Standard" plan with a single, performance-optimized compute unit, making it suitable for workloads requiring more processing power.


### Planning and Applying Changes

Once you've defined the cluster resources, it's time to apply the changes to provision the Zilliz Cloud clusters. Navigate to your Terraform project directory and execute the following commands:

```bash
terraform apply -auto-approve
```

**Note**: The `-auto-approve` flag avoids prompting for confirmation before applying the changes. Use caution, especially in production environments. It's recommended to thoroughly review the plan before applying.

Terraform will orchestrate the creation of the specified clusters based on your configuration. Once the process is complete, Terraform will display the output values for the cluster's connection address, username, and password.

### Verifying Provisioned Clusters

After applying the changes, you can verify the provisioned Zilliz Cloud clusters either through the Zilliz Cloud dashboard. Ensure that the clusters are created with the desired configurations and are functioning as expected.

## Destroying the Cluster(Optional)
If you want to destroy the cluster, you can run the following command:
```
$ terraform destroy
```

## Next Steps
- [Upgrading Zilliz Cloud Cluster Compute Unit Size with Terraform](./scale-cluster.md)
