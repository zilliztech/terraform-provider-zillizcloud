# AWS PrivateLink → Zilliz Cloud verification — design

Date: 2026-04-25
Scope: extend `examples/guides/aws-privatelink/` with a working end-to-end
verification path. Prerequisite AWS networking is created imperatively via
`aws` CLI (4-account profile, us-west-2); Terraform code in `main.tf` is
unchanged.

## Stages

### Stage 1 — Prerequisite AWS networking (CLI)

Disposable VPC dedicated to this guide:

- VPC `10.20.0.0/16`, DNS support + DNS hostnames enabled.
- Internet Gateway attached to the VPC.
- 2 subnets: `10.20.1.0/24` in `us-west-2a`, `10.20.2.0/24` in `us-west-2b`.
- Route table with `0.0.0.0/0 → IGW`, associated with both subnets.
- Security group with self-referencing ingress on TCP/443 and default egress.
- All resources tagged `Project=zilliz-privatelink-guide` for teardown.

Resulting `vpc_id`, `subnet_ids`, `security_group_ids` written into
`terraform.tfvars`.

### Stage 2 — Terraform apply

`main.tf` is unchanged. Apply creates:

- `aws_vpc_endpoint.zilliz` — Interface Endpoint for the Zilliz VPCE service.
- `zillizcloud_endpoint.this` — registers the vpce-id (auto-accepted).
- `zillizcloud_cluster.enterprise_plan_cluster` — Enterprise plan, CU=1.
- `aws_route53_zone.private` + alias A record → VPCE DNS.

### Stage 3 — Verification bastion

- IAM role `zilliz-privatelink-bastion-ssm` with `AmazonSSMManagedInstanceCore`,
  plus matching instance profile.
- t3.micro Amazon Linux 2023 in subnet A, public IP enabled, attached to the
  guide SG. SSM agent reaches the SSM service over IGW; the VPCE is reached
  via its private IP on the same SG.

### Stage 4 — Verification

Inside `aws ssm start-session`:

1. `dig <private_link_address>` → expect a `10.20.x.x` answer.
2. `curl -v https://<private_link_address>` → expect TLS handshake completion.

### Stage 5 — Teardown

`terraform destroy` → terminate bastion → delete SG, subnets, route table,
IGW, VPC → delete IAM instance profile + role.

## Out of scope

- Productionizing into reusable Terraform modules.
- Auth/load testing beyond a TLS handshake.
- Private DNS via SSM endpoints (cheaper public-subnet path used instead).
