data "zillizcloud_private_endpoints" "mine" {
  project_id = "proj-xxxxxxxxxxxxxxxxxxxxxxxx"
}

output "endpoints" {
  value = data.zillizcloud_private_endpoints.mine.endpoints
}
