data "zillizcloud_cluster_versions" "create" {
  region_id = "aws-us-east-2"
}

data "zillizcloud_cluster_versions" "restore" {
  backup_id = "backup-xxxxxxxxxxxxxxxx"
}

output "creation_candidates" {
  value = data.zillizcloud_cluster_versions.create.regions
}

output "restore_candidates" {
  value = data.zillizcloud_cluster_versions.restore.regions
}
