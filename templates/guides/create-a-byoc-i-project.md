# Zilliz Cloud BYOC-I Project Setup Guide (Bootstrap Provider)

This guide explains how to configure a Bring-Your-Own-Cloud-Infrastructure (BYOC-I) environment using the `zillizcloud_byoc_project` Terraform resource. You'll deploy the data plane of your BYOC-I project and a BYOC Agent within your AWS infrastructure while maintaining complete control over your data and network environment.

## Table of Contents
- [Key Concepts](#key-concepts)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Architecture Overview](#architecture-overview)

## Key Concepts <a name="key-concepts"></a>

### What is BYOC-I?

Zilliz Cloud BYOC-I enables you to:

- Maintain data sovereignty with storage in your AWS account
- Deploy Milvus clusters in your private VPC
- Leverage Zilliz's managed service while keeping infrastructure management entirely in your hands
- Comply with enterprise security requirements

## Prerequisites <a name="prerequisites"></a>

- **AWS Requirements**
  - AWS account with admin privileges
  - AWS CLI configured with credentials
  - Target region enabled in your AWS account

- **Zilliz Cloud Requirements**
  - Active Zilliz Cloud account
  - API credentials with project creation permissions

- **Tooling**
  - Terraform v1.0+

## Quick Start <a name="quick-start"></a>

### 1. Clone Reference Architecture

```bash
git clone https://github.com/zilliztech/terraform-zilliz-examples.git
cd terraform-zilliz-examples/examples/aws-project-byoc-I
```

### 2. Initialize the Terraform project

```bash
terraform init
```

### 3. Start the deployment

```bash
export ZILLIZ_API_KEY=<your_api_key>
terraform apply \
    -var="dataplane_id=<your_dataplane_id>" \
    -var="project_id=<your_project_id>" \
    -var="vpc_cidr=10.0.0.0/16"
```

You can find your `your_api_key`, `dataplane_id` and `project_id` in the `Deploy Data Plane` dialog box in the Zilliz Cloud console.

![Deploy Data Plane Dialog Box](https://assets.zilliz.com/docs/deploy-data-plane.png)

## Architecture Overview <a name="architecture-overview"></a>

![BYOC Architecture Diagram](https://assets.zilliz.com/docs/byoc-i-architecture.png)

In the BYOC-I mode, instead of asking for cross-account permissions to manage infrastructure resources on your behalf, Zilliz leaves infrastructure management entirely in your hands, thereby enhancing data control sovereignty.

However, you may choose to grant the agent the necessary permissions so that Zilliz can assist you with infrastructure management if necessary.
