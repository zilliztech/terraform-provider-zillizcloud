# Import an existing restored cluster without starting a new restore.
# Use target_cluster_id,source_cluster_id,backup_id for a configuration using restore defaults.
terraform import zillizcloud_cluster_restore.from_backup 'in01-target,in01-source,backup-abcdefghijklmnop'

# Append job_id,collection_status,version_or_policy when the configuration uses those settings.
# Leave an optional position empty to skip it. The collection status defaults to RELEASE.
# version_or_policy accepts a series and channel separated by a colon, or ORIGINAL or LATEST.
# Use the values from the original restore and verify that the next plan has no replacement.
terraform import zillizcloud_cluster_restore.from_backup 'in01-target,in01-source,backup-abcdefghijklmnop,job-abcdefghijklmnop,RELEASE,2.6.x:ORIGIN'

# Import reads the existing cluster. The credentials from the original restore are unavailable.
# If a restore failed and the target still exists, inspect it before deleting it outside Terraform.
# A later apply can retry after the failed target has been removed.
