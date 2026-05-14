## Acquiring Region IDs for Zilliz Cloud Clusters

To provision Zilliz Cloud clusters, you need to specify the region where the cluster will be deployed. Zilliz Cloud supports multiple regions across various cloud providers, such as AWS, GCP, and Azure.

Refer to the Zilliz Cloud documentation, [Cloud Providers & Regions](https://docs.zilliz.com/docs/cloud-providers-and-regions), for a full list of available cloud providers and regions.

This tutorial demonstrates how to retrieve region IDs for each cloud provider using the `zillizcloud_regions` data source. While this information is also available in the documentation, the data source provides a Terraform-friendly way to integrate it into your infrastructure code.

### 1.Prerequisites

Before you begin, ensure you have the following:

1. **Terraform Installed**: Download and install Terraform from [here](https://www.terraform.io/downloads.html) by following the provided instructions.

2. **Zilliz Cloud Account**: Access to Zilliz Cloud and your API Key are essential. Refer to the [documentation](https://docs.zilliz.com/docs/manage-api-keys) to obtain your API key.



## Listing Zilliz Cloud Regions

This code snippet demonstrates how to retrieve region information for specific cloud providers:

```terraform
data "zillizcloud_regions" "all_regions" {
}

data "zillizcloud_regions" "aws_region" {
  cloud_id = "aws"
}

data "zillizcloud_regions" "gcp_region" {
  cloud_id = "gcp"
}

data "zillizcloud_regions" "azure_region" {
  cloud_id = "azure"
}

output "aws_output" {
  value = data.zillizcloud_regions.aws_region.items
}

output "gcp_output" {
  value = data.zillizcloud_regions.gcp_region.items
}

output "azure_output" {
  value = data.zillizcloud_regions.azure_region.items
}

output "all_output" {
  value = data.zillizcloud_regions.all_regions.items
}
```

**Explanation:**

* We define filtered `zillizcloud_regions` data sources for AWS, GCP, and Azure, plus an unfiltered `all_regions` data source.
* A data source retrieves regions for a specific cloud provider when `cloud_id` is set. Without `cloud_id`, it retrieves all enabled regions.
* The `output` blocks reference the `.items` attribute of each data source, which contains a list of region objects.


### Executing the Terraform Script

Run the following command to execute the Terraform script and retrieve region information:

```
terraform apply --auto-approve
```

This command applies the Terraform configuration and displays the retrieved region details. The output will be similar to this:

```
...
data.zillizcloud_regions.gcp_region: Reading...
data.zillizcloud_regions.aws_region: Reading...
data.zillizcloud_regions.azure_region: Reading...

Outputs:

aws_output = tolist([
  ...
  {
    "api_base_url" = "https://api.cloud.zilliz.com/v2"
    "cloud_id" = "aws"
    "domain" = "api.cloud.zilliz.com"
    "region_id" = "aws-us-west-2"
    "supported_cluster_types" = tolist([
      "free",
      "serverless",
      "dedicated",
    ])
  },
  ...
])
azure_output = tolist([
  ...
])
gcp_output = tolist([
  ...
])
```

The script retrieves region details for each cloud provider specified in the `cloud_id` arguments. In this example, the output showcases the `aws-us-west-2` region for the AWS cloud provider. The `api_base_url` field remains available for compatibility, but it is deprecated; use `domain` for new configurations.

This retrieved region ID (e.g., `aws-us-west-2`) can be used to create Zilliz Cloud clusters within the specified region in subsequent Terraform configurations.
