## Getting Started with Zilliz Cloud Terraform Provider

This guide walks you through the process of installing and configuring the Zilliz Cloud provider.

### 1.Prerequisites

Before you begin, ensure you have the following:

1. **Terraform Installed**: Download and install Terraform from [here](https://www.terraform.io/downloads.html) by following the provided instructions.

2. **Zilliz Cloud Account**: Access to Zilliz Cloud and your API Key are essential. Refer to the [documentation](https://docs.zilliz.com/docs/manage-api-keys) to obtain your API key.

## 2. Download Zilliz Cloud Terraform Provider

Start by configuring the Zilliz Cloud provider within your Terraform configuration file (`main.tf`). Follow these steps:

```hcl
terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}
```

## 3. Initialize Terraform Configuration

Initialize the Terraform configuration by running:

```bash
terraform init
```

Terraform will download the `zillizcloud` provider and install it in a hidden subdirectory of your current working directory, named `.terraform`.

### 4. Authenticate Zilliz Cloud Terraform Provider

Your Zilliz Cloud API Key is required to use the Terraform Provider. There are two ways to configure this.

#### Option 1: Specify API Key in Provider Block

Append the following code to your `main.tf` file:

```hcl
provider "zillizcloud" {
  api_key = "<your-api-key>"
}
```

Replace `<your-api-key>` with your Zilliz Cloud API Key.

#### Option 2: Use Environment Variable

Set the API key as an environment variable:

```bash
export ZILLIZCLOUD_API_KEY="<your-api-key>"
```

Then the provider declaration in your `main.tf` file is simply:

```hcl
provider "zillizcloud" {
}
```

By following these steps, you should have the Zilliz Cloud Terraform provider configured and ready to move on to the next steps.

## Next Steps
- Explore creating a **Free Plan** Cluster: [Creating a Free Plan Cluster](./create-a-free-cluster.md)
- Explore creating a **Serverless Plan** Cluster: [Creating a Serverless Plan Cluster](./create-a-serverless-cluster.md)
- Explore creating a **Standard Plan** Cluster: [Creating a Standard Plan Cluster](./create-a-standard-cluster.md)
