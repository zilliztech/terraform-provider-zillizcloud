terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

# Create a new project with an Enterprise plan and an initial region binding
resource "zillizcloud_project" "example" {
  project_name = "project-099"
  plan         = "Enterprise"
  region_ids   = ["aws-us-west-2"]
}

# You can reference the project ID in other resources
output "project_id" {
  value = zillizcloud_project.example.id
}
