package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestOnDemandClusterRegistered(t *testing.T) {
	ctx := context.Background()
	provider := &ZillizProvider{}

	for _, factory := range provider.Resources(ctx) {
		res := factory()
		var resp resource.MetadataResponse
		res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "zillizcloud"}, &resp)
		if resp.TypeName == "zillizcloud_on_demand_cluster" {
			return
		}
	}

	t.Fatal("zillizcloud_on_demand_cluster is not registered")
}
