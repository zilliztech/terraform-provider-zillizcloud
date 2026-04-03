terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

# Create an API key with Member role scoped to a specific project
resource "zillizcloud_api_key" "member_key" {
  name = "app-readonly-key"
  role = "Member"

  project_access = [{
    project_id  = "proj-xxxxxxxxxxxxxxxxxxxx"
    role        = "Read-Only"
    all_cluster = true
  }]
}

# Create an API key with Billing-Admin role
resource "zillizcloud_api_key" "billing_key" {
  name = "billing-automation"
  role = "Billing-Admin"
}
}

# The key value is only available at creation time.
# After apply, retrieve it with: terraform output -raw api_key_value
output "api_key_value" {
  value     = zillizcloud_api_key.member_key.key_value
  sensitive = true
}
