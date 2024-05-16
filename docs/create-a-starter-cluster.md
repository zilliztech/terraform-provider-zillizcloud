## Tutorial: Creating a Starter Cluster with Terraform

This tutorial guides you through creating a basic **Starter cluster** in Zilliz Cloud using the `zillizcloud_cluster` resource within Terraform. Starter clusters are suitable for initial testing or small-scale applications.

**You'll learn how to:**

- Define a **Starter cluster** in Terraform configuration.
- Review and apply changes to provision the cluster in Zilliz Cloud.
- Verify the creation of your Zilliz Cloud cluster.

### Prerequisites

Before you begin, make sure you have completed the initial setup steps outlined in the [Getting Started with Zilliz Cloud Terraform Provider](./get-start.md) guide. Additionally, ensure that you have the necessary permissions and access credentials to interact with the Zilliz Cloud API.

### Creating a Cluster

Append the following code to your `main.tf` file mentioned in the [Getting Started with Zilliz Cloud Terraform Provider](./get-start.md) guide.


```hcl

data "zillizcloud_project" "default" {
  # Fetching the default project information to be used in cluster provisioning
}

resource "zillizcloud_cluster" "starter_cluster" {
  # Name for your cluster
  cluster_name = "Cluster-01"
  # ID of the default project retrieved from data source
  project_id   = data.zillizcloud_project.default.id 
}

output "cluster_connect_address" {
  value = zillizcloud_cluster.starter_cluster.connect_address
}
output "cluster_username" {
  value = zillizcloud_cluster.starter_cluster.username
}
output "cluster_password" {
  sensitive = true
  value = zillizcloud_cluster.starter_cluster.password
}
```


**Explanation:**

- This configuration defines a `zillizcloud_cluster` resource named `starter_cluster`.
- * `cluster_name`: Sets the name of your cluster (here, "Cluster-01").
  * `project_id`: Retrieves the ID of the default project using the `data.zillizcloud_project.default.id` attribute.

### Planning and Applying Changes

Once you've defined the cluster resources, it's time to apply the changes to provision the Zilliz Cloud clusters. Navigate to your Terraform project directory and execute the following commands:

```bash
terraform apply -auto-approve

Outputs:

cluster_connect_address = "https://in03-559dde3b4b6de3a.api.gcp-us-west1.zillizcloud.com"
cluster_password = <sensitive>
cluster_username = "db_559dde3b4b6de3a"
```

**Note**: The `-auto-approve` flag avoids prompting for confirmation before applying the changes. Use caution, especially in production environments. It's recommended to thoroughly review the plan before applying.

Terraform will orchestrate the creation of the specified clusters based on your configuration. Once the process is complete, Terraform will display the output values for the cluster's connection address, username, and password.

### Verifying Provisioned Clusters

After applying the changes, you can verify the provisioned Zilliz Cloud clusters either through the Zilliz Cloud dashboard. Ensure that the clusters are created with the desired configurations and are functioning as expected.

## Destroying the Cluster(Optional)
if you want to destroy the cluster, you can run the following command:
```
$ terraform destroy
```

## Next Steps
- Explore creating a Standard Cluster with more resources using the guide: [Creating a Standard Cluster](./create-a-standard-cluster.md)

