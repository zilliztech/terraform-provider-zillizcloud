resource "zillizcloud_private_endpoint" "aws" {
  project_id  = "proj-xxxxxxxxxxxxxxxxxxxxxxxx"
  region_id   = "aws-us-west-2"
  endpoint_id = "vpce-072eaf2b4a747c24f"
}

# GCP example — gcp_project_id is required
resource "zillizcloud_private_endpoint" "gcp" {
  project_id     = "proj-xxxxxxxxxxxxxxxxxxxxxxxx"
  region_id      = "gcp-us-west1"
  endpoint_id    = "my-psc-endpoint"
  gcp_project_id = "my-gcp-project"
}
