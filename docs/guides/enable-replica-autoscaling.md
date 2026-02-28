# Migrating from Fixed Replica to Replica Autoscaling

This guide walks you through migrating your Zilliz Cloud cluster from a fixed replica count to replica autoscaling. Zilliz Cloud supports two types of replica autoscaling:

- **Dynamic Autoscaling**: Automatically adjusts the number of replicas based on real-time workload demands
- **Scheduled Scaling**: Scales replicas at predetermined times using cron expressions

For CU (compute unit) autoscaling, see the [CU Autoscaling Guide](./enable-cluster-autoscaling.md).

## Prerequisites

Before you begin, ensure you have:

- Completed the initial setup steps outlined in the [Getting Started with Zilliz Cloud Terraform Provider](./get-start.md) guide
- An existing cluster configured with a fixed `replica` value
- Necessary permissions and access credentials to interact with the Zilliz Cloud API
- An Enterprise plan cluster (replica autoscaling is available on Enterprise clusters)

## Understanding Replica Autoscaling

### What are Replicas?

Replicas are copies of your data distributed across multiple nodes. Increasing replicas improves:
- **Query throughput**: More replicas can handle more concurrent queries
- **High availability**: If one replica fails, others continue serving requests
- **Read scalability**: Distribute read load across multiple replicas

### Fixed Replica vs. Autoscaling Options

**Fixed Replica (`replica`)**:
- Static number of replicas
- Cluster always maintains the same replica count
- Simple to configure but may be over-provisioned during low usage or under-provisioned during peak times

**Dynamic Autoscaling (`replica_settings.dynamic_scaling`)**:
- Automatically scales replicas between a minimum and maximum threshold
- Adjusts based on actual workload demands
- Optimizes costs by using fewer replicas during low activity periods
- Ensures performance by scaling up during high demand

**Scheduled Scaling (`replica_settings.schedule_scaling`)**:
- Scales replicas at specific times defined by cron expressions
- Ideal for predictable traffic patterns (e.g., business hours, batch processing windows)
- Supports timezone configuration for global deployments
- Can define multiple schedules for different scaling targets

### When to Use Each Approach

| Scenario | Recommended Approach |
|----------|---------------------|
| Unpredictable query loads | Dynamic Autoscaling |
| Predictable daily/weekly traffic patterns | Scheduled Scaling |
| Simple, consistent workloads | Fixed Replica |

### Important Constraints

- **Mutual Exclusivity**: `replica` and `replica_settings` cannot be configured together on the same cluster
- **Validation Rules**:
  - Minimum replica must be at least 1 (for dynamic scaling)
  - Maximum replica must be greater than or equal to minimum replica (for dynamic scaling)
  - Target replica must be at least 1 (for scheduled scaling)
- **Migration Requirement**: You must remove `replica` before adding `replica_settings` (requires two separate Terraform operations)

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
  cu_size      = 4
  replica      = 2  # Fixed at 2 replicas
}
```

### Step 2: Enable Replica Autoscaling Configuration

Now update your configuration to add the `replica_settings` block with your desired autoscaling parameters.

**Important:** Make sure to remove the `replica` attribute from your configuration.

```hcl
resource "zillizcloud_cluster" "example" {
  cluster_name = "my-production-cluster"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  cu_size      = 4

  replica_settings = {
    dynamic_scaling = {
      min = 1  # Minimum replicas during low activity
      max = 4  # Maximum replicas during peak load
    }
  }
}
```

Apply the autoscaling configuration:

```bash
terraform plan
terraform apply
```

**Success!** Your cluster now has dynamic replica autoscaling enabled and will automatically adjust replicas between 1 and 4 based on workload.

## Configuring Scheduled Scaling

Scheduled scaling allows you to scale your cluster's replicas at specific times using cron expressions. This is ideal for predictable workload patterns.

### Schedule Scaling Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `timezone` | String | No | The timezone for the cron expression. Defaults to `Etc/UTC`. |
| `cron` | String | Yes | Cron expression defining when the scheduled scaling should occur. |
| `target` | Int64 | Yes | Target number of replicas for the scheduled scaling. Must be at least 1. |

### Cron Expression Format

The cron expression follows the standard format: `minute hour day-of-month month day-of-week`

| Field | Values |
|-------|--------|
| Minute | 0-59 |
| Hour | 0-23 |
| Day of Month | 1-31 |
| Month | 1-12 |
| Day of Week | 0-6 (Sunday = 0) |

**Common Examples:**
- `0 9 * * 1-5` - 9:00 AM on weekdays
- `0 18 * * *` - 6:00 PM every day
- `0 0 * * 0` - Midnight on Sundays
- `30 8 * * 1` - 8:30 AM every Monday

### Example: Scale Up During Business Hours

This configuration scales the cluster to 4 replicas at 9:00 AM and scales down to 1 replica at 6:00 PM on weekdays (US Eastern time):

```hcl
resource "zillizcloud_cluster" "example" {
  cluster_name = "my-production-cluster"
  region_id    = "aws-us-east-1"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  cu_size      = 4

  replica_settings = {
    schedule_scaling = [
      {
        timezone = "America/New_York"
        cron     = "0 9 * * 1-5"  # 9:00 AM on weekdays
        target   = 4              # Scale up to 4 replicas
      },
      {
        timezone = "America/New_York"
        cron     = "0 18 * * 1-5" # 6:00 PM on weekdays
        target   = 1              # Scale down to 1 replica
      }
    ]
  }
}
```

### Example: Weekend Scaling

Scale down during weekends to reduce costs:

```hcl
resource "zillizcloud_cluster" "example" {
  cluster_name = "my-production-cluster"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  cu_size      = 4

  replica_settings = {
    schedule_scaling = [
      {
        timezone = "Etc/UTC"
        cron     = "0 0 * * 6"    # Saturday midnight
        target   = 1              # Minimal replicas for weekend
      },
      {
        timezone = "Etc/UTC"
        cron     = "0 6 * * 1"    # Monday 6:00 AM
        target   = 3              # Scale up for the work week
      }
    ]
  }
}
```

## Combining Dynamic and Scheduled Scaling

For maximum flexibility, you can combine both dynamic autoscaling and scheduled scaling. This is useful when you have predictable baseline patterns but also experience unpredictable spikes.

### Example: Business Hours with Dynamic Scaling

```hcl
resource "zillizcloud_cluster" "example" {
  cluster_name = "my-production-cluster"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  cu_size      = 4

  replica_settings = {
    # Dynamic scaling handles unpredictable load variations
    dynamic_scaling = {
      min = 1   # Never go below 1 replica
      max = 6   # Can scale up to 6 replicas if needed
    }

    # Scheduled scaling sets baseline capacity
    schedule_scaling = [
      {
        timezone = "America/Los_Angeles"
        cron     = "0 8 * * 1-5"  # 8:00 AM weekdays
        target   = 3              # Start business hours with 3 replicas
      },
      {
        timezone = "America/Los_Angeles"
        cron     = "0 20 * * 1-5" # 8:00 PM weekdays
        target   = 1              # Scale down after hours
      },
      {
        timezone = "America/Los_Angeles"
        cron     = "0 0 * * 0,6"  # Midnight on weekends
        target   = 1              # Minimal weekend capacity
      }
    ]
  }
}
```

## Common Errors and Troubleshooting

### Error: "These attributes cannot be configured together"

**Cause**: You attempted to set both `replica` and `replica_settings` simultaneously.

**Solution**: Follow the two-step migration process outlined above. Remove `replica` first, apply the change, then add `replica_settings`.

### Error: "Invalid autoscaling configuration: Minimum replica must be less than or equal to maximum replica"

**Cause**: You set `min` greater than `max`.

**Solution**: Ensure `min` <= `max` in your configuration:

```hcl
replica_settings = {
  dynamic_scaling = {
    min = 1   # Must be <= max
    max = 4
  }
}
```

### Error: Validation failed on minimum replica

**Cause**: Minimum replica value is less than 1.

**Solution**: Set `min` to at least 1:

```hcl
replica_settings = {
  dynamic_scaling = {
    min = 1   # Minimum allowed value
    max = 4
  }
}
```

### Error: Invalid cron expression

**Cause**: The cron expression format is incorrect.

**Solution**: Ensure your cron expression follows the standard 5-field format: `minute hour day-of-month month day-of-week`

```hcl
replica_settings = {
  schedule_scaling = [
    {
      cron   = "0 9 * * 1-5"  # Correct: 5 fields
      target = 3
    }
  ]
}
```

### Error: Invalid timezone

**Cause**: The timezone string is not recognized.

**Solution**: Use a valid IANA timezone name:

```hcl
replica_settings = {
  schedule_scaling = [
    {
      timezone = "America/New_York"  # Valid IANA timezone
      cron     = "0 9 * * 1-5"
      target   = 3
    }
  ]
}
```

**Common Valid Timezones:**
- `Etc/UTC` (default)
- `America/New_York`
- `America/Los_Angeles`
- `Europe/London`
- `Europe/Paris`
- `Asia/Tokyo`
- `Asia/Shanghai`

### Error: Target replica validation failed

**Cause**: The `target` value in scheduled scaling is less than 1.

**Solution**: Ensure the target is at least 1:

```hcl
replica_settings = {
  schedule_scaling = [
    {
      cron   = "0 9 * * *"
      target = 1  # Minimum allowed value
    }
  ]
}
```

## Best Practices

### Scheduled Scaling Tips

1. **Use consistent timezones**: When defining multiple schedules, use the same timezone to avoid confusion
2. **Plan for transitions**: Ensure scale-up schedules trigger before peak load, not during it
3. **Account for timezone changes**: Be aware of daylight saving time if using timezones that observe it
4. **Start conservative**: Begin with wider time windows and adjust based on observed behavior
5. **Test schedules in non-production first**: Validate your cron expressions work as expected

### Combining Dynamic and Scheduled Scaling Tips

1. **Set dynamic scaling bounds appropriately**: When combining, ensure `dynamic_scaling.min` allows scheduled scale-downs
2. **Use scheduled scaling for baselines**: Let dynamic scaling handle unexpected spikes
3. **Monitor costs**: Combined scaling can lead to unexpected resource usage if not carefully configured

### High Availability Considerations

1. **Minimum replicas for HA**: Consider keeping `min` at 2 or higher for production workloads requiring high availability
2. **Geographic distribution**: Replicas help with availability but consider your region's fault tolerance requirements
3. **Balance cost vs. availability**: More replicas increase both cost and availability - find the right balance for your workload

### Monitoring Recommendations

- Set up alerts for scaling events in the Zilliz Cloud console
- Review scaling patterns weekly during initial rollout
- Track query latency and throughput to ensure performance targets are met
- Monitor replica count changes to understand scaling behavior

## Reverting to Fixed Replica

If you need to revert from autoscaling to a fixed replica count:

### Step 1: Remove `replica_settings`

```hcl
resource "zillizcloud_cluster" "example" {
  cluster_name = "my-production-cluster"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  cu_size      = 4
  # replica_settings removed
}
```

Apply the change:

```bash
terraform apply
```

### Step 2: Add Fixed `replica`

```hcl
resource "zillizcloud_cluster" "example" {
  cluster_name = "my-production-cluster"
  region_id    = "aws-us-west-2"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id
  cu_size      = 4
  replica      = 2  # Return to fixed replica count
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
- Learn about [CU Autoscaling](./enable-cluster-autoscaling.md) to scale compute units
- Learn about [Scaling Clusters](./scale-cluster.md) for more cluster management techniques
- Explore [Creating Enterprise Clusters](./create-a-standard-cluster.md) for best practices

## Additional Resources

- [Zilliz Cloud Documentation](https://docs.zilliz.com/)
- [Terraform Provider Registry](https://registry.terraform.io/providers/zillizcloud/zillizcloud/latest)
