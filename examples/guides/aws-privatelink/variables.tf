variable "aws_region" {
  description = "AWS region where the VPC endpoint is created (must match the Zilliz cluster region)."
  type        = string
}

variable "vpc_id" {
  description = "ID of the existing VPC that will host the interface endpoint."
  type        = string
}

variable "subnet_ids" {
  description = "Subnet IDs the interface endpoint ENIs are placed in. One per AZ is typical."
  type        = list(string)
}

variable "security_group_ids" {
  description = "Security group IDs attached to the interface endpoint. Must allow inbound 443 from clients."
  type        = list(string)
}

variable "zilliz_project_id" {
  description = "Zilliz Cloud project ID that owns the target cluster."
  type        = string
}

variable "zilliz_region_id" {
  description = "Zilliz Cloud region ID (e.g. aws-us-west-2)."
  type        = string
  default     = "aws-us-west-2"
}

variable "zilliz_api_key" {
  description = "Zilliz Cloud API key."
  type        = string
  sensitive   = true
}
