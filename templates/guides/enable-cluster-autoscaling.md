# Migrating from Fixed CU Size to Dynamic CU Autoscaling

This guide walks you through migrating your Zilliz Cloud cluster from a fixed compute unit (CU) size to dynamic CU autoscaling. Dynamic autoscaling automatically adjusts your cluster's compute resources based on workload demands, optimizing both performance and cost.

## Prerequisites

Before you begin, ensure you have:

- Completed the initial setup steps outlined in the [Getting Started with Zilliz Cloud Terraform Provider](./get-start.md) guide
- An existing cluster configured with `cu_size`
- Necessary permissions and access credentials to interact with the Zilliz Cloud API
- An Enterprise plan cluster (autoscaling is available on Enterprise clusters)

## Understanding CU Autoscaling

### Fixed CU Size vs. Dynamic Autoscaling

**Fixed CU Size (`cu_size`)**:
- Static allocation of compute units
- Cluster always uses the same amount of resources
- Simple to configure but may be over-provisioned during low usage or under-provisioned during peak times

**Dynamic CU Autoscaling (`cu_settings`)**:
- Automatically scales compute units between a minimum and maximum threshold
- Adjusts resources based on actual workload demands
- Optimizes costs by using fewer resources during low activity periods
- Ensures performance by scaling up during high demand

### Important Constraints

- **Mutual Exclusivity**: `cu_size` and `cu_settings` cannot be configured together on the same cluster
- **Validation Rules**:
  - Minimum CU must be at least 1
  - Maximum CU must be greater than or equal to minimum CU
- **Migration Requirement**: You must remove `cu_size` before adding `cu_settings` (requires two separate Terraform operations)

## Migration Process

### Step 1: Review Your Current Configuration

First, examine your existing cluster configuration. It likely looks similar to this:

```hcl
resource "zillizcloud_cluster" "example" {
  cluster_name = "my-production-cluster"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  cu_size      = 4  # Fixed at 4 CUs
}
```



### Step 2: Enable CU Autoscaling Configuration

Now update your configuration to add the `cu_settings` block with your desired autoscaling parameters.

**Important:** Make sure to remove the `cu_size` attribute from your configuration.

```hcl
resource "zillizcloud_cluster" "example" {
  cluster_name = "my-production-cluster"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id

  cu_settings = {
    dynamic_scaling = {
      min = 2  # Minimum CUs during low activity
      max = 8  # Maximum CUs during peak load
    }
  }
}
```

Apply the autoscaling configuration:

```bash
terraform plan
terraform apply
```

**Success!** Your cluster now has dynamic autoscaling enabled and will automatically adjust compute resources between 2 and 8 CUs based on workload.



## Common Errors and Troubleshooting

### Error: "These attributes cannot be configured together"

**Cause**: You attempted to set both `cu_size` and `cu_settings` simultaneously.

**Solution**: Follow the two-step migration process outlined above. Remove `cu_size` first, apply the change, then add `cu_settings`.

### Error: "Invalid autoscaling configuration: Minimum CU must be less than or equal to maximum CU"

**Cause**: You set `min` greater than `max`.

**Solution**: Ensure `min` ≤ `max` in your configuration:

```hcl
cu_settings = {
  dynamic_scaling = {
    min = 2   # Must be <= max
    max = 8
  }
}
```

### Error: Validation failed on minimum CU

**Cause**: Minimum CU value is less than 1.

**Solution**: Set `min` to at least 1:

```hcl
cu_settings = {
  dynamic_scaling = {
    min = 1   # Minimum allowed value
    max = 4
  }
}
```

## Reverting to Fixed CU Size

If you need to revert from autoscaling to a fixed CU size:

### Step 1: Remove `cu_settings`

```hcl
resource "zillizcloud_cluster" "example" {
  cluster_name = "my-production-cluster"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  # cu_settings removed
}
```

Apply the change:

```bash
terraform apply
```

### Step 2: Add Fixed `cu_size`

```hcl
resource "zillizcloud_cluster" "example" {
  cluster_name = "my-production-cluster"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  cu_size = 6  # Return to fixed size
}
```

Apply the change:

```bash
terraform apply
```

## Monitoring Autoscaling Behavior

After enabling autoscaling:

1. **Monitor via Zilliz Cloud Console**: Log in to your Zilliz Cloud account and navigate to your cluster's details page to observe autoscaling behavior
2. **Review Performance Metrics**: Check query latencies and throughput to ensure autoscaling meets your performance requirements
3. **Analyze Cost Impact**: Compare costs before and after enabling autoscaling to quantify savings

## Next Steps

- Review the [Cluster Resource Documentation](https://registry.terraform.io/providers/zillizcloud/zillizcloud/latest/docs/resources/cluster) for complete configuration options
- Learn about [Scaling Clusters](./scale-cluster.md) for more cluster management techniques
- Explore [Creating Enterprise Clusters](./create-a-standard-cluster.md) for best practices

## Additional Resources

- [Zilliz Cloud Documentation](https://docs.zilliz.com/)
- [Terraform Provider Registry](https://registry.terraform.io/providers/zillizcloud/zillizcloud/latest)
