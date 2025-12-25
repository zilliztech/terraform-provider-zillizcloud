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
      storage = {
        client_id   = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
        resource_id = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/storage-identity"
      }
      kubelet = {
        client_id   = "b2c3d4e5-f6a7-8901-bcde-f12345678901"
        resource_id = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/kubelet-identity"
      }
      maintenance = {
        client_id   = "c3d4e5f6-a7b8-9012-cdef-123456789012"
        resource_id = "/subscriptions/sub-xxxxx/resourceGroups/rg-xxxxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/maintenance-identity"
      }
    }

    storage = {
      storage_account_name = "zilliz-storage-0af21b"
      container_name       = "zilliz-container-i9fip2"
    }
  }

  depends_on = [zillizcloud_byoc_i_project_settings.this, zillizcloud_byoc_i_project_agent.this]
}