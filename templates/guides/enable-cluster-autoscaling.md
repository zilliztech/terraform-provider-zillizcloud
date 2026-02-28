# Migrating from Fixed CU Size to CU Autoscaling

This guide walks you through migrating your Zilliz Cloud cluster from a fixed compute unit (CU) size to CU autoscaling. Zilliz Cloud supports two types of autoscaling:

- **Dynamic Autoscaling**: Automatically adjusts compute resources based on real-time workload demands
- **Scheduled Scaling**: Scales compute resources at predetermined times using cron expressions

Note: either `cu_settings` or `replica_settings` can be used, but not both.


## Prerequisites

Before you begin, ensure you have:

- Completed the initial setup steps outlined in the [Getting Started with Zilliz Cloud Terraform Provider](./get-start.md) guide
- An existing cluster configured with `cu_size`
- Necessary permissions and access credentials to interact with the Zilliz Cloud API
- An Enterprise plan cluster (autoscaling is available on Enterprise clusters)

## Understanding CU Autoscaling

### Fixed CU Size vs. Autoscaling Options

**Fixed CU Size (`cu_size`)**:
- Static allocation of compute units
- Cluster always uses the same amount of resources
- Simple to configure but may be over-provisioned during low usage or under-provisioned during peak times

**Dynamic Autoscaling (`cu_settings.dynamic_scaling`)**:
- Automatically scales compute units between a minimum and maximum threshold
- Adjusts resources based on actual workload demands
- Optimizes costs by using fewer resources during low activity periods
- Ensures performance by scaling up during high demand

**Scheduled Scaling (`cu_settings.schedule_scaling`)**:
- Scales compute units at specific times defined by cron expressions
- Ideal for predictable traffic patterns (e.g., business hours, batch processing windows)
- Supports timezone configuration for global deployments
- Can define multiple schedules for different scaling targets

### When to Use Each Approach

| Scenario | Recommended Approach |
|----------|---------------------|
| Unpredictable workloads | Dynamic Autoscaling |
| Predictable daily/weekly patterns | Scheduled Scaling |
| Simple, consistent workloads | Fixed CU Size |

### Important Constraints

- **Mutual Exclusivity**: `cu_size` and `cu_settings` cannot be configured together on the same cluster
- **Validation Rules**:
  - Minimum CU must be at least 1 (for dynamic scaling)
  - Maximum CU must be greater than or equal to minimum CU (for dynamic scaling)
  - Target CU must be at least 1 (for scheduled scaling)
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

## Configuring Scheduled Scaling

Scheduled scaling allows you to scale your cluster's compute units at specific times using cron expressions. This is ideal for predictable workload patterns.

### Schedule Scaling Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `timezone` | String | No | The timezone for the cron expression. Defaults to `Etc/UTC`. |
| `cron` | String | Yes | Cron expression defining when the scheduled scaling should occur. |
| `target` | Int64 | Yes | Target number of compute units (CU) for the scheduled scaling. Must be at least 1. |

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

This configuration scales the cluster to 8 CUs at 9:00 AM and scales down to 2 CUs at 6:00 PM on weekdays (US Eastern time):

```hcl
resource "zillizcloud_cluster" "example" {
  cluster_name = "my-production-cluster"
  region_id    = "aws-us-east-1"
  plan         = "Enterprise"
  cu_type      = "Performance-optimized"
  project_id   = data.zillizcloud_project.default.id

  cu_settings = {
    schedule_scaling = [
      {
        timezone = "America/New_York"
        cron     = "0 9 * * 1-5"  # 9:00 AM on weekdays
        target   = 8              # Scale up to 8 CUs
      },
      {
        timezone = "America/New_York"
        cron     = "0 18 * * 1-5" # 6:00 PM on weekdays
        target   = 2              # Scale down to 2 CUs
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

  cu_settings = {
    schedule_scaling = [
      {
        timezone = "Etc/UTC"
        cron     = "0 0 * * 6"    # Saturday midnight
        target   = 1              # Minimal CUs for weekend
      },
      {
        timezone = "Etc/UTC"
        cron     = "0 6 * * 1"    # Monday 6:00 AM
        target   = 4              # Scale up for the work week
      }
    ]
  }
}
```


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

### Error: Invalid cron expression

**Cause**: The cron expression format is incorrect.

**Solution**: Ensure your cron expression follows the standard 5-field format: `minute hour day-of-month month day-of-week`

```hcl
cu_settings = {
  schedule_scaling = [
    {
      cron   = "0 9 * * 1-5"  # Correct: 5 fields
      target = 4
    }
  ]
}
```

### Error: Invalid timezone

**Cause**: The timezone string is not recognized.

**Solution**: Use a valid IANA timezone name:

```hcl
cu_settings = {
  schedule_scaling = [
    {
      timezone = "America/New_York"  # Valid IANA timezone
      cron     = "0 9 * * 1-5"
      target   = 4
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

### Error: Target CU validation failed

**Cause**: The `target` value in scheduled scaling is less than 1.

**Solution**: Ensure the target is at least 1:

```hcl
cu_settings = {
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

### Monitoring Recommendations

- Set up alerts for scaling events in the Zilliz Cloud console
- Review scaling patterns weekly during initial rollout
- Track query latency and throughput to ensure performance targets are met

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
- Learn about [Replica Autoscaling](./enable-replica-autoscaling.md) to scale cluster replicas independently
- Learn about [Scaling Clusters](./scale-cluster.md) for more cluster management techniques
- Explore [Creating Enterprise Clusters](./create-a-standard-cluster.md) for best practices

## Additional Resources

- [Zilliz Cloud Documentation](https://docs.zilliz.com/)
- [Terraform Provider Registry](https://registry.terraform.io/providers/zillizcloud/zillizcloud/latest)
