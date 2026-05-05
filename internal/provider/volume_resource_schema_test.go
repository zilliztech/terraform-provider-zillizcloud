package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	volumeRequiresReplaceDescription = "If the value of this attribute changes, Terraform will destroy and recreate the resource."
	volumeUseStateDescription        = "Once set, the value of this attribute in state will not change."
)

func TestVolumeResourceMetadata(t *testing.T) {
	res := NewVolumeResource()
	var resp fwresource.MetadataResponse

	res.Metadata(context.Background(), fwresource.MetadataRequest{ProviderTypeName: "zillizcloud"}, &resp)

	if resp.TypeName != "zillizcloud_volume" {
		t.Fatalf("TypeName=%q, want %q", resp.TypeName, "zillizcloud_volume")
	}
}

func TestProviderRegistersVolumeResource(t *testing.T) {
	ctx := context.Background()
	provider := &ZillizProvider{}

	for _, factory := range provider.Resources(ctx) {
		res := factory()
		var resp fwresource.MetadataResponse
		res.Metadata(ctx, fwresource.MetadataRequest{ProviderTypeName: "zillizcloud"}, &resp)
		if resp.TypeName == "zillizcloud_volume" {
			return
		}
	}

	t.Fatal("provider Resources() does not register zillizcloud_volume")
}

func TestVolumeResourceSchemaAttributes(t *testing.T) {
	ctx := context.Background()
	res := NewVolumeResource().(*VolumeResource)
	schema := testVolumeResourceSchema(t, res)

	tests := []struct {
		name             string
		required         bool
		optional         bool
		computed         bool
		requiresReplace  bool
		useStateForValue bool
	}{
		{name: "id", computed: true, useStateForValue: true},
		{name: "project_id", required: true, requiresReplace: true},
		{name: "region_id", required: true, requiresReplace: true},
		{name: "volume_name", required: true, requiresReplace: true},
		{name: "type", optional: true, computed: true, requiresReplace: true},
		{name: "storage_integration_id", optional: true, requiresReplace: true},
		{name: "path", optional: true, requiresReplace: true},
		{name: "status", computed: true},
		{name: "create_time", computed: true},
	}

	if len(schema.Attributes) != len(tests) {
		t.Fatalf("schema has %d attributes, want %d", len(schema.Attributes), len(tests))
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := requireVolumeStringAttribute(t, schema, tt.name)
			if attr.Required != tt.required {
				t.Fatalf("%s Required=%t, want %t", tt.name, attr.Required, tt.required)
			}
			if attr.Optional != tt.optional {
				t.Fatalf("%s Optional=%t, want %t", tt.name, attr.Optional, tt.optional)
			}
			if attr.Computed != tt.computed {
				t.Fatalf("%s Computed=%t, want %t", tt.name, attr.Computed, tt.computed)
			}
			if tt.requiresReplace {
				requireVolumePlanModifier(t, attr, volumeRequiresReplaceDescription)
			}
			if tt.useStateForValue {
				requireVolumePlanModifier(t, attr, volumeUseStateDescription)
			}
			if !tt.requiresReplace && !tt.useStateForValue && len(attr.PlanModifiers) != 0 {
				t.Fatalf("%s has unexpected plan modifiers: %d", tt.name, len(attr.PlanModifiers))
			}
		})
	}

	typeAttr := requireVolumeStringAttribute(t, schema, "type")
	requireVolumeStringDefault(t, ctx, typeAttr, "MANAGED")
	requireVolumeValidatorResult(t, ctx, typeAttr, "MANAGED", false)
	requireVolumeValidatorResult(t, ctx, typeAttr, "EXTERNAL", false)
	requireVolumeValidatorResult(t, ctx, typeAttr, "LOCAL", true)

	pathAttr := requireVolumeStringAttribute(t, schema, "path")
	requireVolumeValidatorResult(t, ctx, pathAttr, "", false)
	requireVolumeValidatorResult(t, ctx, pathAttr, "s3://bucket/prefix/", false)
	requireVolumeValidatorResult(t, ctx, pathAttr, "s3://bucket/prefix", true)
}

func requireVolumeStringAttribute(t *testing.T, s schema.Schema, name string) schema.StringAttribute {
	t.Helper()
	attr, ok := s.Attributes[name]
	if !ok {
		t.Fatalf("missing schema attribute %q", name)
	}
	stringAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatalf("schema attribute %q has type %T, want schema.StringAttribute", name, attr)
	}
	return stringAttr
}

func requireVolumePlanModifier(t *testing.T, attr schema.StringAttribute, wantDescription string) {
	t.Helper()
	for _, modifier := range attr.PlanModifiers {
		if modifier.MarkdownDescription(context.Background()) == wantDescription {
			return
		}
	}
	t.Fatalf("attribute missing plan modifier with description %q", wantDescription)
}

func requireVolumeStringDefault(t *testing.T, ctx context.Context, attr schema.StringAttribute, want string) {
	t.Helper()
	if attr.Default == nil {
		t.Fatal("attribute missing string default")
	}
	var resp defaults.StringResponse
	attr.Default.DefaultString(ctx, defaults.StringRequest{Path: path.Root("type")}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("default diagnostics: %s", resp.Diagnostics.Errors()[0].Summary())
	}
	if !resp.PlanValue.Equal(types.StringValue(want)) {
		t.Fatalf("default=%q, want %q", resp.PlanValue.ValueString(), want)
	}
}

func requireVolumeValidatorResult(t *testing.T, ctx context.Context, attr schema.StringAttribute, value string, wantError bool) {
	t.Helper()
	var resp validator.StringResponse
	for _, v := range attr.Validators {
		v.ValidateString(ctx, validator.StringRequest{
			Path:        path.Root("test"),
			ConfigValue: types.StringValue(value),
		}, &resp)
	}
	if resp.Diagnostics.HasError() != wantError {
		t.Fatalf("validator error for %q=%t, want %t", value, resp.Diagnostics.HasError(), wantError)
	}
}
