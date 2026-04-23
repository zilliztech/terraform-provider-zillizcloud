# AWS PrivateLink Example Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship an end-to-end Terraform example under `examples/guides/aws-privatelink/` that provisions an AWS VPC Interface Endpoint against a Zilliz PrivateLink service and registers it via `zillizcloud_endpoint`.

**Architecture:** Five HCL files in a new guide directory. All infra inputs (VPC id, subnet ids, security group ids, project id, API key) come from variables — the example creates no networking primitives. `aws_vpc_endpoint.id` feeds `zillizcloud_endpoint.endpoint_id` to establish implicit ordering.

**Tech Stack:** Terraform ≥ 1.3, `hashicorp/aws` ~> 5.0, `zilliztech/zillizcloud` (unpinned). Validation via `terraform fmt -check` + `terraform init -backend=false` + `terraform validate`.

**Spec:** `docs/superpowers/specs/2026-04-23-aws-privatelink-example-design.md`

---

## File Structure

```
examples/guides/aws-privatelink/
├── versions.tf              # terraform + provider blocks
├── variables.tf             # all inputs
├── main.tf                  # aws_vpc_endpoint + zillizcloud_endpoint
├── outputs.tf               # endpoint id / state / zilliz resource
└── terraform.tfvars.example # copyable sample values
```

Each file has a single responsibility. No code is repeated across files.

---

### Task 1: Create the guide directory and versions.tf

**Files:**
- Create: `examples/guides/aws-privatelink/versions.tf`

- [ ] **Step 1: Create the directory**

Run: `mkdir -p examples/guides/aws-privatelink`
Expected: no output, directory exists.

- [ ] **Step 2: Write versions.tf**

```hcl
terraform {
  required_version = ">= 1.3"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

provider "zillizcloud" {
  api_key = var.zilliz_api_key
}
```

- [ ] **Step 3: Format check**

Run: `terraform fmt examples/guides/aws-privatelink/versions.tf`
Expected: no output (already formatted) or the filename (reformatted in place).

- [ ] **Step 4: Commit**

```bash
git add examples/guides/aws-privatelink/versions.tf
git commit -m "docs(examples): add versions.tf for aws-privatelink guide"
```

---

### Task 2: Declare all input variables

**Files:**
- Create: `examples/guides/aws-privatelink/variables.tf`

- [ ] **Step 1: Write variables.tf**

```hcl
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

variable "zilliz_service_name" {
  description = "AWS PrivateLink service name exposed by Zilliz Cloud for your cluster."
  type        = string
  default     = "com.amazonaws.vpce.us-west-2.vpce-svc-0ecd272cca7aa6d6e"
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
```

- [ ] **Step 2: Format check**

Run: `terraform fmt examples/guides/aws-privatelink/variables.tf`
Expected: no output or filename.

- [ ] **Step 3: Commit**

```bash
git add examples/guides/aws-privatelink/variables.tf
git commit -m "docs(examples): add variables.tf for aws-privatelink guide"
```

---

### Task 3: Wire the two resources in main.tf

**Files:**
- Create: `examples/guides/aws-privatelink/main.tf`

- [ ] **Step 1: Write main.tf**

```hcl
# End-to-end PrivateLink wiring between AWS and Zilliz Cloud.
#
# Apply order is implicit:
#   1. aws_vpc_endpoint creates the Interface Endpoint in your VPC.
#   2. zillizcloud_endpoint registers that vpce-id with the Zilliz project,
#      which triggers auto-acceptance on the Zilliz side. No manual step
#      in the Zilliz console is required.
#
# private_dns_enabled is false: Zilliz PrivateLink services do not publish
# a private DNS name, so enabling it would fail at apply time. Use the DNS
# entries returned by aws_vpc_endpoint (or layer a Route 53 private zone
# on top) to resolve the cluster hostname from inside the VPC.

resource "aws_vpc_endpoint" "zilliz" {
  vpc_id              = var.vpc_id
  service_name        = var.zilliz_service_name
  vpc_endpoint_type   = "Interface"
  subnet_ids          = var.subnet_ids
  security_group_ids  = var.security_group_ids
  private_dns_enabled = false
}

resource "zillizcloud_endpoint" "this" {
  project_id  = var.zilliz_project_id
  region_id   = var.zilliz_region_id
  endpoint_id = aws_vpc_endpoint.zilliz.id
}
```

- [ ] **Step 2: Format check**

Run: `terraform fmt examples/guides/aws-privatelink/main.tf`
Expected: no output or filename.

- [ ] **Step 3: Commit**

```bash
git add examples/guides/aws-privatelink/main.tf
git commit -m "docs(examples): add main.tf wiring aws_vpc_endpoint into zillizcloud_endpoint"
```

---

### Task 4: Surface outputs

**Files:**
- Create: `examples/guides/aws-privatelink/outputs.tf`

- [ ] **Step 1: Write outputs.tf**

```hcl
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
```

- [ ] **Step 2: Format check**

Run: `terraform fmt examples/guides/aws-privatelink/outputs.tf`
Expected: no output or filename.

- [ ] **Step 3: Commit**

```bash
git add examples/guides/aws-privatelink/outputs.tf
git commit -m "docs(examples): add outputs.tf for aws-privatelink guide"
```

---

### Task 5: Provide a tfvars template

**Files:**
- Create: `examples/guides/aws-privatelink/terraform.tfvars.example`

- [ ] **Step 1: Write the template**

```hcl
aws_region = "us-west-2"

vpc_id     = "vpc-0123456789abcdef0"
subnet_ids = [
  "subnet-0123456789abcdef0",
  "subnet-0fedcba9876543210",
]
security_group_ids = [
  "sg-0123456789abcdef0",
]

zilliz_service_name = "com.amazonaws.vpce.us-west-2.vpce-svc-0ecd272cca7aa6d6e"
zilliz_project_id   = "proj-ebc5ac7f430702aec8c57b"
zilliz_region_id    = "aws-us-west-2"

# Prefer exporting TF_VAR_zilliz_api_key in your shell instead of committing this.
zilliz_api_key = "REPLACE_ME"
```

- [ ] **Step 2: Commit**

```bash
git add examples/guides/aws-privatelink/terraform.tfvars.example
git commit -m "docs(examples): add tfvars template for aws-privatelink guide"
```

---

### Task 6: Validate the whole example

**Files:**
- No file changes. Pure verification.

- [ ] **Step 1: Format check across the directory**

Run: `terraform -chdir=examples/guides/aws-privatelink fmt -check -recursive`
Expected: exit code 0, no filenames printed.

- [ ] **Step 2: Init without backend (pulls providers)**

Run: `terraform -chdir=examples/guides/aws-privatelink init -backend=false`
Expected: "Terraform has been successfully initialized!" and both `aws` and `zillizcloud` providers resolved.

- [ ] **Step 3: Validate**

Run: `terraform -chdir=examples/guides/aws-privatelink validate`
Expected: "Success! The configuration is valid."

- [ ] **Step 4: Clean up init artifacts (do not commit them)**

Run: `rm -rf examples/guides/aws-privatelink/.terraform examples/guides/aws-privatelink/.terraform.lock.hcl`
Expected: no output.

- [ ] **Step 5: Confirm clean tree**

Run: `git status examples/guides/aws-privatelink`
Expected: "nothing to commit, working tree clean".

No commit — this task is verification only.
