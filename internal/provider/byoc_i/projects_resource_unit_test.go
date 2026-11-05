package byoc_op_test

import (
	"context"
	"strings"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func TestByocAzureCmekSchema(t *testing.T) {
	ctx := context.Background()
	var response frameworkresource.SchemaResponse
	byoc_op.NewBYOCOpProjectResource().Schema(ctx, frameworkresource.SchemaRequest{}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	azure, ok := response.Schema.Attributes["azure"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("Azure configuration must be a nested attribute")
	}
	cse, ok := azure.Attributes["cse"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("Azure CSE must be a nested attribute")
	}
	if !cse.Optional || cse.Required {
		t.Fatal("Azure CSE must remain optional")
	}
	identity, ok := cse.Attributes["identity"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("CSE identity must be a nested attribute")
	}
	if !identity.Required {
		t.Fatal("configured CSE requires a dedicated identity")
	}
	for _, name := range []string{"client_id", "principal_id", "resource_id"} {
		attribute, ok := identity.Attributes[name].(schema.StringAttribute)
		if !ok {
			t.Fatalf("CSE identity %s must be a string attribute", name)
		}
		if !attribute.Required || len(attribute.PlanModifiers) == 0 {
			t.Fatalf("CSE %s must be required and replacement-only", name)
		}
	}
	for _, name := range []string{"tenant_id", "default_key_identifier"} {
		attribute, ok := cse.Attributes[name].(schema.StringAttribute)
		if !ok {
			t.Fatalf("CSE %s must be a string attribute", name)
		}
		if !attribute.Required || len(attribute.PlanModifiers) == 0 {
			t.Fatalf("CSE %s must be required and replacement-only", name)
		}
	}
	workloadIdentity, ok := azure.Attributes["identity"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("Azure workload identity must remain a nested attribute")
	}
	storages, ok := workloadIdentity.Attributes["storages"].(schema.SetNestedAttribute)
	if !ok {
		t.Fatal("Azure storage identities must remain a set nested attribute")
	}
	if !storages.Required || len(storages.Validators) != 1 {
		t.Fatal("existing storage identity requirements changed")
	}
	key, ok := cse.Attributes["default_key_identifier"].(schema.StringAttribute)
	if !ok {
		t.Fatal("CSE key identifier must be a string attribute")
	}
	for _, tc := range []struct {
		key   string
		valid bool
	}{
		{"https://customer-kms.vault.azure.net/keys/cmek-key", true},
		{"", false},
		{"https://customer-kms.vault.azure.net/keys/cmek-key/version", false},
		{"https://customer-kms.managedhsm.azure.net/keys/cmek-key", false},
		{"http://customer-kms.vault.azure.net/keys/cmek-key", false},
		{"projects/p/locations/us-west1/keyRings/r/cryptoKeys/k", false},
		{"arn:aws:kms:us-west-2:123456789012:key/key", false},
	} {
		t.Run(tc.key, func(t *testing.T) {
			var result validator.StringResponse
			for _, v := range key.Validators {
				v.ValidateString(ctx, validator.StringRequest{ConfigValue: types.StringValue(tc.key)}, &result)
			}
			if result.Diagnostics.HasError() == tc.valid {
				t.Fatalf("key valid=%t, diagnostics=%v", tc.valid, result.Diagnostics)
			}
		})
	}
}
