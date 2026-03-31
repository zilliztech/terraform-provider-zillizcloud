resource "zillizcloud_api_key" "metrics_reader" {
  name = "metrics-reader"
  role = "Member"

  project_access {
    project_id  = zillizcloud_project.prod.id
    role        = "Project Member"
    all_cluster = true
    all_stage   = true
  }
}

output "api_key_value" {
  value     = zillizcloud_api_key.metrics_reader.key_value
  sensitive = true
}
