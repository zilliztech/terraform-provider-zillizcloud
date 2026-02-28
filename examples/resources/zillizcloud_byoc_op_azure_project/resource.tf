terraform {
  required_providers {
    zillizcloud = {
      source = "zilliztech/zillizcloud"
    }
  }
}

provider "zillizcloud" {
}




resource "zillizcloud_byoc_i_project_settings" "this" {

  cloud_provider = "azure"
  region         = "eastus"
  project_name   = "byoc-zilliz-azure-test"

  instances = {
    core = {
      vm    = "Standard_D8s_v5"
      count = 3
    }

    fundamental = {
      vm        = "Standard_D8s_v5"
      min_count = 1
      max_count = 1
    }

    search = {
      vm        = "Standard_D16ds_v5"
      min_count = 1
      max_count = 1
    }

    index = {
      vm        = "Standard_D8s_v5"
      min_count = 2
      max_count = 2
    }

    auto_scaling = true
    arch         = "X86"
  }

}


resource "zillizcloud_byoc_i_project_agent" "this" {
  project_id    = zillizcloud_byoc_i_project_settings.this.project_id
  data_plane_id = zillizcloud_byoc_i_project_settings.this.data_plane_id
}

resource "zillizcloud_byoc_i_project" "this" {

  lifecycle {
    ignore_changes = [data_plane_id, project_id, azure, ext_config]
  }

  data_plane_id = zillizcloud_byoc_i_project_settings.this.data_plane_id
  project_id    = zillizcloud_byoc_i_project_settings.this.project_id

  azure = {

    region = "eastus"

    network = {
      vnet_id = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.Network/virtualNetworks/my-vnet"

      subnet_ids = [
        "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/subnet-1",
        "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/subnet-2",
        "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/subnet-3"
      ]
      nsg_ids = [
        "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.Network/networkSecurityGroups/my-nsg"
      ]
      private_endpoint_id = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.Network/privateEndpoints/my-private-endpoint"
    }

    identity = {
      storages = [
        {
          client_id    = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
          principal_id = "e4a2b9d1-7c5a-4f3e-8d6b-1a2c3d4e5f6a"
          resource_id  = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/storage-identity-1"
        },
        {
          client_id    = "5f4e3d2c-1b0a-4987-a654-3210fedcba98"
          principal_id = "8d7c6b5a-4e3f-4210-9a8b-7c6d5e4f3a2b"
          resource_id  = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/storage-identity-2"
        },
        {
          client_id    = "1a2b3c4d-5e6f-7890-abcd-ef0123456789"
          principal_id = "f1e2d3c4-b5a6-7890-1234-567890abcdef"
          resource_id  = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/storage-identity-3"
        },
        {
          client_id    = "2b3c4d5e-6f7a-8901-bcde-f01234567890"
          principal_id = "a2b3c4d5-e6f7-8901-2345-678901abcdef"
          resource_id  = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/storage-identity-4"
        },
        {
          client_id    = "3c4d5e6f-7a8b-9012-cdef-012345678901"
          principal_id = "b3c4d5e6-f7a8-9012-3456-789012abcdef"
          resource_id  = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/storage-identity-5"
        },
        {
          client_id    = "4d5e6f7a-8b9c-0123-def0-123456789012"
          principal_id = "c4d5e6f7-a8b9-0123-4567-890123abcdef"
          resource_id  = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/storage-identity-6"
        },
        {
          client_id    = "5e6f7a8b-9c0d-1234-ef01-234567890123"
          principal_id = "d5e6f7a8-b9c0-1234-5678-901234abcdef"
          resource_id  = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/storage-identity-7"
        },
        {
          client_id    = "6f7a8b9c-0d1e-2345-f012-345678901234"
          principal_id = "e6f7a8b9-c0d1-2345-6789-012345abcdef"
          resource_id  = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/storage-identity-8"
        },
        {
          client_id    = "7a8b9c0d-1e2f-3456-0123-456789012345"
          principal_id = "f7a8b9c0-d1e2-3456-7890-123456abcdef"
          resource_id  = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/storage-identity-9"
        },
        {
          client_id    = "8b9c0d1e-2f3a-4567-1234-567890123456"
          principal_id = "a8b9c0d1-e2f3-4567-8901-234567abcdef"
          resource_id  = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/storage-identity-10"
        }
      ]
      kubelet = {
        client_id    = "b2c3d4e5-f6a7-8901-bcde-f12345678901"
        principal_id = "9e8d7c6b-5a4f-3e2d-1c0b-8a796f5e4d3c"
        resource_id  = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/kubelet-identity"
      }
      maintenance = {
        client_id    = "c3d4e5f6-a7b8-9012-cdef-123456789012"
        principal_id = "4d3c2b1a-0987-6543-2109-876543210987"
        resource_id  = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/maintenance-identity"
      }
    }

    storage = {
      storage_account_name = "zilliz-storage-0af21b"
      container_name       = "zilliz-container-i9fip2"
    }
  }

  depends_on = [zillizcloud_byoc_i_project_settings.this, zillizcloud_byoc_i_project_agent.this]
}