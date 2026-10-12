package byoc_op_test

import (
	"context"
	"strings"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

func TestByocGcpCmekSchemaRequiresDedicatedIdentity(t *testing.T) {
	var response frameworkresource.SchemaResponse
	byoc_op.NewBYOCOpProjectResource().Schema(context.Background(), frameworkresource.SchemaRequest{}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	gcp, ok := response.Schema.Attributes["gcp"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("GCP configuration must be a nested attribute")
	}
	cse, ok := gcp.Attributes["cse"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("GCP CMEK configuration must be a nested attribute")
	}
	if !cse.Optional {
		t.Fatal("CMEK must remain optional")
	}
	email, ok := cse.Attributes["service_account_email"].(schema.StringAttribute)
	if !ok || !email.Required {
		t.Fatal("dedicated CMEK service account must be required when CSE is enabled")
	}
	key, ok := cse.Attributes["default_key_name"].(schema.StringAttribute)
	if !ok || !key.Required {
		t.Fatal("default key must be required when CSE is enabled")
	}
	identity, ok := gcp.Attributes["identity"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("GCP identity must remain a nested attribute")
	}
	if _, ok := identity.Attributes["storage_sa"]; !ok {
		t.Fatal("existing bucket storage identity must be preserved")
	}
	aws, ok := response.Schema.Attributes["aws"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("AWS configuration must remain a nested attribute")
	}
	awsCSE, ok := aws.Attributes["cse"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("AWS CSE must remain a nested attribute")
	}
	if _, ok := awsCSE.Attributes["aws_cse_role_arn"]; !ok {
		t.Fatal("existing AWS CSE schema must remain unchanged")
	}
}
