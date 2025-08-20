# Managing Milvus Clusters in BYOC Environments 

This document provides a comprehensive guide on using the Zilliz Terraform provider to manage Milvus clusters in a Bring Your Own Cloud (BYOC) environment. BYOC allows you to deploy and manage Milvus clusters on your own cloud infrastructure while leveraging Zilliz Cloud's management capabilities. The examples focus on key operations such as creating, destroying, suspending, resuming, scaling (via CU size changes), and adjusting replicas for clusters.

The Zilliz Terraform provider simplifies infrastructure-as-code practices, enabling automated provisioning and management of resources. Ensure you have Terraform installed  and the Zilliz provider configured correctly.

**Prerequisites:**

- A Zilliz Cloud account with API access.
- An API key generated from the Zilliz Cloud console.
- Access to a BYOC-enabled project in Zilliz Cloud.
- Familiarity with Terraform basics, including `terraform init`, `terraform plan`, and `terraform apply`.

All examples use the `zillizcloud` provider. Replace placeholders (e.g., API keys, project IDs, cluster names) with your actual values.

## 1. Creating a Milvus Cluster

To create a new Milvus cluster, define the provider configuration and the cluster resource in your Terraform configuration file (e.g., `main.tf`). This example provisions an enterprise-plan cluster optimized for performance.

```hcl
terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
  # Your Zilliz Cloud API key.
  # For security, it's highly recommended to use an environment variable
  # instead of hardcoding the key here.
  # Example: export ZILLIZCLOUD_API_KEY="your_api_key"
  api_key = "YOUR_API_KEY"
}

data "zillizcloud_project" "this" {
  # Fetches the default project information for cluster provisioning.
  # You can specify a project by name or ID. Here, we use an existing project ID.
  id = "proj-xxxxxxxxx"
}

resource "zillizcloud_cluster" "a_cluster" {
  # (Required) A unique and descriptive name for the cluster.
  cluster_name = "a-cluster-name"

  # (Optional) The size of the Compute Unit (CU), defining the cluster's base resources.
  cu_size      = 1

  # (Optional) The type of Compute Unit. Valid values: "Performance-optimized", "Capacity-optimized".
  cu_type      = "Performance-optimized"

  # (Required) The ID of the project where the cluster will be created.
  project_id   = data.zillizcloud_project.this.id

  # (Optional) Labels are key-value pairs that are attached to your Milvus pods. By using these labels, you can tag each pod with specific identifiers, such as a team name or service name etc. This allows your team to easily monitor compute resource consumption for each labeled workload and accurately allocate the associated infrastructure costs to the correct team or project.
  labels = {
    tenant = "infra"
  }
}
```

### Steps to Apply:

1. Run `terraform init` to initialize the provider.
2. Run `terraform plan` to preview changes.
3. Run `terraform apply` to create the cluster. Confirm with `yes` when prompted.

After application, the cluster will be in a "RUNNING" state. Monitor the status in the Zilliz Cloud console or via Terraform outputs. Creation may take several minutes depending on the region and plan.

## 2. Destroying a Milvus Cluster

To permanently delete a cluster and its associated resources, use Terraform's destroy command. This is irreversible—ensure all data is backed up.

Update your `main.tf` if needed (e.g., remove or comment out unrelated resources), then run:

```bash
terraform apply
```

Confirm with `yes`. This will deprovision the cluster defined in the configuration. Always review the plan with `terraform plan -destroy` beforehand to avoid accidental deletions.

**Note:** Destroying a cluster releases all resources, including data. Use suspension (see below) for temporary pauses.

## 3. Suspending a Milvus Cluster

Suspending a cluster stops it to save costs while preserving data and configuration. Update the `desired_status` attribute in your cluster resource.

```hcl
resource "zillizcloud_cluster" "enterprise_plan_cluster" {
  # ... (other attributes remain the same)
  desired_status = "SUSPENDED"  # Sets the cluster to suspended state
  # ...
}
```

Apply the change with `terraform apply`. The cluster will transition to "SUSPENDED" status, halting compute costs but retaining storage.

**Best Practices:** Use suspension for non-production environments or during off-hours. Resuming (next section) is quick and restores the cluster to its previous state.

## 4. Resuming a Milvus Cluster

To restart a suspended cluster, update the `desired_status` to "RUNNING".

```hcl
resource "zillizcloud_cluster" "enterprise_plan_cluster" {
  # ... (other attributes remain the same)
  desired_status = "RUNNING"  # Resumes the cluster to active state
  # ...
}
```

Run `terraform apply` to apply. The cluster will become operational again, typically within minutes.

**Monitoring:** After resuming, verify connectivity and performance in the Zilliz Cloud dashboard.

## 5. Scaling CU Size

Scale the cluster's capacity by adjusting the `cu_size`. This is useful for handling increased load without downtime (horizontal scaling).

```hcl
resource "zillizcloud_cluster" "enterprise_plan_cluster" {
  # ... (other attributes remain the same)
  cu_size = 2  # Increases CU size from 1 to 2 
  # ...
}
```

Apply with `terraform apply`. Scaling operations are online, but monitor for brief performance impacts. 

**Considerations:** Scaling up increases costs; test in a staging environment. Use Zilliz metrics to determine optimal size based on query volume and data size.

## 6. Adjusting Replicas

Increase replicas for high availability and fault tolerance. This creates additional copies of the cluster's data shards.

```hcl
resource "zillizcloud_cluster" "enterprise_plan_cluster" {
  # ... (other attributes remain the same)
  replica = 2  # Sets number of replicas 
  # ...
}
```

Apply with `terraform apply`. Replicas enhance read availability and resilience but increase storage costs.



## Additional Tips

- **Idempotency:** Terraform ensures configurations are idempotent—repeated applies won't recreate resources unnecessarily.

- **Error Handling:** Common errors include invalid API keys or region mismatches. Check Terraform logs and Zilliz Cloud console for details.

- **Outputs:** Add outputs in `main.tf` to expose details like cluster endpoint:

  ```hcl
  output "cluster_endpoint" {
    value = zillizcloud_cluster.this.endpoint
  }
  ```

- **Version Control:** Store your `.tf` files in Git for collaboration and versioning.

- **Security:** Never hardcode sensitive data like API keys; use environment variables (e.g., `TF_VAR_api_key`) or secrets managers.

- **Limitations:** BYOC requires pre-configured cloud accounts. Consult Zilliz documentation for supported regions and plans.

- **Further Reading:** Refer to the official Zilliz Terraform provider docs at [Zilliz GitHub](https://github.com/zilliztech/terraform-provider-zillizcloud) for advanced configurations.

If you encounter issues or need customizations, provide more details for further assistance!

