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
