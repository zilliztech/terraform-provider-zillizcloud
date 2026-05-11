terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}


// specific project
data "zillizcloud_project" "selected_project" {
  id = "proj-xxxxxxxxxxxxxxx"
}

output "selected_project_details" {
  value = data.zillizcloud_project.selected_project
}
