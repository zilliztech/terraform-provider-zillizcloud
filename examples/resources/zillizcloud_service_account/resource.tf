# Organization account with project access.
resource "zillizcloud_service_account" "ingestion" {
  name        = "ingestion-service"
  description = "Production ingestion"
  project_assignments = {
    "proj-a" = ["r-project-writer"]
    "proj-b" = []
  }
}

# Organization account with one predefined organization role.
resource "zillizcloud_service_account" "organization" {
  name        = "billing-automation"
  org_role_id = "r-billing-admin"
}

# Project-local account. Its owning project cannot be changed in place.
resource "zillizcloud_service_account" "project" {
  name             = "project-automation"
  project_scope    = "proj-a"
  project_role_ids = ["r-project-writer"]
}
