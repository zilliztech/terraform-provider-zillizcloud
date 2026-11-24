data "zillizcloud_service_accounts" "project" {
  project_id = "proj-a"
  status     = "active"
  name       = "ingestion"
}
