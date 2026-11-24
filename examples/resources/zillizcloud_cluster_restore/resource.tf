resource "zillizcloud_cluster_restore" "from_backup" {
  source_cluster_id = "in01-xxxxxxxxxxxxxxxx"
  backup_id         = "backup-xxxxxxxxxxxxxxxx"
  project_id        = "proj-xxxxxxxxxxxxxxxx"
  cluster_name      = "restored-cluster"
  cu_size           = 1

  # Choose a version returned by the restore candidates query for this backup.
  version = {
    milvus_version = "2.6.x"
    channel        = "ORIGIN"
  }

  timeouts {
    create = "60m"
  }
}

output "restore_status" {
  value = zillizcloud_cluster_restore.from_backup.restore_status
}
