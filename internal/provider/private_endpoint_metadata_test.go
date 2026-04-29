package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestPrivateEndpointTerraformTypeNames(t *testing.T) {
	ctx := context.Background()

	dataSourceCases := []struct {
		name    string
		factory func() datasource.DataSource
		want    string
	}{
		{
			name:    "endpoint services",
			factory: NewEndpointServicesDataSource,
			want:    "zillizcloud_private_endpoint_services",
		},
		{
			name:    "endpoints",
			factory: NewEndpointsDataSource,
			want:    "zillizcloud_private_endpoints",
		},
	}

	for _, tc := range dataSourceCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := tc.factory()
			var resp datasource.MetadataResponse
			ds.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "zillizcloud"}, &resp)
			if resp.TypeName != tc.want {
				t.Fatalf("TypeName=%q, want %q", resp.TypeName, tc.want)
			}
		})
	}

	resourceCases := []struct {
		name    string
		factory func() resource.Resource
		want    string
	}{
		{
			name:    "endpoint",
			factory: NewEndpointResource,
			want:    "zillizcloud_private_endpoint",
		},
		{
			name:    "endpoint whitelist",
			factory: NewEndpointWhitelistResource,
			want:    "zillizcloud_private_endpoint_whitelist",
		},
	}

	for _, tc := range resourceCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.factory()
			var resp resource.MetadataResponse
			res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "zillizcloud"}, &resp)
			if resp.TypeName != tc.want {
				t.Fatalf("TypeName=%q, want %q", resp.TypeName, tc.want)
			}
		})
	}
}
