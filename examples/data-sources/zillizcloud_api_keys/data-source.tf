data "zillizcloud_api_keys" "all" {}

output "key_names" {
  value = [for k in data.zillizcloud_api_keys.all.api_keys : k.name]
}
