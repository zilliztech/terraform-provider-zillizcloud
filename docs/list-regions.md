
## Acquiring Region IDs for Zilliz Cloud Cluster



To provision Zilliz Cloud clusters, you'll must specify the region where the cluster will be deployed. Zilliz Cloud supports multiple regions across various cloud providers, such as AWS, GCP, and Azure. You can retrieve the region IDs for each cloud provider using the zillizcloud_regions data source.

```terraform

data "zillizcloud_regions" "aws_region" {
  cloud_id = "aws"
}

data "zillizcloud_regions" "gcp_region" {
  cloud_id = "gcp"
}

data "zillizcloud_regions" "azure_region" {
  cloud_id = "azure"
}

output "aws_ouput" {
  value = data.zillizcloud_regions.aws_region.items
}


output "gcp_ouput" {
  value = data.zillizcloud_regions.gcp_region.items
}

output "azure_ouput" {
  value = data.zillizcloud_regions.azure_region.items
}
```


```
$ terraform apply --auto-approve

learn-terraform terraform apply -auto-approve
data.zillizcloud_regions.gcp_region: Reading...
data.zillizcloud_regions.aws_region: Reading...
data.zillizcloud_regions.azure_region: Reading...
data.zillizcloud_project.default: Reading...
data.zillizcloud_regions.aws_region: Read complete after 0s
data.zillizcloud_regions.gcp_region: Read complete after 0s
data.zillizcloud_regions.azure_region: Read complete after 0s
data.zillizcloud_project.default: Read complete after 0s [id=proj-4487580fcfe2c8a4391686]

You can apply this plan to save these new output values to the Terraform state, without changing any real infrastructure.

Apply complete! Resources: 0 added, 0 changed, 0 destroyed.

Outputs:

aws_ouput = tolist([
...
  {
    "api_base_url" = "https://api.aws-us-west-2.zillizcloud.com"
    "cloud_id" = "aws"
    "region_id" = "aws-us-west-2"
  },
...
])
azure_ouput = tolist([
...
])
gcp_ouput = tolist([
...
])
```

Upon execution, this Terraform script retrieves region details for each cloud provider, facilitating subsequent cluster provisioning steps. In this session, we would use **aws-us-east-2** in the following example.
