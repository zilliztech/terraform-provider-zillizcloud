# Tutorial: Creating Zilliz Cloud Cluster Resources with Terraform

This tutorial guides you through managing Zilliz Cloud clusters using the `zillizcloud_cluster` resource within Terraform. You'll learn how to:

- Retrieve project and region IDs for cluster creation.
- Define clusters with various configurations in Terraform.
- Plan and apply changes to provision clusters in Zilliz Cloud.
- Verify the creation of your Zilliz Cloud clusters.

### Prerequisites

Before you begin, make sure you have completed the initial setup steps outlined in the [Getting Started with Zilliz Cloud Terraform Provider](./get-start.md) guide. Additionally, ensure that you have the necessary permissions and access credentials to interact with the Zilliz Cloud API.

## Retrieving Project IDs in Your Zilliz Cloud Account

Every Zilliz Cloud cluster is associated with a project. To create a cluster, it's imperative to obtain the corresponding project ID.

To access information regarding available projects, utilize the zillizcloud_projects data source as demonstrated below:

```hcl
data "zillizcloud_project" "default" {}

output "projects" {
  value = data.zillizcloud_project.default
}
```

```shell
$ terraform apply --auto-approve

data.zillizcloud_project.default: Reading...
data.zillizcloud_project.default: Read complete after 1s [id=proj-4487580fcfe2c8a4391686]

Apply complete! Resources: 0 added, 0 changed, 0 destroyed.

Outputs:

projects = {
  "created_at" = 1714892175000
  "id" = "proj-4487580fcfe2cxxxxx"
  "instance_count" = 0
  "name" = "Default Project"
}
```

The project ID **"proj-4487580fcfe2cxxxxx"** is displayed in the output section. You can use your ID to create Zilliz Cloud clusters within the specified project, or refer to the project ID via `data.zillizcloud_project.default.id` in your Terraform configuration file.

## Acquiring Region IDs for Zilliz Cloud Cluster



To provision Zilliz Cloud clusters, you'll must specify the region where the cluster will be deployed. Zilliz Cloud supports multiple regions across various cloud providers, such as AWS, GCP, and Azure. You can retrieve the region IDs for each cloud provider using the zillizcloud_regions data source.

```terraform

data "zillizcloud_regions" "aws_region" {
  cloud_id = "aws"
}

data "zillizcloud_regions" "gcp_region" {
  cloud_id = "gcp"
}

data "zillizcloud_regions" "azure_region" {
  cloud_id = "azure"
}

output "aws_ouput" {
  value = data.zillizcloud_regions.aws_region.items
}


output "gcp_ouput" {
  value = data.zillizcloud_regions.gcp_region.items
}

output "azure_ouput" {
  value = data.zillizcloud_regions.azure_region.items
}
```


```
$ terraform apply --auto-approve

learn-terraform terraform apply -auto-approve
data.zillizcloud_regions.gcp_region: Reading...
data.zillizcloud_regions.aws_region: Reading...
data.zillizcloud_regions.azure_region: Reading...
data.zillizcloud_project.default: Reading...
data.zillizcloud_regions.aws_region: Read complete after 0s
data.zillizcloud_regions.gcp_region: Read complete after 0s
data.zillizcloud_regions.azure_region: Read complete after 0s
data.zillizcloud_project.default: Read complete after 0s [id=proj-4487580fcfe2c8a4391686]

You can apply this plan to save these new output values to the Terraform state, without changing any real infrastructure.

Apply complete! Resources: 0 added, 0 changed, 0 destroyed.

Outputs:

aws_ouput = tolist([
...
  {
    "api_base_url" = "https://api.aws-us-west-2.zillizcloud.com"
    "cloud_id" = "aws"
    "region_id" = "aws-us-west-2"
  },
...
])
azure_ouput = tolist([
...
])
gcp_ouput = tolist([
...
])
```

Upon execution, this Terraform script retrieves region details for each cloud provider, facilitating subsequent cluster provisioning steps. In this session, we would use **aws-us-east-2** in the following example.


### Creating a Cluster

With the `project ID` and `region ID` at hand, creating Zilliz Cloud clusters is straightforward. Below is an illustrative example defining zillizcloud_cluster resources within Terraform configuration:


```hcl
resource "zillizcloud_cluster" "starter_cluster" {
  cluster_name = "Cluster-01"
  project_id   = data.zillizcloud_project.default.id
}

resource "zillizcloud_cluster" "standard_plan_cluster" {
  cluster_name = "Cluster-02"
  region_id    = "aws-us-east-2"
  plan         = "Standard"
  cu_size      = "1"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
}
```
This example will create two Zilliz Cloud clusters:
- **Starter Cluster**: This configuration creates a basic cluster named "Cluster-01" within the default project. It's suitable for initial testing or small-scale applications.
- **Standard Plan Cluster**: This configuration creates a more advanced cluster named "Cluster-02" in the `aws-us-east-2` region. It uses the "Standard" service plan with a compute unit size of 1, optimized for performance, making it suitable for more demanding workloads.

### Planning and Applying Changes

Once you've defined the cluster resources, it's time to apply the changes to provision the Zilliz Cloud clusters. Navigate to your Terraform project directory and execute the following commands:

```bash
terraform apply -auto-approve
```

**Note**: The `-auto-approve` flag avoids prompting for confirmation before applying the changes. Use caution, especially in production environments. It's recommended to thoroughly review the plan before applying.

Review the plan generated by Terraform and confirm to apply the changes. Terraform will orchestrate the creation of the specified clusters based on your configuration.

### Verifying Provisioned Clusters

After applying the changes, you can verify the provisioned Zilliz Cloud clusters either through the Zilliz Cloud dashboard. Ensure that the clusters are created with the desired configurations and are functioning as expected.

