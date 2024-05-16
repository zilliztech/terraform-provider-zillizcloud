
## Retrieving Project IDs in Your Zilliz Cloud Account

Every Zilliz Cloud cluster is associated with a project. To create a cluster, it's imperative to obtain the corresponding project ID.

To access information regarding available projects, utilize the zillizcloud_projects data source as demonstrated below:

```hcl
data "zillizcloud_project" "default" {}

output "projects" {
  value = data.zillizcloud_project.default
}
```

```shell
$ terraform apply --auto-approve

data.zillizcloud_project.default: Reading...
data.zillizcloud_project.default: Read complete after 1s [id=proj-4487580fcfe2c8a4391686]

Apply complete! Resources: 0 added, 0 changed, 0 destroyed.

Outputs:

projects = {
  "created_at" = 1714892175000
  "id" = "proj-4487580fcfe2cxxxxx"
  "instance_count" = 0
  "name" = "Default Project"
}
```

The project ID **"proj-4487580fcfe2cxxxxx"** is displayed in the output section. You can use your ID to create Zilliz Cloud clusters within the specified project, or refer to the project ID via `data.zillizcloud_project.default.id` in your Terraform configuration file.
