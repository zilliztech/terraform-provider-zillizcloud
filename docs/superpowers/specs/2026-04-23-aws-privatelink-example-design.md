# AWS PrivateLink End-to-End Example — Design

## Goal

Provide a complete, runnable Terraform example that provisions an AWS VPC
Interface Endpoint against a Zilliz Cloud PrivateLink service and wires the
resulting endpoint id into `zillizcloud_endpoint`.

The existing minimal example at
`examples/resources/zillizcloud_endpoint/resource.tf` (consumed by
tfplugindocs) is left unchanged. This new example lives under
`examples/guides/` so it will not pollute the auto-generated resource docs.

## Directory Layout

```
examples/guides/aws-privatelink/
├── main.tf                   # aws_vpc_endpoint + zillizcloud_endpoint
├── variables.tf              # all external inputs
├── outputs.tf                # exposes endpoint id / state
├── versions.tf               # provider version constraints + provider blocks
└── terraform.tfvars.example  # sample values users copy to terraform.tfvars
```

## Inputs (variables.tf)

All inputs are external — the example does not create VPC, subnets, or
security groups.

| Variable | Type | Default | Notes |
|---|---|---|---|
| `aws_region` | string | — | e.g. `us-west-2` |
| `vpc_id` | string | — | Existing VPC |
| `subnet_ids` | list(string) | — | Subnets for the Interface Endpoint |
| `security_group_ids` | list(string) | — | SGs attached to the endpoint |
| `zilliz_service_name` | string | `com.amazonaws.vpce.us-west-2.vpce-svc-0ecd272cca7aa6d6e` | PrivateLink service name |
| `zilliz_project_id` | string | — | Zilliz project id |
| `zilliz_region_id` | string | `aws-us-west-2` | Zilliz region |
| `zilliz_api_key` | string (sensitive) | — | Zilliz API key |

## Core Orchestration (main.tf)

```hcl
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

Key points:

- `zillizcloud_endpoint.endpoint_id` references `aws_vpc_endpoint.zilliz.id`,
  creating an implicit dependency so Terraform provisions the AWS endpoint
  first and then registers it with Zilliz.
- `private_dns_enabled = false` — third-party PrivateLink services typically
  do not support private DNS; enabling it would fail at apply time.
- No VPC / Subnet / SG are created by this example.

## Providers (versions.tf)

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

The `zillizcloud` provider version is intentionally unpinned so users pick up
the latest release that ships this resource.

## Outputs (outputs.tf)

- `aws_vpc_endpoint_id` — the `vpce-…` id, useful for debugging / console lookup.
- `aws_vpc_endpoint_state` — so users can see whether the AWS side is
  `available` vs `pendingAcceptance`.
- `zilliz_endpoint` — the `zillizcloud_endpoint` resource output, for
  downstream modules that need the Zilliz-side binding.

## terraform.tfvars.example

Ships a copyable template with placeholder values for every required
variable, including a sample `zilliz_api_key = "…"` line marked as sensitive
so users immediately see what they must fill in.

## Notes Included in main.tf Comments

A short comment block at the top of `main.tf` explains:

1. The implicit ordering (AWS endpoint → Zilliz registration).
2. Why `private_dns_enabled` is `false`.
3. That the PrivateLink service auto-accepts connections once the matching
   `zillizcloud_endpoint` resource is registered — no manual acceptance
   needed in the Zilliz console.

## Out of Scope

- Creating VPC, subnets, or security groups.
- Route 53 private hosted zones for custom DNS (users who want friendly
  hostnames can layer this on top).
- Multi-region / multi-AZ redundancy beyond what `subnet_ids` already
  supports.
