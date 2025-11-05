terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}

# Create a new standard project
resource "zillizcloud_project" "example" {
  project_name = "project-099"
  plan         = "Standard"
}

# You can reference the project ID in other resources
output "project_id" {
  value = zillizcloud_project.example.id
}
