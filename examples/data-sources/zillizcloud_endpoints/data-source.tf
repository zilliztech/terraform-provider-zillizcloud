data "zillizcloud_endpoints" "mine" {
  project_id = "proj-xxxxxxxxxxxxxxxxxxxxxxxx"
}

output "endpoints" {
  value = data.zillizcloud_endpoints.mine.endpoints
}
