# AWS PrivateLink Guide — Design

**Date:** 2026-04-25
**Status:** Approved, ready for implementation plan

## Goal

Publish a registry-visible guide that walks a user end-to-end through provisioning a private, in-VPC connection from AWS to a Zilliz Cloud Enterprise cluster using AWS PrivateLink, with all DNS wiring handled in Terraform.

The guide will appear at `registry.terraform.io/providers/zilliztech/zillizcloud/latest/docs/guides/aws-privatelink` after the next provider release.

## Why

The example under `examples/guides/aws-privatelink/` is currently runnable but undocumented. Users hitting PrivateLink for the first time need conceptual framing (what each piece does and why), and `examples/*` files are not published to the Terraform Registry — only `templates/guides/*.md` (rendered into `docs/guides/*.md` by `tfplugindocs`) are. Without a guide template, this configuration is effectively invisible to registry visitors.

The guide also tightens the example: the previously hardcoded `zilliz_service_name` variable is replaced with a `data "zillizcloud_endpoint_services"` lookup so users don't need to source the `vpce-svc-…` string manually.

## Scope

### In scope

1. **New file:** `templates/guides/aws-privatelink.md` — the guide page.
2. **Modified:** `examples/guides/aws-privatelink/main.tf` — richer per-resource comments explaining the role of each block in the PrivateLink flow. (The data-source-based service-name lookup is already in place.)
3. **Modified:** `examples/guides/aws-privatelink/variables.tf` — remove the now-orphaned `zilliz_service_name` variable.
4. **Modified:** `templates/guides/README.md` — add a discoverable link to the new guide so it surfaces in the registry's Guides index.
5. **Generated:** `docs/guides/aws-privatelink.md` — produced by `make doc`, committed alongside the template (matches existing repo convention).

### Out of scope

- No screenshots or `images/` additions — text-only, matching most existing guides.
- No changes to `outputs.tf`, `versions.tf`, or `terraform.tfvars.example`.
- No acceptance tests for the example.
- No changes to the underlying provider Go code (`zillizcloud_endpoint`, `zillizcloud_endpoint_services`, `zillizcloud_cluster`, etc.).

## Audience and tone

- **Audience:** infrastructure engineers familiar with Terraform and AWS networking primitives (VPC, subnets, security groups, Route 53). Not assumed to know Zilliz Cloud's PrivateLink mechanics.
- **Tone:** conceptual + walkthrough, matching `templates/guides/create-a-standard-cluster.md` and `templates/guides/prepare-resources-for-aws-byoc.md`. Short conceptual framing up front, then numbered steps with HCL snippets, then apply / verify / destroy.

## Guide structure

The guide is organized as follows. Each step embeds a focused HCL snippet inline (not a single `tffile` directive) so explanation sits next to the relevant code.

1. **Title + intro paragraph** — what AWS PrivateLink does in this context; what the guide produces (private, in-VPC connectivity to a Zilliz Enterprise cluster with no public-internet path).
2. **You'll learn how to** — 4–5 bullets summarizing outcomes.
3. **Architecture** — short prose plus an ASCII/box diagram:
   - Client (in VPC) → Interface VPC Endpoint (ENI) → AWS PrivateLink → Zilliz Cloud Enterprise cluster.
   - Side note showing Route 53 private hosted zone resolving the Zilliz-issued hostname to the endpoint's ENI.
4. **Prerequisites** — bulleted checklist:
   - Existing VPC, subnets (one per AZ), security group allowing inbound 443.
   - AWS credentials, Zilliz API key, Zilliz project ID.
   - The Zilliz cluster region must match the AWS region.
   - Link to `get-start.md` for provider setup.
5. **Step 1 — Discover the PrivateLink service name** — `data "zillizcloud_endpoint_services"`; explain region matching and why this beats hardcoding.
6. **Step 2 — Create the AWS interface endpoint** — `aws_vpc_endpoint`; explain `private_dns_enabled = false` (Zilliz issues its own hostname; resolution is handled by Route 53 in Step 5).
7. **Step 3 — Create the Zilliz Enterprise cluster** — `zillizcloud_cluster` with `plan = "Enterprise"`; note that PrivateLink requires Enterprise; introduce `private_link_address` as the attribute consumed later.
8. **Step 4 — Register the endpoint with Zilliz** — `zillizcloud_endpoint`; explain that Zilliz must authorize a specific AWS VPCE id to consume the service.
9. **Step 5 — Wire DNS via Route 53 private hosted zone** — the `locals` extracting the hostname from `private_link_address`, `aws_route53_zone`, and the alias `A` record pointing to the VPCE DNS entry. Explain why a private zone is required given we disabled PrivateDNS on the endpoint.
10. **Apply** — `terraform init` + `terraform apply`; note Terraform handles ordering automatically.
11. **Verifying connectivity** — from an EC2 instance inside the VPC: `dig <private_link_host>` should resolve to a private ENI IP; connect with a Milvus client using `zillizcloud_cluster.*.private_link_address` and the cluster credentials.
12. **Destroying** — `terraform destroy`; note that Zilliz endpoint deregistration happens before VPCE removal in dependency order.
13. **Next steps** — link back to scaling, import, etc.

## Example file changes (detail)

### `examples/guides/aws-privatelink/main.tf`

Each block gets a 2–4 line comment explaining *what it does* and *why it's needed in the PrivateLink flow*. Targets:

- `data "zillizcloud_endpoint_services"` — what it returns and why we use index `[0]` (the first service for the region).
- `aws_vpc_endpoint` — the role of an Interface endpoint; the meaning of `private_dns_enabled = false` and the consequence (must wire DNS ourselves).
- `zillizcloud_endpoint` — the authorization step on the Zilliz side; pairs the VPCE id with the project/region.
- `zillizcloud_cluster` — Enterprise-plan requirement; mention `private_link_address` as the source-of-truth hostname.
- `locals.private_link_host` — why the URL is parsed (`private_link_address` is a full `https://host:port`; Route 53 needs the bare hostname).
- `aws_route53_zone` — why a private hosted zone is needed (PrivateDNS is off on the endpoint; Zilliz issues a public-DNS-shaped hostname that must be overridden in the VPC).
- `aws_route53_record` — why an alias to the first VPCE DNS entry; what `evaluate_target_health = false` implies.

### `examples/guides/aws-privatelink/variables.tf`

Remove `variable "zilliz_service_name"` (orphaned by the data-source lookup). Keep the rest unchanged.

## README index

`templates/guides/README.md` gets one new bullet linking to `./aws-privatelink.md`. Placement: appended to the existing flow, with a short framing phrase like "Connecting privately to Zilliz Cloud."

## Generation step

After authoring the template and example edits:

```
make doc
```

This regenerates `docs/guides/aws-privatelink.md`. Both `templates/guides/aws-privatelink.md` and `docs/guides/aws-privatelink.md` are committed (existing repo convention — every other guide is committed both ways).

## Risks / open questions

- **`make doc` output drift:** if `tfplugindocs` produces unexpected formatting differences vs. the template, address inline at implementation time (no design impact).
- **Verification step realism:** the "from an EC2 instance" verification is described but not scripted. Acceptable — matches the level of other guides which describe outcomes without bundling verification harnesses.
- **Enterprise plan availability:** the guide states PrivateLink requires Enterprise. If this constraint changes, the guide needs an update — but that's a future concern, not a present blocker.

## Success criteria

- `make doc` completes cleanly; `docs/guides/aws-privatelink.md` exists and renders correctly.
- A reader who knows Terraform and AWS basics can follow the guide top-to-bottom and end up with a working PrivateLink connection to a Zilliz Enterprise cluster.
- The new guide is reachable from `templates/guides/README.md` and will appear in the registry's Guides index after release.
- The example under `examples/guides/aws-privatelink/` runs unchanged (modulo the removed `zilliz_service_name` variable, which had a hardcoded default and is now sourced from the data source).
