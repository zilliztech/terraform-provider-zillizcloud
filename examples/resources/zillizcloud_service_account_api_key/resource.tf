resource "zillizcloud_service_account" "ingestion" {
  name = "ingestion-service"
  project_assignments = {
    "proj-a" = ["r-project-writer"]
  }
}

resource "zillizcloud_service_account_api_key" "ingestion" {
  service_account_id = zillizcloud_service_account.ingestion.id
  name               = "ingestion-key"
  description        = "Primary credential"
  expire_days        = 90

  lifecycle {
    create_before_destroy = true
  }
}

output "ingestion_secret" {
  value     = zillizcloud_service_account_api_key.ingestion.secret
  sensitive = true
}
