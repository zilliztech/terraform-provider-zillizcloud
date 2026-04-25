# NOTE: Destroy is a no-op - the whitelist entry is not removed from Zilliz Cloud
# on `terraform destroy`. Manual cleanup via console or support is required.
resource "zillizcloud_endpoint_whitelist" "azure" {
  project_id    = "proj-xxxxxxxxxxxxxxxxxxxxxxxx"
  region_id     = "azure-eastus2"
  outer_user_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
