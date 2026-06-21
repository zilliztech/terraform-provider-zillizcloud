package byoc_op_test

import (
	"context"
	"strings"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	byoc_op "github.com/zilliztech/terraform-provider-zillizcloud/internal/provider/byoc_i"
)

func TestByocOpProjectResourceExactlyOneOfIncludesGCP(t *testing.T) {
	ctx := context.Background()
	resourceWithValidators, ok := byoc_op.NewBYOCOpProjectResource().(frameworkresource.ResourceWithConfigValidators)
	if !ok {
		t.Fatal("BYOC project resource must implement config validators")
	}

	validators := resourceWithValidators.ConfigValidators(ctx)
	if len(validators) != 1 {
		t.Fatalf("ConfigValidators length = %d, want 1", len(validators))
	}

	description := validators[0].Description(ctx)
	for _, blockName := range []string{"aws", "azure", "gcp"} {
		if !strings.Contains(description, blockName) {
			t.Fatalf("ExactlyOneOf validator description %q does not include %q", description, blockName)
		}
	}
}
