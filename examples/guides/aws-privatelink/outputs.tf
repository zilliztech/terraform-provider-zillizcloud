output "aws_vpc_endpoint_id" {
  description = "The vpce-… ID of the AWS Interface Endpoint."
  value       = aws_vpc_endpoint.zilliz.id
}

output "aws_vpc_endpoint_state" {
  description = "Current state of the AWS Interface Endpoint (e.g. available, pendingAcceptance)."
  value       = aws_vpc_endpoint.zilliz.state
}

output "aws_vpc_endpoint_dns_entries" {
  description = "DNS names AWS assigns to the Interface Endpoint. Use these (or a Route 53 private zone) to reach the Zilliz cluster."
  value       = aws_vpc_endpoint.zilliz.dns_entry
}

output "zilliz_endpoint" {
  description = "The zillizcloud_endpoint resource, including its server-side state."
  value       = zillizcloud_endpoint.this
}
