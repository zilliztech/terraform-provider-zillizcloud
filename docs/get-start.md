## Getting Started with Zilliz Cloud Terraform Provider

This guide walks you through the process of installing and configuring the Zilliz Cloud provider.

### 1.Prerequisites

Before you begin, ensure you have the following:

1. **Terraform Installed**: Download and install Terraform from [here](https://www.terraform.io/downloads.html) by following the provided instructions.

2. **Zilliz Cloud Account**: Access to Zilliz Cloud and your API Key are essential. Refer to the [documentation](https://docs.zilliz.com/docs/manage-api-keys) to obtain your API key.

## 2. Configure Terraform Provider

Start by configuring the Zilliz Cloud provider within your Terraform configuration file (`main.tf`). Follow these steps:

```hcl
terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
  api_key = "<your-api-key>"
}
```

Replace `<your-api-key>` with your Zilliz Cloud API Key.

Alternatively, you can use the environment variable `ZILLIZCLOUD_API_KEY` instead of specifying it in the provider block.

```bash
$ export ZILLIZCLOUD_API_KEY="<your-api-key>"
```

## 3. Initialize Terraform Configuration

Initialize the Terraform configuration by running:

```bash
terraform init
```

Terraform will download the `zillizcloud` provider and install it in a hidden subdirectory of your current working directory, named `.terraform`.
