## Terraform Integration Overview with Zilliz Cloud

Terraform, an open-source infrastructure as code (IaC) tool by HashiCorp, empowers you to automate the provisioning and management of your cloud resources. 

The Terraform provider for Zilliz Cloud acts as a bridge between Terraform and the Zilliz Cloud platform. This plugin seamlessly integrates Zilliz Cloud resources into your Terraform workflow, allowing you to manage them alongside your existing infrastructure using Terraform configurations.


### Getting Started with Terraform and Zilliz Cloud

This guide provides a comprehensive overview of using Terraform with Zilliz Cloud. Here are the resources to get you started:

* **Installation**: Learn how to install the Zilliz Cloud Terraform provider and configure it within your Terraform project. Refer to the tutorial: [Getting Started with Zilliz Cloud Terraform Provider](./get-start.md)
* **Creating Cluster:** Discover how to define, provision, and manage Zilliz Cloud clusters using Terraform configurations. 
    * Create Free Plan Clusters for learning, experimenting, and prototype: [Creating a Free Plan Cluster](./create-a-free-cluster.md)
    * Create Serverless Plan Clusters designed for serverless applications with variable or infrequent traffic: [Creating a Serverless Plan Cluster](./create-a-serverless-cluster.md)
    * Create Standard Plan Clusters with more resources for production workloads: [Creating a Standard Plan Cluster](./create-a-standard-cluster.md)
* **Scaling Clusters**: Learn how to leverage Terraform to upgrade the compute unit size of your Zilliz Cloud clusters to meet changing workload demands: [Upgrading Zilliz Cloud Cluster Compute Unit Size with Terraform](./scale-cluster.md)
* **Importing Existing Clusters**: Utilize Terraform to import existing Zilliz Cloud clusters into your Terraform state, enabling them to be managed alongside other infrastructure using Terraform configurations: [Import Existing Zilliz Cloud Cluster With Terraform](./import-cluster.md)
* **Retrieving Cloud Region**: Retrieve region IDs for Zilliz Cloud clusters across various cloud providers using the `zillizcloud_regions` data source. This tutorial demonstrates how to list regions for AWS, GCP, and Azure: [Acquiring Region IDs for Zilliz Cloud Clusters](./list-regions.md)

By leveraging Terraform and the Zilliz Cloud Terraform provider, you can streamline your Zilliz Cloud infrastructure management, promoting efficiency and consistency within your cloud deployments.

