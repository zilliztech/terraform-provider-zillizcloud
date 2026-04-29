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
data "zillizcloud_private_endpoint_services" "this" {
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
  service_name        = data.zillizcloud_private_endpoint_services.this.endpoint_services[0].endpoint_service
  vpc_endpoint_type   = "Interface"
  subnet_ids          = var.subnet_ids
  security_group_ids  = var.security_group_ids
  private_dns_enabled = false
}

# Authorize this specific VPC endpoint on the Zilliz side. The Zilliz
# PrivateLink service won't accept traffic from a VPCE that hasn't been
# registered against the project + region.
resource "zillizcloud_private_endpoint" "this" {
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
