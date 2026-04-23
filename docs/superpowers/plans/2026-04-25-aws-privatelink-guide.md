# AWS PrivateLink Guide Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish a registry-visible guide that walks users end-to-end through provisioning AWS PrivateLink to a Zilliz Cloud Enterprise cluster, and tighten the existing example with explanatory comments.

**Architecture:** Documentation work only. Three artifacts: a new `templates/guides/aws-privatelink.md` (rendered to the registry by `tfplugindocs`), per-resource comments in the runnable example under `examples/guides/aws-privatelink/`, and an index entry in `templates/guides/README.md`. `make doc` regenerates `docs/guides/aws-privatelink.md` which is committed alongside the template.

**Tech Stack:** Markdown, HCL, `terraform-plugin-docs` (invoked via `make doc` → `go generate ./...`).

**Spec:** `docs/superpowers/specs/2026-04-25-aws-privatelink-guide-design.md`

---

## File map

- **Modify:** `examples/guides/aws-privatelink/main.tf` — add per-resource explanatory comments.
- **Modify:** `examples/guides/aws-privatelink/variables.tf` — remove orphaned `zilliz_service_name` variable.
- **Create:** `templates/guides/aws-privatelink.md` — the new guide.
- **Modify:** `templates/guides/README.md` — add a discoverable bullet to the index.
- **Generated (commit):** `docs/guides/aws-privatelink.md` — produced by `make doc`.

Tasks are sequenced so each commit is reviewable on its own.

---

## Task 1: Annotate `examples/guides/aws-privatelink/main.tf`

**Files:**
- Modify: `examples/guides/aws-privatelink/main.tf`

- [ ] **Step 1: Replace the file contents with the annotated version**

Overwrite the file with this exact content:

```hcl
# End-to-end PrivateLink wiring between AWS and Zilliz Cloud.
#
# This configuration provisions a private, in-VPC connection from an
# existing AWS VPC to a new Zilliz Cloud Enterprise-plan cluster. No
# traffic traverses the public internet.
#
# Pieces, in order of dependency:
#   1. Discover the Zilliz PrivateLink service name for the region.
#   2. Create an AWS Interface VPC Endpoint that consumes that service.
#   3. Create the Zilliz Enterprise cluster (issues a private_link_address).
#   4. Register the AWS VPCE id with Zilliz so the service authorizes it.
#   5. Wire DNS via a Route 53 private hosted zone so the cluster's
#      Zilliz-issued hostname resolves to the VPCE inside the VPC.

# Look up the AWS PrivateLink service name Zilliz exposes in this region.
# We use index [0] because Zilliz publishes a single service per region.
# Sourcing this from a data source avoids hardcoding a vpce-svc-… string
# that is region- and deployment-specific.
data "zillizcloud_endpoint_services" "this" {
  region_id = var.zilliz_region_id
}


# Interface VPC Endpoint: an ENI placed in each subnet that becomes the
# in-VPC entry point to the Zilliz PrivateLink service. private_dns_enabled
# is intentionally false — Zilliz issues its own hostname for the cluster,
# and we override resolution explicitly via Route 53 below. Leaving AWS
# PrivateDNS on would attempt to register service-owned domains we don't
# control.
resource "aws_vpc_endpoint" "zilliz" {
  vpc_id              = var.vpc_id
  service_name        = data.zillizcloud_endpoint_services.this.endpoint_services[0].endpoint_service
  vpc_endpoint_type   = "Interface"
  subnet_ids          = var.subnet_ids
  security_group_ids  = var.security_group_ids
  private_dns_enabled = false
}

# Authorize this specific VPC endpoint on the Zilliz side. The Zilliz
# PrivateLink service won't accept traffic from a VPCE that hasn't been
# registered against the project + region.
resource "zillizcloud_endpoint" "this" {
  project_id  = var.zilliz_project_id
  region_id   = var.zilliz_region_id
  endpoint_id = aws_vpc_endpoint.zilliz.id
}

# The target cluster. PrivateLink is only available on the Enterprise
# plan. After creation, the cluster exposes a `private_link_address`
# attribute (a full https://host:port URL) — that hostname is what we
# resolve via Route 53 in the next block.
resource "zillizcloud_cluster" "enterprise_plan_cluster" {
  cluster_name = "enterprise_plan_cluster"
  region_id    = var.zilliz_region_id
  plan         = "Enterprise"
  cu_size      = 1
  cu_type      = "Performance-optimized"
  project_id   = var.zilliz_project_id
}



# private_link_address is a full URL (https://host:port). Route 53 needs
# the bare hostname, so strip the scheme and trim the port.
locals {
  private_link_host = element(
    split(":", replace(zillizcloud_cluster.enterprise_plan_cluster.private_link_address, "https://", "")),
    0,
  )
}

# Private hosted zone whose name is the Zilliz-issued hostname. Because
# we disabled PrivateDNS on the VPCE, this zone is what makes the
# cluster's hostname resolvable from inside the VPC. The zone is scoped
# to var.vpc_id so it does not leak to other VPCs.
resource "aws_route53_zone" "private" {
  name = local.private_link_host

  vpc {
    vpc_id = var.vpc_id
  }

  comment = "Private hosted zone for Zilliz Cloud private link"
}


# Alias A record -> VPC endpoint. An Interface VPC endpoint exposes one
# or more regional DNS entries; we use the first entry's dns_name and
# hosted_zone_id as the alias target. evaluate_target_health is false
# because Route 53 cannot health-check a private VPCE alias and the
# endpoint's availability is already managed by AWS.
resource "aws_route53_record" "alias" {
  zone_id = aws_route53_zone.private.zone_id
  name    = local.private_link_host
  type    = "A"

  alias {
    name                   = aws_vpc_endpoint.zilliz.dns_entry[0].dns_name
    zone_id                = aws_vpc_endpoint.zilliz.dns_entry[0].hosted_zone_id
    evaluate_target_health = false
  }
}
```

- [ ] **Step 2: Verify the file parses**

Run: `terraform -chdir=examples/guides/aws-privatelink fmt -check`
Expected: exit 0. If `fmt` rewrites whitespace, accept the result and re-stage.

- [ ] **Step 3: Commit**

```bash
git add examples/guides/aws-privatelink/main.tf
git commit -m "$(cat <<'EOF'
docs(examples): annotate aws-privatelink main.tf

Add per-resource comments explaining the role of each block in the
PrivateLink flow (service discovery, interface endpoint, cluster,
endpoint registration, Route 53 private zone wiring).
EOF
)"
```

---

## Task 2: Remove orphaned `zilliz_service_name` variable

**Files:**
- Modify: `examples/guides/aws-privatelink/variables.tf`

- [ ] **Step 1: Delete the variable block**

Use Edit to remove this exact block from `examples/guides/aws-privatelink/variables.tf`:

```hcl
variable "zilliz_service_name" {
  description = "AWS PrivateLink service name exposed by Zilliz Cloud for your cluster."
  type        = string
  default     = "com.amazonaws.vpce.us-west-2.vpce-svc-0ecd272cca7aa6d6e"
}

```

(Leave the surrounding variables — `aws_region`, `vpc_id`, `subnet_ids`, `security_group_ids`, `zilliz_project_id`, `zilliz_region_id`, `zilliz_api_key` — untouched. Remove any leftover blank line so there are not two consecutive blanks.)

- [ ] **Step 2: Confirm no remaining references**

Run: `grep -n zilliz_service_name examples/guides/aws-privatelink/`
Expected: no output (the variable is fully orphaned now that `main.tf` uses the data source).

- [ ] **Step 3: Verify the example still parses**

Run: `terraform -chdir=examples/guides/aws-privatelink fmt -check && terraform -chdir=examples/guides/aws-privatelink validate -no-color || true`
Note: `validate` will require `terraform init` and may fail without credentials; the goal here is HCL parse cleanliness. `fmt -check` exiting 0 is sufficient.

- [ ] **Step 4: Commit**

```bash
git add examples/guides/aws-privatelink/variables.tf
git commit -m "$(cat <<'EOF'
docs(examples): drop zilliz_service_name var from aws-privatelink

Replaced by the zillizcloud_endpoint_services data source lookup, so
the variable is orphaned. Removing it cleans up the example surface.
EOF
)"
```

---

## Task 3: Create `templates/guides/aws-privatelink.md`

**Files:**
- Create: `templates/guides/aws-privatelink.md`

- [ ] **Step 1: Write the guide**

Create the file with this exact content:

````markdown
## Tutorial: Connecting Privately to Zilliz Cloud with AWS PrivateLink

This tutorial walks you end-to-end through provisioning a private, in-VPC connection from AWS to a Zilliz Cloud **Enterprise Plan** cluster using AWS PrivateLink. When you finish, traffic between your application and the cluster stays inside AWS — no public internet path is involved.

The runnable Terraform configuration that backs this guide lives at [`examples/guides/aws-privatelink/`](https://github.com/zilliztech/terraform-provider-zillizcloud/tree/master/examples/guides/aws-privatelink).

**You'll learn how to:**
- Discover the AWS PrivateLink service name Zilliz Cloud exposes for your region.
- Provision an AWS Interface VPC Endpoint that consumes that service.
- Create a Zilliz Cloud Enterprise cluster and register the VPC endpoint with it.
- Wire DNS resolution via a Route 53 private hosted zone so the cluster's hostname resolves to the endpoint inside your VPC.
- Verify the connection and tear it down cleanly.

### Prerequisites

Before you begin, complete the setup in [Getting Started with Zilliz Cloud Terraform Provider](./get-start.md). In addition, ensure you have:

- An existing AWS **VPC**, with one or more **subnets** (one per AZ is typical) and at least one **security group** that allows inbound TCP 443 from the clients that will reach the cluster.
- AWS credentials configured for the Terraform AWS provider.
- A Zilliz Cloud **API key** and a **project ID**. See [How Can I Obtain the Project ID?](https://support.zilliz.com/hc/en-us/articles/22048954409755-How-Can-I-Obtain-the-Project-ID).
- A Zilliz Cloud **region** that matches the AWS region you're deploying into (for example `aws-us-west-2` on the Zilliz side and `us-west-2` on the AWS side).
- The AWS provider configured in your Terraform project alongside the Zilliz Cloud provider.

### Architecture

```
+-------------------- AWS VPC --------------------+
|                                                 |
|   Application                                   |
|       |                                         |
|       | DNS lookup (resolved by Route 53        |
|       |             private hosted zone)        |
|       v                                         |
|   Interface VPC Endpoint  (ENIs in subnets)     |
|       |                                         |
+-------|-----------------------------------------+
        |
        | AWS PrivateLink (private backbone)
        v
   Zilliz Cloud Enterprise cluster
```

The Interface VPC Endpoint is the in-VPC entry point. AWS PrivateDNS is intentionally **disabled** on the endpoint — Zilliz issues its own hostname for your cluster, and we override resolution explicitly with a Route 53 private hosted zone scoped to the VPC.

### Step 1 — Discover the PrivateLink service name

Zilliz Cloud publishes one PrivateLink service per region. Use the `zillizcloud_endpoint_services` data source to look it up by region instead of hardcoding the `vpce-svc-…` string.

```hcl
data "zillizcloud_endpoint_services" "this" {
  region_id = var.zilliz_region_id
}
```

The first element of `endpoint_services[*].endpoint_service` is the AWS service name you'll feed into the VPC endpoint.

### Step 2 — Create the AWS Interface VPC Endpoint

```hcl
resource "aws_vpc_endpoint" "zilliz" {
  vpc_id              = var.vpc_id
  service_name        = data.zillizcloud_endpoint_services.this.endpoint_services[0].endpoint_service
  vpc_endpoint_type   = "Interface"
  subnet_ids          = var.subnet_ids
  security_group_ids  = var.security_group_ids
  private_dns_enabled = false
}
```

`private_dns_enabled = false` is deliberate: Zilliz Cloud issues its own per-cluster hostname, and we will resolve it ourselves in Step 5. Leaving PrivateDNS on would attempt to register service-owned domains in your VPC.

### Step 3 — Create the Zilliz Enterprise cluster

PrivateLink connectivity requires the **Enterprise** plan. After the cluster is created, the resource exposes a `private_link_address` attribute (a full `https://host:port` URL) — that hostname is the one we will resolve to the VPC endpoint in Step 5.

```hcl
resource "zillizcloud_cluster" "enterprise_plan_cluster" {
  cluster_name = "enterprise_plan_cluster"
  region_id    = var.zilliz_region_id
  plan         = "Enterprise"
  cu_size      = 1
  cu_type      = "Performance-optimized"
  project_id   = var.zilliz_project_id
}
```

### Step 4 — Register the endpoint with Zilliz Cloud

The Zilliz PrivateLink service won't accept traffic from a VPC endpoint that hasn't been registered against the project and region. Use `zillizcloud_endpoint` to authorize the AWS VPCE id you just created.

```hcl
resource "zillizcloud_endpoint" "this" {
  project_id  = var.zilliz_project_id
  region_id   = var.zilliz_region_id
  endpoint_id = aws_vpc_endpoint.zilliz.id
}
```

### Step 5 — Wire DNS via a Route 53 private hosted zone

`private_link_address` is a full URL, but Route 53 needs the bare hostname. Strip the scheme and the port:

```hcl
locals {
  private_link_host = element(
    split(":", replace(zillizcloud_cluster.enterprise_plan_cluster.private_link_address, "https://", "")),
    0,
  )
}
```

Create a private hosted zone whose name is that hostname, scoped to the VPC, and add an alias `A` record pointing at the VPC endpoint's first regional DNS entry:

```hcl
resource "aws_route53_zone" "private" {
  name = local.private_link_host

  vpc {
    vpc_id = var.vpc_id
  }

  comment = "Private hosted zone for Zilliz Cloud private link"
}

resource "aws_route53_record" "alias" {
  zone_id = aws_route53_zone.private.zone_id
  name    = local.private_link_host
  type    = "A"

  alias {
    name                   = aws_vpc_endpoint.zilliz.dns_entry[0].dns_name
    zone_id                = aws_vpc_endpoint.zilliz.dns_entry[0].hosted_zone_id
    evaluate_target_health = false
  }
}
```

`evaluate_target_health = false` because Route 53 cannot health-check a private VPCE alias; AWS already manages endpoint availability.

### Applying the configuration

From the directory containing your `.tf` files:

```bash
terraform init
terraform apply
```

Terraform handles ordering automatically: the data source resolves first, then the VPCE and cluster are created in parallel, then the Zilliz endpoint registration and Route 53 records pick up their dependencies.

### Verifying connectivity

From an EC2 instance or other compute resource **inside the VPC** (using the security group attached to the endpoint):

1. Resolve the cluster hostname (the bare host portion of `zillizcloud_cluster.enterprise_plan_cluster.private_link_address`):

   ```bash
   dig +short <your-cluster-private-link-host>
   ```

   You should see one or more private (`10.x.x.x` or `172.16-31.x.x`) addresses corresponding to the endpoint ENIs.

2. Connect with a Milvus client using the cluster's `private_link_address` and the credentials from the `zillizcloud_cluster` resource (`username` / `password`).

If the hostname resolves to a public IP, the Route 53 private hosted zone is not in scope for the VPC you're testing from — verify the zone's `vpc { vpc_id = … }` block references the correct VPC.

### Destroying the resources

```bash
terraform destroy
```

Terraform will deregister the Zilliz endpoint, delete the Route 53 record and zone, then remove the VPC endpoint. The cluster is destroyed alongside everything else.

### Next Steps

- [Upgrading Zilliz Cloud Cluster Compute Unit Size with Terraform](./scale-cluster.md)
- [Import Existing Zilliz Cloud Cluster With Terraform](./import-cluster.md)
- [Acquiring Region IDs for Zilliz Cloud Clusters](./list-regions.md)
````

- [ ] **Step 2: Confirm the file landed**

Run: `wc -l templates/guides/aws-privatelink.md`
Expected: a line count in the 150-220 range. The exact number doesn't matter — this is just a sanity check that the file was written.

- [ ] **Step 3: Commit**

```bash
git add templates/guides/aws-privatelink.md
git commit -m "$(cat <<'EOF'
docs(guides): add AWS PrivateLink end-to-end tutorial template

New template under templates/guides/ that will render to the registry
as docs/guides/aws-privatelink. Walks through service discovery, VPCE
provisioning, cluster creation, endpoint registration, and Route 53
private-zone DNS wiring.
EOF
)"
```

---

## Task 4: Index the new guide in `templates/guides/README.md`

**Files:**
- Modify: `templates/guides/README.md`

- [ ] **Step 1: Insert a new bullet before the closing paragraph**

Use Edit to replace this exact `old_string`:

```
* **Retrieving Cloud Region**: Retrieve region IDs for Zilliz Cloud clusters across various cloud providers using the `zillizcloud_regions` data source. This tutorial demonstrates how to list regions for AWS, GCP, and Azure: [Acquiring Region IDs for Zilliz Cloud Clusters](./list-regions.md)

By leveraging Terraform and the Zilliz Cloud Terraform provider, you can streamline your Zilliz Cloud infrastructure management, promoting efficiency and consistency within your cloud deployments.
```

with this `new_string`:

```
* **Retrieving Cloud Region**: Retrieve region IDs for Zilliz Cloud clusters across various cloud providers using the `zillizcloud_regions` data source. This tutorial demonstrates how to list regions for AWS, GCP, and Azure: [Acquiring Region IDs for Zilliz Cloud Clusters](./list-regions.md)
* **Connecting Privately to Zilliz Cloud**: Provision a private, in-VPC connection from AWS to a Zilliz Cloud Enterprise cluster using AWS PrivateLink, including service discovery, VPC endpoint creation, and Route 53 DNS wiring: [Connecting Privately to Zilliz Cloud with AWS PrivateLink](./aws-privatelink.md)

By leveraging Terraform and the Zilliz Cloud Terraform provider, you can streamline your Zilliz Cloud infrastructure management, promoting efficiency and consistency within your cloud deployments.
```

- [ ] **Step 2: Verify the link target exists**

Run: `test -f templates/guides/aws-privatelink.md && echo OK`
Expected: `OK`.

- [ ] **Step 3: Commit**

```bash
git add templates/guides/README.md
git commit -m "$(cat <<'EOF'
docs(guides): link AWS PrivateLink guide from index

Adds a discoverable bullet under the Guides index so the new tutorial
appears in the registry's Guides listing alongside the existing ones.
EOF
)"
```

---

## Task 5: Regenerate `docs/` and commit the rendered output

**Files:**
- Generated/Modify: `docs/guides/aws-privatelink.md` (and possibly other regenerated files in `docs/`).

- [ ] **Step 1: Run the doc generator**

Run: `make doc`
Expected: exits 0. Watch for warnings about missing files or template errors.

- [ ] **Step 2: Confirm the generated guide exists**

Run: `test -f docs/guides/aws-privatelink.md && echo OK && head -5 docs/guides/aws-privatelink.md`
Expected: `OK` followed by the first lines of the rendered guide. The rendered output should closely match the template (the project's existing guides are 1:1 with their templates).

- [ ] **Step 3: Review the diff**

Run: `git status -- docs/ && git diff --stat docs/`
Expected: `docs/guides/aws-privatelink.md` is new. If other files under `docs/` were also touched (e.g. minor regeneration churn), inspect with `git diff docs/` and decide per file whether to include — the convention in this repo is to commit whatever `make doc` produces.

- [ ] **Step 4: Commit**

```bash
git add docs/
git commit -m "$(cat <<'EOF'
docs: regenerate after AWS PrivateLink guide

Output of make doc for the new templates/guides/aws-privatelink.md.
EOF
)"
```

---

## Task 6: Final verification

**Files:** none.

- [ ] **Step 1: Confirm the branch contains the expected commits**

Run: `git log --oneline master..HEAD`
Expected: five new commits (annotate example, drop variable, add template, link from README, regenerate docs) plus any earlier branch commits that were already present.

- [ ] **Step 2: Sanity-check the rendered guide content**

Run: `grep -E "Step [1-5] —|## Tutorial|### " docs/guides/aws-privatelink.md`
Expected: the section headers from the template appear in the rendered output (Tutorial title, Prerequisites, Architecture, the five Step headings, Applying, Verifying, Destroying, Next Steps).

- [ ] **Step 3: Confirm the example is still self-consistent**

Run: `grep -n zilliz_service_name examples/guides/aws-privatelink/`
Expected: no output (variable fully removed; data source is the sole source).

Run: `grep -n endpoint_services examples/guides/aws-privatelink/main.tf`
Expected: at least the `data "zillizcloud_endpoint_services"` block and one consumer reference.

If all three checks pass, the implementation is complete. The new guide will appear at `registry.terraform.io/providers/zilliztech/zillizcloud/latest/docs/guides/aws-privatelink` after the next provider release.
