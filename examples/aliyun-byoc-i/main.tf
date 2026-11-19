# Use a provider build containing the Aliyun BYOC-I registration resource.
# The control-api, cloud-meta and activity-provider must support aliParam/aliConfig.
terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

# Configure the API key and China control-plane address through provider settings.
# Inputs below are outputs from the customer-owned ACK/RAM/OSS infrastructure.
variable "project_id" { type = string }
variable "data_plane_id" { type = string }
variable "ali" {
  type = object({
    region               = string
    account_id           = string
    ack_cluster_id       = string
    vpc_id               = string
    bucket_id            = string
    node_role_arn        = string
    maintenance_role_arn = string
    storage_role_arn     = string
    oidc_provider_arn    = string
    vswitch_ids          = set(string)
    security_group_ids   = set(string)
    zones                = set(string)
    endpoint_id          = optional(string)
  })
}

# Apply after settings has issued a data plane ID and the agent is CONNECTED.
# Registration returns a deployment job; wait for RUNNING before creating clusters.
resource "zillizcloud_byoc_i_project" "this" {
  project_id    = var.project_id
  data_plane_id = var.data_plane_id
  ali           = var.ali
}
