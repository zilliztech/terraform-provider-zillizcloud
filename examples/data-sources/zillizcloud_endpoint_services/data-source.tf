data "zillizcloud_endpoint_services" "this" {
  region_id = "aws-us-west-2"
}

output "available_services" {
  value = data.zillizcloud_endpoint_services.this.endpoint_services[0].endpoint_service
}
